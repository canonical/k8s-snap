package app

import (
	"fmt"
	"net"

	apiv1 "github.com/canonical/k8s/api/v1"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	"github.com/canonical/microcluster/state"
)

// onPostJoin is called when a control plane node joins the cluster.
// onPostJoin retrieves the cluster config from the database and configures local services.
func (a *App) onPostJoin(s *state.State, initConfig map[string]string) error {
	snap := a.Snap()

	joinConfig, err := apiv1.ControlPlaneJoinConfigFromMicrocluster(initConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal control plane join config: %w", err)
	}

	cfg, err := databaseutil.GetClusterConfig(s.Context, s)
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

	// Certificates
	extraIPs, extraNames := utils.SplitIPAndDNSSANs(joinConfig.ExtraSANS)

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
			return fmt.Errorf("failed to initialize k8s-dqlite certificates: %w", err)
		}
		if _, err := setup.EnsureK8sDqlitePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write k8s-dqlite certificates: %w", err)
		}
	case "embedded":
		certificates := pki.NewEtcdPKI(pki.EtcdPKIOpts{
			Hostname: s.Name(),
			IPSANs:   append([]net.IP{nodeIP}, extraIPs...),
			DNSSANs:  append([]string{s.Name()}, extraNames...),
			Years:    20,
		})

		certificates.CACert = cfg.Datastore.GetEmbeddedCACert()
		certificates.CAKey = cfg.Datastore.GetEmbeddedCAKey()
		certificates.ServerCert = joinConfig.GetEmbeddedServerCert()
		certificates.ServerKey = joinConfig.GetEmbeddedServerKey()
		certificates.ServerPeerCert = joinConfig.GetEmbeddedServerPeerCert()
		certificates.ServerPeerKey = joinConfig.GetEmbeddedServerPeerKey()
		certificates.APIServerClientCert = cfg.Datastore.GetEmbeddedAPIServerClientCert()
		certificates.APIServerClientKey = cfg.Datastore.GetEmbeddedAPIServerClientKey()

		if err := certificates.CompleteCertificates(); err != nil {
			return fmt.Errorf("failed to initialize embedded datastore certificates: %w", err)
		}
		if _, err := setup.EnsureEtcdPKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write embedded datastore certificates: %w", err)
		}

	case "external":
		certificates := &pki.ExternalDatastorePKI{
			DatastoreCACert:     cfg.Datastore.GetExternalCACert(),
			DatastoreClientCert: cfg.Datastore.GetExternalClientCert(),
			DatastoreClientKey:  cfg.Datastore.GetExternalClientKey(),
		}
		if err := certificates.CheckCertificates(); err != nil {
			return fmt.Errorf("failed to initialize external datastore certificates: %w", err)
		}
		if _, err := setup.EnsureExtDatastorePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write external datastore certificates: %w", err)
		}
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.Datastore.GetType(), setup.SupportedDatastores)
	}

	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:                  s.Name(),
		IPSANs:                    append(append([]net.IP{nodeIP}, serviceIPs...), extraIPs...),
		DNSSANs:                   extraNames,
		Years:                     20,
		IncludeMachineAddressSANs: true,
	})

	// load shared cluster certificates
	certificates.CACert = cfg.Certificates.GetCACert()
	certificates.CAKey = cfg.Certificates.GetCAKey()
	certificates.ClientCACert = cfg.Certificates.GetClientCACert()
	certificates.ClientCAKey = cfg.Certificates.GetClientCAKey()
	certificates.FrontProxyCACert = cfg.Certificates.GetFrontProxyCACert()
	certificates.FrontProxyCAKey = cfg.Certificates.GetFrontProxyCAKey()
	certificates.APIServerKubeletClientCert = cfg.Certificates.GetAPIServerKubeletClientCert()
	certificates.APIServerKubeletClientKey = cfg.Certificates.GetAPIServerKubeletClientKey()
	certificates.ServiceAccountKey = cfg.Certificates.GetServiceAccountKey()
	certificates.K8sdPublicKey = cfg.Certificates.GetK8sdPublicKey()
	certificates.K8sdPrivateKey = cfg.Certificates.GetK8sdPrivateKey()

	// load certificates from joinConfig
	certificates.APIServerCert = joinConfig.GetAPIServerCert()
	certificates.APIServerKey = joinConfig.GetAPIServerKey()
	certificates.FrontProxyClientCert = joinConfig.GetFrontProxyClientCert()
	certificates.FrontProxyClientKey = joinConfig.GetFrontProxyClientKey()
	certificates.KubeletCert = joinConfig.GetKubeletCert()
	certificates.KubeletKey = joinConfig.GetKubeletKey()

	// generate missing certificates
	if err := certificates.CompleteCertificates(); err != nil {
		return fmt.Errorf("failed to initialize control plane certificates: %w", err)
	}

	// Pre-init checks
	if err := snap.PreInitChecks(s.Context, cfg); err != nil {
		return fmt.Errorf("pre-init checks failed for joining node: %w", err)
	}

	// Write certificates to disk
	if _, err := setup.EnsureControlPlanePKI(snap, certificates); err != nil {
		return fmt.Errorf("failed to write control plane certificates: %w", err)
	}

	if err := setupKubeconfigs(s, snap.KubernetesConfigDir(), cfg.APIServer.GetSecurePort(), *certificates); err != nil {
		return fmt.Errorf("failed to generate kubeconfigs: %w", err)
	}

	// Configure datastore
	switch cfg.Datastore.GetType() {
	case "external":
		// no-op
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
			cluster = append(cluster, utils.JoinHostPort(member.Address.Addr().String(), cfg.Datastore.GetK8sDqlitePort()))
		}

		address := utils.JoinHostPort(nodeIP.String(), cfg.Datastore.GetK8sDqlitePort())
		if err := setup.K8sDqlite(snap, address, cluster, joinConfig.ExtraNodeK8sDqliteArgs); err != nil {
			return fmt.Errorf("failed to configure k8s-dqlite with address=%s cluster=%v: %w", address, cluster, err)
		}
	case "embedded":
		leader, err := s.Leader()
		if err != nil {
			return fmt.Errorf("failed to get dqlite leader: %w", err)
		}
		members, err := leader.GetClusterMembers(s.Context)
		if err != nil {
			return fmt.Errorf("failed to get microcluster members: %w", err)
		}
		clientURLs := make([]string, len(members))
		for _, member := range members {
			clientURLs = append(clientURLs, fmt.Sprintf("https://%s", utils.JoinHostPort(member.Address.Addr().String(), cfg.Datastore.GetEmbeddedPort())))
		}

		clientURL := fmt.Sprintf("https://%s", utils.JoinHostPort(nodeIP.String(), cfg.Datastore.GetEmbeddedPort()))
		peerURL := fmt.Sprintf("https://%s", utils.JoinHostPort(nodeIP.String(), cfg.Datastore.GetEmbeddedPeerPort()))
		if err := setup.K8sDqliteEmbedded(snap, s.Name(), clientURL, peerURL, clientURLs, joinConfig.ExtraNodeK8sDqliteArgs); err != nil {
			return fmt.Errorf("failed to config k8s-dqlite embedded with peerURL=%s cluster=%v: %w", peerURL, clientURLs, err)
		}
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.Datastore.GetType(), setup.SupportedDatastores)
	}

	// Configure services
	if err := setup.Containerd(snap, nil, joinConfig.ExtraNodeContainerdArgs); err != nil {
		return fmt.Errorf("failed to configure containerd: %w", err)
	}
	if err := setup.KubeletControlPlane(snap, s.Name(), nodeIP, cfg.Kubelet.GetClusterDNS(), cfg.Kubelet.GetClusterDomain(), cfg.Kubelet.GetCloudProvider(), cfg.Kubelet.GetControlPlaneTaints(), joinConfig.ExtraNodeKubeletArgs); err != nil {
		return fmt.Errorf("failed to configure kubelet: %w", err)
	}
	if err := setup.KubeProxy(s.Context, snap, s.Name(), cfg.Network.GetPodCIDR(), joinConfig.ExtraNodeKubeProxyArgs); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}
	if err := setup.KubeControllerManager(snap, joinConfig.ExtraNodeKubeControllerManagerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-controller-manager: %w", err)
	}
	if err := setup.KubeScheduler(snap, joinConfig.ExtraNodeKubeSchedulerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-scheduler: %w", err)
	}
	if err := setup.KubeAPIServer(snap, cfg.Network.GetServiceCIDR(), s.Address().Path("1.0", "kubernetes", "auth", "webhook").String(), true, cfg.Datastore, cfg.APIServer.GetAuthorizationMode(), s.Address().Hostname(), joinConfig.ExtraNodeKubeAPIServerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-apiserver: %w", err)
	}

	if err := setup.ExtraNodeConfigFiles(snap, joinConfig.ExtraNodeConfigFiles); err != nil {
		return fmt.Errorf("failed to write extra node config files: %w", err)
	}

	if err := snapdconfig.SetSnapdFromK8sd(s.Context, cfg.ToUserFacing(), snap); err != nil {
		return fmt.Errorf("failed to set snapd configuration from k8sd: %w", err)
	}

	// Start services
	if err := startControlPlaneServices(s.Context, snap, cfg.Datastore.GetType()); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait until Kube-API server is ready
	if err := waitApiServerReady(s.Context, snap); err != nil {
		return fmt.Errorf("failed to wait for kube-apiserver to become ready: %w", err)
	}

	return nil
}

func (a *App) onPreRemove(s *state.State, force bool) error {
	snap := a.Snap()

	cfg, err := databaseutil.GetClusterConfig(s.Context, s)
	if err != nil {
		return fmt.Errorf("failed to retrieve k8sd cluster config: %w", err)
	}

	// configure datastore
	switch cfg.Datastore.GetType() {
	case "external":
		// no-op
		c, err := snap.KubernetesClient("")
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}
		if err := c.DeleteNode(s.Context, s.Name()); err != nil {
			return fmt.Errorf("failed to remove k8s node %q: %w", s.Name(), err)
		}
	case "k8s-dqlite":
		client, err := snap.K8sDqliteClient(s.Context)
		if err != nil {
			return fmt.Errorf("failed to create k8s-dqlite client: %w", err)
		}

		nodeAddress := utils.JoinHostPort(s.Address().Hostname(), cfg.Datastore.GetK8sDqlitePort())
		if err := client.RemoveNodeByAddress(s.Context, nodeAddress); err != nil {
			return fmt.Errorf("failed to remove node with address %s from k8s-dqlite cluster: %w", nodeAddress, err)
		}

		c, err := snap.KubernetesClient("")
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}
		if err := c.DeleteNode(s.Context, s.Name()); err != nil {
			return fmt.Errorf("failed to remove k8s node %q: %w", s.Name(), err)
		}
	case "embedded":
		// for embedded, we first delete the kubernetes node and then proceed with removing the node from the embedded cluster
		c, err := snap.KubernetesClient("")
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}
		if err := c.DeleteNode(s.Context, s.Name()); err != nil {
			return fmt.Errorf("failed to remove k8s node %q: %w", s.Name(), err)
		}
		client := snap.EmbeddedClient()
		nodeAddress := fmt.Sprintf("https://%s", utils.JoinHostPort(s.Address().Hostname(), cfg.Datastore.GetEmbeddedPeerPort()))
		if err := client.RemoveNodeByAddress(s.Context, nodeAddress); err != nil {
			return fmt.Errorf("failed to remove node with address %s from embedded cluster: %w", nodeAddress, err)
		}
	default:
	}
	fmt.Println("RUNNING ON NODE", s.Name, s.Address().Hostname())

	return nil
}
