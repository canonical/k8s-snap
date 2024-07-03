package app

import (
	"context"
	"fmt"
	"path"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/microcluster/state"
)

func setupKubeconfigs(s *state.State, kubeConfigDir string, securePort int, pki pki.ControlPlanePKI) error {
	// Generate kubeconfigs
	for _, kubeconfig := range []struct {
		file string
		crt  string
		key  string
	}{
		{file: "admin.conf", crt: pki.AdminClientCert, key: pki.AdminClientKey},
		{file: "controller.conf", crt: pki.KubeControllerManagerClientCert, key: pki.KubeControllerManagerClientKey},
		{file: "proxy.conf", crt: pki.KubeProxyClientCert, key: pki.KubeProxyClientKey},
		{file: "scheduler.conf", crt: pki.KubeSchedulerClientCert, key: pki.KubeSchedulerClientKey},
		{file: "kubelet.conf", crt: pki.KubeletClientCert, key: pki.KubeletClientKey},
	} {
		if err := setup.Kubeconfig(path.Join(kubeConfigDir, kubeconfig.file), fmt.Sprintf("127.0.0.1:%d", securePort), pki.CACert, kubeconfig.crt, kubeconfig.key); err != nil {
			return fmt.Errorf("failed to write kubeconfig %s: %w", kubeconfig.file, err)
		}
	}
	return nil

}

func startControlPlaneServices(ctx context.Context, snap snap.Snap, datastore string) error {
	// Start services
	switch datastore {
	case "k8s-dqlite", "etcd":
		if err := snaputil.StartK8sDBService(ctx, snap); err != nil {
			return fmt.Errorf("failed to start datastore: %w", err)
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
	case "k8s-dqlite", "etcd":
		if err := snaputil.StopK8sDBService(ctx, snap); err != nil {
			return fmt.Errorf("failed to stop datastore: %w", err)
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
