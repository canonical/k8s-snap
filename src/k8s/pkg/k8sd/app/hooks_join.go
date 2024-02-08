package app

import (
	"fmt"
	"log"
	"net"
	"path"

	old_setup "github.com/canonical/k8s/pkg/k8s/setup"
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
	certificates := pki.NewControlPlanePKI(s.Name(), nil, []net.IP{nodeIP}, 10, false)

	// load existing certificates, then generate certificates for the node
	certificates.CACert = cfg.Certificates.CACert
	certificates.CAKey = cfg.Certificates.CAKey
	certificates.FrontProxyCACert = cfg.Certificates.FrontProxyCACert
	certificates.FrontProxyCAKey = cfg.Certificates.FrontProxyCAKey
	certificates.APIServerKubeletClientCert = cfg.Certificates.APIServerKubeletClientCert
	certificates.APIServerKubeletClientKey = cfg.Certificates.APIServerKubeletClientKey
	certificates.K8sDqliteCert = cfg.Certificates.K8sDqliteCert
	certificates.K8sDqliteKey = cfg.Certificates.K8sDqliteKey
	certificates.ServiceAccountKey = cfg.APIServer.ServiceAccountKey

	for action, f := range map[string]func() error{
		"initialize cluster certificates": func() error { return certificates.CompleteCertificates() },
		"create cluster directories":      func() error { return setup.EnsureAllDirectories(snap) },
		"write cluster certificates":      func() error { return setup.EnsureControlPlanePKI(snap, certificates) },
	} {
		if err := f(); err != nil {
			return fmt.Errorf("failed to %s: %w", action, err)
		}
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

	// join k8s-dqlite
	if cfg.APIServer.Datastore == "k8s-dqlite" {
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
	}

	for action, f := range map[string]func() error{
		"configure containerd": func() error { return setup.Containerd(snap) },
		"configure kubelet": func() error {
			return setup.Kubelet(snap, s.Name(), nodeIP, cfg.Kubelet.ClusterDNS, cfg.Kubelet.ClusterDomain, cfg.Kubelet.CloudProvider)
		},
		"configure kube-proxy":              func() error { return setup.KubeProxy(snap, s.Name(), cfg.Network.PodCIDR) },
		"configure kube-controller-manager": func() error { return setup.KubeControllerManager(snap) },
		"configure kube-scheduler":          func() error { return setup.KubeScheduler(snap) },
		"configure kube-apiserver": func() error {
			return setup.KubeAPIServer(snap, cfg.Network.ServiceCIDR, s.Address().Path("1.0/kubernetes/auth/webhook").String(), true, cfg.APIServer.Datastore, cfg.APIServer.AuthorizationMode)
		},
		"start control plane services": func() error { return snaputil.StartControlPlaneServices(s.Context, snap) },
	} {
		if err := f(); err != nil {
			return fmt.Errorf("failed to %s: %w", action, err)
		}
	}

	k8sClient, err := k8s.NewClient(snap)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	// The apiserver needs to be ready to start components.
	err = k8s.WaitApiServerReady(s.Context, k8sClient)
	if err != nil {
		return fmt.Errorf("k8s api server did not become ready in time: %w", err)
	}

	return nil
}

func onPreRemove(s *state.State, force bool) error {
	snap := snap.SnapFromContext(s.Context)

	// Remove k8s dqlite node from cluster.
	// Fails if the k8s-dqlite cluster would not have a leader afterwards.
	log.Println("Leave k8s-dqlite cluster")
	err := old_setup.LeaveK8sDqliteCluster(s.Context, snap, s)
	if err != nil {
		return fmt.Errorf("failed to leave k8s-dqlite cluster: %w", err)
	}

	// TODO: Remove node from kubernetes

	return nil
}
