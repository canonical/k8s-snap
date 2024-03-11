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
	serviceIPs, err := utils.GetKubernetesServiceIPsFromServiceCIDRs(cfg.Network.ServiceCIDR)
	if err != nil {
		return fmt.Errorf("failed to get IP address(es) from ServiceCIDR %q: %w", cfg.Network.ServiceCIDR, err)
	}

	// Certificates
	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:  s.Name(),
		IPSANs:    append([]net.IP{nodeIP}, serviceIPs...),
		Years:     10,
		Datastore: cfg.APIServer.Datastore,
	})

	// load existing certificates, then generate certificates for the node
	certificates.CACert = cfg.Certificates.CACert
	certificates.CAKey = cfg.Certificates.CAKey
	certificates.FrontProxyCACert = cfg.Certificates.FrontProxyCACert
	certificates.FrontProxyCAKey = cfg.Certificates.FrontProxyCAKey
	certificates.APIServerKubeletClientCert = cfg.Certificates.APIServerKubeletClientCert
	certificates.APIServerKubeletClientKey = cfg.Certificates.APIServerKubeletClientKey
	certificates.ServiceAccountKey = cfg.APIServer.ServiceAccountKey

	switch cfg.APIServer.Datastore {
	case "k8s-dqlite":
		certificates.K8sDqliteCert = cfg.Certificates.K8sDqliteCert
		certificates.K8sDqliteKey = cfg.Certificates.K8sDqliteKey
	case "external-etcd":
		certificates.DatastoreCACert = cfg.Certificates.DatastoreCACert
		certificates.DatastoreClientCert = cfg.Certificates.DatastoreClientCert
		certificates.DatastoreClientKey = cfg.Certificates.DatastoreClientKey
	default:
		return fmt.Errorf("unsupported datastore %q, must be one of 'k8s-dqlite, external-etcd'", cfg.APIServer.Datastore)
	}

	if err := certificates.CompleteCertificates(); err != nil {
		return fmt.Errorf("failed to initialize cluster certificates: %w", err)
	}
	if err := setup.EnsureControlPlanePKI(snap, certificates, cfg.APIServer.Datastore); err != nil {
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
		if err := setup.Kubeconfig(path.Join(snap.KubernetesConfigDir(), kubeconfig.file), token, fmt.Sprintf("127.0.0.1:%d", cfg.APIServer.SecurePort), cfg.Certificates.CACert); err != nil {
			return fmt.Errorf("failed to write kubeconfig %s: %w", kubeconfig.file, err)
		}
	}

	// Configure datastore
	switch cfg.APIServer.Datastore {
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
			cluster = append(cluster, fmt.Sprintf("%s:%d", member.Address.Addr(), cfg.K8sDqlite.Port))
		}

		address := fmt.Sprintf("%s:%d", nodeIP.String(), cfg.K8sDqlite.Port)
		if err := setup.K8sDqlite(snap, address, cluster); err != nil {
			return fmt.Errorf("failed to configure k8s-dqlite with address=%s cluster=%v: %w", address, cluster, err)
		}
	default:
		return fmt.Errorf("unsupported datastore %q, must be one of 'k8s-dqlite, external-etcd'", cfg.APIServer.Datastore)
	}

	// Configure services
	if err := setup.Containerd(snap); err != nil {
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

	var datastoreUrl = cfg.APIServer.DatastoreURL
	switch cfg.APIServer.Datastore {
	case "k8s-dqlite":
		datastoreUrl = fmt.Sprintf("unix://%s", path.Join(snap.K8sDqliteStateDir(), "k8s-dqlite.sock"))
	case "external-etcd":
	default:
		return fmt.Errorf("unsupported datastore %s. must be one of 'k8s-dqlite, external-etcd'", cfg.APIServer.Datastore)
	}

	if err := setup.KubeAPIServer(snap, cfg.Network.ServiceCIDR, s.Address().Path("1.0", "kubernetes", "auth", "webhook").String(), true, cfg.APIServer.Datastore, datastoreUrl, cfg.APIServer.AuthorizationMode); err != nil {
		return fmt.Errorf("failed to configure kube-apiserver: %w", err)
	}

	// Start services
	if err := snaputil.StartControlPlaneServices(s.Context, snap, cfg.APIServer.Datastore); err != nil {
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
	switch cfg.APIServer.Datastore {
	case "k8s-dqlite":
		client, err := snap.K8sDqliteClient(s.Context)
		if err != nil {
			return fmt.Errorf("failed to create k8s-dqlite client: %w", err)
		}

		nodeAddress := net.JoinHostPort(s.Address().Hostname(), fmt.Sprintf("%d", cfg.K8sDqlite.Port))
		if err := client.RemoveNodeByAddress(s.Context, nodeAddress); err != nil {
			return fmt.Errorf("failed to remove node with address %s from k8s-dqlite cluster: %w", nodeAddress, err)
		}
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
