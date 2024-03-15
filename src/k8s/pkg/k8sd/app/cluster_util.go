package app

import (
	"context"
	"fmt"
	"net"
	"path"

	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/microcluster/state"
)

func setupDatastoreCertificates(snap snap.Snap, cfg types.ClusterConfig, nodeName string, allowSelfSignedCA bool) error {
	// Certificates
	switch cfg.APIServer.Datastore {
	case "k8s-dqlite":
		certificates := pki.NewK8sDqlitePKI(pki.K8sDqlitePKIOpts{
			Hostname:          nodeName,
			IPSANs:            []net.IP{{127, 0, 0, 1}},
			Years:             20,
			AllowSelfSignedCA: allowSelfSignedCA,
		})
		certificates.K8sDqliteCert = cfg.Certificates.K8sDqliteCert
		certificates.K8sDqliteKey = cfg.Certificates.K8sDqliteKey
		if err := certificates.CompleteCertificates(); err != nil {
			return fmt.Errorf("failed to initialize cluster certificates: %w", err)
		}
		if err := setup.EnsureK8sDqlitePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write cluster certificates: %w", err)
		}
	case "external":
		certificates := &pki.ExternalDatastorePKI{
			DatastoreCACert:     cfg.Certificates.DatastoreCACert,
			DatastoreClientCert: cfg.Certificates.DatastoreClientCert,
			DatastoreClientKey:  cfg.Certificates.DatastoreClientKey,
		}
		if err := certificates.CheckCertificates(); err != nil {
			return fmt.Errorf("failed to initialize cluster certificates: %w", err)
		}
		if err := setup.EnsureExtDatastorePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write cluster certificates: %w", err)
		}
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.APIServer.Datastore, setup.SupportedDatastores)
	}

	return nil

}

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
		token, err := impl.GetOrCreateAuthToken(s.Context, s, kubeconfig.username, kubeconfig.groups)
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
	if err := setup.KubeletControlPlane(snap, s.Name(), nodeIP, cfg.Kubelet.ClusterDNS, cfg.Kubelet.ClusterDomain, cfg.Kubelet.CloudProvider); err != nil {
		return fmt.Errorf("failed to configure kubelet: %w", err)
	}
	if err := setup.KubeProxy(s.Context, snap, s.Name(), cfg.Network.PodCIDR); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}
	if err := setup.KubeControllerManager(snap); err != nil {
		return fmt.Errorf("failed to configure kube-controller-manager: %w", err)
	}
	if err := setup.KubeScheduler(snap); err != nil {
		return fmt.Errorf("failed to configure kube-scheduler: %w", err)
	}
	if err := setup.KubeAPIServer(snap, cfg.Network.ServiceCIDR, s.Address().Path("1.0", "kubernetes", "auth", "webhook").String(), true, cfg.APIServer.Datastore, cfg.APIServer.DatastoreURL, cfg.APIServer.AuthorizationMode); err != nil {
		return fmt.Errorf("failed to configure kube-apiserver: %w", err)
	}
	return nil
}

func startServicesControlPlane(ctx context.Context, snap snap.Snap, datastore string) error {
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

	// Wait for API server to come up
	client, err := k8s.NewClient(snap)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	if err := client.WaitApiServerReady(ctx); err != nil {
		return fmt.Errorf("k8s api server did not become ready in time: %w", err)
	}

	return nil
}
