package app

import (
	"context"
	"fmt"
	"net"
	"path"

	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/microcluster/state"
)

func setupKubeconfigs(s *state.State, kubeConfigDir string, securePort int, caCert string) error {
	// Generate kubeconfigs
	for _, kubeconfig := range []struct {
		file     string
		username string
		groups   []string
	}{
		{file: "admin.conf", username: "kubernetes-admin", groups: []string{"system:masters"}},
		{file: "controller.conf", username: "system:kube-controller-manager"},
		{file: "proxy.conf", username: "system:kube-proxy"},
		{file: "scheduler.conf", username: "system:kube-scheduler"},
		{file: "kubelet.conf", username: fmt.Sprintf("system:node:%s", s.Name()), groups: []string{"system:nodes"}},
	} {
		token, err := databaseutil.GetOrCreateAuthToken(s.Context, s, kubeconfig.username, kubeconfig.groups)
		if err != nil {
			return fmt.Errorf("failed to generate token for username=%s groups=%v: %w", kubeconfig.username, kubeconfig.groups, err)
		}
		if err := setup.Kubeconfig(path.Join(kubeConfigDir, kubeconfig.file), token, fmt.Sprintf("127.0.0.1:%d", securePort), caCert); err != nil {
			return fmt.Errorf("failed to write kubeconfig %s: %w", kubeconfig.file, err)
		}
	}
	return nil

}

func setupControlPlaneServices(snap snap.Snap, s *state.State, cfg types.ClusterConfig, nodeIP net.IP) error {
	// Configure services
	if err := setup.Containerd(snap, nil); err != nil {
		return fmt.Errorf("failed to configure containerd: %w", err)
	}
	if err := setup.KubeletControlPlane(snap, s.Name(), nodeIP, cfg.Kubelet.GetClusterDNS(), cfg.Kubelet.GetClusterDomain(), cfg.Kubelet.GetCloudProvider(), cfg.Kubelet.GetControlPlaneTaints()); err != nil {
		return fmt.Errorf("failed to configure kubelet: %w", err)
	}
	if err := setup.KubeProxy(s.Context, snap, s.Name(), cfg.Network.GetPodCIDR()); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}
	if err := setup.KubeControllerManager(snap); err != nil {
		return fmt.Errorf("failed to configure kube-controller-manager: %w", err)
	}
	if err := setup.KubeScheduler(snap); err != nil {
		return fmt.Errorf("failed to configure kube-scheduler: %w", err)
	}
	if err := setup.KubeAPIServer(snap, cfg.Network.GetServiceCIDR(), s.Address().Path("1.0", "kubernetes", "auth", "webhook").String(), true, cfg.Datastore, cfg.APIServer.GetAuthorizationMode()); err != nil {
		return fmt.Errorf("failed to configure kube-apiserver: %w", err)
	}
	return nil
}

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

func waitApiServerReady(ctx context.Context, snap snap.Snap) error {
	// Wait for API server to come up
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	if err := client.WaitApiServerReady(ctx); err != nil {
		return fmt.Errorf("kube-apiserver did not become ready in time: %w", err)
	}

	return nil
}
