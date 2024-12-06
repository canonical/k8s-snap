package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/control"
	mctypes "github.com/canonical/microcluster/v2/rest/types"
)

func startControlPlaneServices(ctx context.Context, snap snap.Snap, datastore string) error {
	// Start services
	switch datastore {
	case "k8s-dqlite":
		if err := snaputil.StartK8sDqliteServices(ctx, snap); err != nil {
			return fmt.Errorf("failed to start control plane services: %w", err)
		}
	case "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", datastore, setup.SupportedDatastores)
	}

	if err := snaputil.StartControlPlaneServices(ctx, snap); err != nil {
		return fmt.Errorf("failed to start control plane services: %w", err)
	}
	return nil
}

func stopControlPlaneServices(ctx context.Context, snap snap.Snap, datastore string) error {
	// Stop services
	switch datastore {
	case "k8s-dqlite":
		if err := snaputil.StopK8sDqliteServices(ctx, snap); err != nil {
			return fmt.Errorf("failed to stop k8s-dqlite service: %w", err)
		}
	case "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", datastore, setup.SupportedDatastores)
	}

	if err := snaputil.StopControlPlaneServices(ctx, snap); err != nil {
		return fmt.Errorf("failed to stop control plane services: %w", err)
	}
	return nil
}

func waitApiServerReady(ctx context.Context, snap snap.Snap) error {
	// Wait for API server to come up
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	if err := client.WaitKubernetesEndpointAvailable(ctx); err != nil {
		return fmt.Errorf("kubernetes endpoints not ready yet: %w", err)
	}

	return nil
}

func waitControlPlaneServices(ctx context.Context, snap snap.Snap) error {
	// The services may be able to start, appearing to be "active", but they might eventually fail due to
	// various reasons, and they may be restarted. We're checking their activity a few times.
	if err := control.Consistently(ctx, 3, 5*time.Second, func() error {
		if err := snaputil.CheckControlPlaneServicesStates(ctx, snap, "active"); err != nil {
			return fmt.Errorf("failed to ensure all control plane services are active: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed after retry: %w", err)
	}
	return nil
}

func DetermineLocalhostAddress(clusterMembers []mctypes.ClusterMember) (string, error) {
	// Check if any of the cluster members have an IPv6 address, if so return "::1"
	// if one member has an IPv6 address, other members should also have IPv6 interfaces
	for _, clusterMember := range clusterMembers {
		memberAddress := clusterMember.Address.Addr().String()
		nodeIP := net.ParseIP(memberAddress)
		if nodeIP == nil {
			return "", fmt.Errorf("failed to parse node IP address %q", memberAddress)
		}

		if nodeIP.To4() == nil {
			return "[::1]", nil
		}
	}

	// If no IPv6 addresses are found this means the cluster is IPv4 only
	return "127.0.0.1", nil
}
