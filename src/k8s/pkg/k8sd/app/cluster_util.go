package app

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/canonical/k8s/pkg/client/dqlite"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
)

func startControlPlaneServices(ctx context.Context, snap snap.Snap, datastore string, nodeAdress string) error {
	// Start services
	switch datastore {
	case "k8s-dqlite":
		if err := snaputil.StartK8sDqliteServices(ctx, snap); err != nil {
			return fmt.Errorf("failed to start k8s-dqlite services: %w", err)
		}

		if err := waitK8sDqliteReady(ctx, snap, nodeAdress); err != nil {
			return fmt.Errorf("failed to ensure that the node joined the cluster: %w", err)
		}

	case "etcd":
		if err := snaputil.StartEtcdServices(ctx, snap); err != nil {
			return fmt.Errorf("failed to start etcd services: %w", err)
		}
	case "external":
		// For external datastore, we do not start any services here.
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", datastore, setup.SupportedDatastores)
	}

	if err := snaputil.StartControlPlaneServices(ctx, snap); err != nil {
		return fmt.Errorf("failed to start control plane services: %w", err)
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

// waitK8sDqliteReady waits until the joining node is reflected as a cluster member by the k8s-qlite leader.
func waitK8sDqliteReady(ctx context.Context, snap snap.Snap, nodeAddress string) error {
	log := log.FromContext(ctx)
	log.Info("waiting for k8s-dqlite to be ready", "nodeAddress", nodeAddress)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			// Create a dqlite client to check cluster membership
			clusterYamlPath := path.Join(snap.K8sDqliteStateDir(), "cluster.yaml")
			clusterCertPath := path.Join(snap.K8sDqliteStateDir(), "cluster.crt")
			clusterKeyPath := path.Join(snap.K8sDqliteStateDir(), "cluster.key")

			client, err := dqlite.NewClient(ctx, dqlite.ClientOpts{
				ClusterYAML: clusterYamlPath,
				ClusterCert: clusterCertPath,
				ClusterKey:  clusterKeyPath,
			})
			if err != nil {
				log.Info("failed to create dqlite client, retrying", "error", err)
				continue
			}

			// Get cluster members from the dqlite leader
			members, err := client.ListMembers(ctx)
			if err != nil {
				log.Info("failed to get dqlite cluster members, retrying", "error", err)
				continue
			}

			// Check if the current node is present in the cluster
			for _, member := range members {
				if member.Address == nodeAddress {
					log.Info("node found in k8s-dqlite cluster", "nodeAddress", nodeAddress, "role", member.Role)
					return nil
				}
			}

			log.Info("node not yet found in k8s-dqlite cluster, retrying", "nodeAddress", nodeAddress)
		}
	}
}
