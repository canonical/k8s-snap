package app

import (
	"fmt"
	"net"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap"
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
	switch cfg.APIServer.Datastore {
	case "k8s-dqlite":
		certificates := pki.NewK8sDqlitePKI(pki.K8sDqlitePKIOpts{
			Hostname: s.Name(),
			IPSANs:   []net.IP{{127, 0, 0, 1}},
			Years:    20,
		})
		certificates.K8sDqliteCert = cfg.Certificates.K8sDqliteCert
		certificates.K8sDqliteKey = cfg.Certificates.K8sDqliteKey
		if err := certificates.CompleteCertificates(); err != nil {
			return fmt.Errorf("failed to initialize k8s-dqlite certificates: %w", err)
		}
		if err := setup.EnsureK8sDqlitePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write k8s-dqlite certificates: %w", err)
		}
	case "external":
		certificates := &pki.ExternalDatastorePKI{
			DatastoreCACert:     cfg.Certificates.DatastoreCACert,
			DatastoreClientCert: cfg.Certificates.DatastoreClientCert,
			DatastoreClientKey:  cfg.Certificates.DatastoreClientKey,
		}
		if err := certificates.CheckCertificates(); err != nil {
			return fmt.Errorf("failed to initialize external datastore certificates: %w", err)
		}
		if err := setup.EnsureExtDatastorePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write external datastore certificates: %w", err)
		}
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.APIServer.Datastore, setup.SupportedDatastores)
	}

	// Certificates
	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:                  s.Name(),
		IPSANs:                    append([]net.IP{nodeIP}, serviceIPs...),
		Years:                     20,
		IncludeMachineAddressSANs: true,
	})

	// load existing certificates, then generate certificates for the node
	certificates.CACert = cfg.Certificates.CACert
	certificates.CAKey = cfg.Certificates.CAKey
	certificates.FrontProxyCACert = cfg.Certificates.FrontProxyCACert
	certificates.FrontProxyCAKey = cfg.Certificates.FrontProxyCAKey
	certificates.APIServerKubeletClientCert = cfg.Certificates.APIServerKubeletClientCert
	certificates.APIServerKubeletClientKey = cfg.Certificates.APIServerKubeletClientKey
	certificates.ServiceAccountKey = cfg.APIServer.ServiceAccountKey

	if err := certificates.CompleteCertificates(); err != nil {
		return fmt.Errorf("failed to initialize control plane certificates: %w", err)
	}
	if err := setup.EnsureControlPlanePKI(snap, certificates); err != nil {
		return fmt.Errorf("failed to write control plane certificates: %w", err)
	}

	if err := setupKubeconfigs(s, snap.KubernetesConfigDir(), cfg.APIServer.SecurePort, cfg.Certificates.CACert); err != nil {
		return fmt.Errorf("failed to generate kubeconfigs: %w", err)
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
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.APIServer.Datastore, setup.SupportedDatastores)
	}

	// Configure services
	if err := setupControlPlaneServices(snap, s, cfg, nodeIP); err != nil {
		return fmt.Errorf("failed to configure services: %w", err)
	}

	// Start services
	if err := startControlPlaneServices(s.Context, snap, cfg.APIServer.Datastore); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait until Kube-API server is ready
	if err := waitReadyApiServer(s.Context, snap); err != nil {
		return fmt.Errorf("failed to wait for kube-apiserver: %w", err)
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
