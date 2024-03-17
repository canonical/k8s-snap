package app

import (
	"fmt"
	"net"
	"path"

	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/microcluster/state"
)

// onPostJoin is called when a control plane node joins the cluster.
// onPostJoin retrieves the cluster config from the database and configures local services.
func onPostJoin(s *state.State, initConfig map[string]string) error {
	snap := snap.SnapFromContext(s.Context)

	cfg, err := utils.GetClusterConfig(s.Context, s)
	if err != nil {
		return fmt.Errorf("failed to get cluster config: %w", err)
	}
	nodeIP := net.ParseIP(s.Address().Hostname())
	if nodeIP == nil {
		return fmt.Errorf("failed to parse node IP address %q", s.Address().Hostname())
	}

	// Create directories
	if err := setup.EnsureAllDirectories(snap); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// cfg.Network.ServiceCIDR may be "IPv4CIDR[,IPv6CIDR]". get the first ip from CIDR(s).
	serviceIPs, err := utils.GetKubernetesServiceIPsFromServiceCIDRs(cfg.Network.GetServiceCIDR())
	if err != nil {
		return fmt.Errorf("failed to get IP address(es) from ServiceCIDR %q: %w", cfg.Network.GetServiceCIDR(), err)
	}

	switch cfg.Datastore.GetType() {
	case "k8s-dqlite":
		certificates := pki.NewK8sDqlitePKI(pki.K8sDqlitePKIOpts{
			Hostname: s.Name(),
			IPSANs:   []net.IP{{127, 0, 0, 1}},
			Years:    20,
		})
		certificates.K8sDqliteCert = cfg.Datastore.GetK8sDqliteCert()
		certificates.K8sDqliteKey = cfg.Datastore.GetK8sDqliteKey()
		if err := certificates.CompleteCertificates(); err != nil {
			return fmt.Errorf("failed to initialize cluster certificates: %w", err)
		}
		if err := setup.EnsureK8sDqlitePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write cluster certificates: %w", err)
		}
	case "external":
		certificates := &pki.ExternalDatastorePKI{
			DatastoreCACert:     cfg.Datastore.GetExternalCACert(),
			DatastoreClientCert: cfg.Datastore.GetExternalClientCert(),
			DatastoreClientKey:  cfg.Datastore.GetExternalClientKey(),
		}
		if err := certificates.CheckCertificates(); err != nil {
			return fmt.Errorf("failed to initialize cluster certificates: %w", err)
		}
		if err := setup.EnsureExtDatastorePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write cluster certificates: %w", err)
		}
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.Datastore.GetType(), setup.SupportedDatastores)
	}

	// Certificates
	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:                  s.Name(),
		IPSANs:                    append([]net.IP{nodeIP}, serviceIPs...),
		Years:                     20,
		IncludeMachineAddressSANs: true,
	})

	// load existing certificates, then generate certificates for the node
	certificates.CACert = cfg.Certificates.GetCACert()
	certificates.CAKey = cfg.Certificates.GetCAKey()
	certificates.FrontProxyCACert = cfg.Certificates.GetFrontProxyCACert()
	certificates.FrontProxyCAKey = cfg.Certificates.GetFrontProxyCAKey()
	certificates.APIServerKubeletClientCert = cfg.Certificates.GetAPIServerKubeletClientCert()
	certificates.APIServerKubeletClientKey = cfg.Certificates.GetAPIServerKubeletClientKey()
	certificates.ServiceAccountKey = cfg.Certificates.GetServiceAccountKey()

	if err := certificates.CompleteCertificates(); err != nil {
		return fmt.Errorf("failed to initialize cluster certificates: %w", err)
	}
	if err := setup.EnsureControlPlanePKI(snap, certificates); err != nil {
		return fmt.Errorf("failed to write cluster certificates: %w", err)
	}

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
		if err := setup.Kubeconfig(path.Join(snap.KubernetesConfigDir(), kubeconfig.file), token, fmt.Sprintf("127.0.0.1:%d", cfg.APIServer.GetSecurePort()), cfg.Certificates.GetCACert()); err != nil {
			return fmt.Errorf("failed to write kubeconfig %s: %w", kubeconfig.file, err)
		}
	}

	// Configure datastore
	switch cfg.Datastore.GetType() {
	case "k8s-dqlite":
		leader, err := s.Leader()
		if err != nil {
			return fmt.Errorf("failed to get dqlite leader: %w", err)
		}
		members, err := leader.GetClusterMembers(s.Context)
		if err != nil {
			return fmt.Errorf("failed to get microcluster members: %w", err)
		}
		cluster := make([]string, len(members))
		for _, member := range members {
			cluster = append(cluster, fmt.Sprintf("%s:%d", member.Address.Addr(), cfg.Datastore.GetK8sDqlitePort()))
		}

		address := fmt.Sprintf("%s:%d", nodeIP.String(), cfg.Datastore.GetK8sDqlitePort())
		if err := setup.K8sDqlite(snap, address, cluster); err != nil {
			return fmt.Errorf("failed to configure k8s-dqlite with address=%s cluster=%v: %w", address, cluster, err)
		}
	case "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.Datastore.GetType(), setup.SupportedDatastores)
	}

	// Configure services
	if err := setup.Containerd(snap, nil); err != nil {
		return fmt.Errorf("failed to configure containerd: %w", err)
	}
	if err := setup.KubeletControlPlane(snap, s.Name(), nodeIP, cfg.Kubelet.GetClusterDNS(), cfg.Kubelet.GetClusterDomain(), cfg.Kubelet.GetCloudProvider()); err != nil {
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

	if err := setup.KubeAPIServer(snap, cfg.Network.GetServiceCIDR(), s.Address().Path("1.0", "kubernetes", "auth", "webhook").String(), true, cfg.Datastore.GetType(), cfg.Datastore.GetExternalURL(), cfg.APIServer.GetAuthorizationMode()); err != nil {
		return fmt.Errorf("failed to configure kube-apiserver: %w", err)
	}

	// Start services
	switch cfg.Datastore.GetType() {
	case "k8s-dqlite":
		if err := snaputil.StartK8sDqliteServices(s.Context, snap); err != nil {
			return fmt.Errorf("failed to start control plane services: %w", err)
		}
	case "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.Datastore.GetType(), setup.SupportedDatastores)
	}

	if err := snaputil.StartControlPlaneServices(s.Context, snap); err != nil {
		return fmt.Errorf("failed to start control plane services: %w", err)
	}

	// Wait for API server to come up
	client, err := k8s.NewClient(snap)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if err := client.WaitApiServerReady(s.Context); err != nil {
		return fmt.Errorf("kube-apiserver did not become ready: %w", err)
	}

	return nil
}

func onPreRemove(s *state.State, force bool) error {
	snap := snap.SnapFromContext(s.Context)

	cfg, err := utils.GetClusterConfig(s.Context, s)
	if err != nil {
		return fmt.Errorf("failed to retrieve k8sd cluster config: %w", err)
	}

	// configure datastore
	switch cfg.Datastore.GetType() {
	case "k8s-dqlite":
		client, err := snap.K8sDqliteClient(s.Context)
		if err != nil {
			return fmt.Errorf("failed to create k8s-dqlite client: %w", err)
		}

		nodeAddress := net.JoinHostPort(s.Address().Hostname(), fmt.Sprintf("%d", cfg.Datastore.GetK8sDqlitePort()))
		if err := client.RemoveNodeByAddress(s.Context, nodeAddress); err != nil {
			return fmt.Errorf("failed to remove node with address %s from k8s-dqlite cluster: %w", nodeAddress, err)
		}
	case "external":
	default:
	}

	c, err := k8s.NewClient(snap)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	if err := c.DeleteNode(s.Context, s.Name()); err != nil {
		return fmt.Errorf("failed to remove k8s node %q: %w", s.Name(), err)
	}

	return nil
}
