package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	apiv1 "github.com/canonical/k8s/api/v1"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	"github.com/canonical/microcluster/cluster"
	"github.com/canonical/microcluster/state"
)

// onPostJoin is called when a control plane node joins the cluster.
// onPostJoin retrieves the cluster config from the database and configures local services.
func (a *App) onPostJoin(s *state.State, initConfig map[string]string) (rerr error) {
	snap := a.Snap()

	// NOTE(neoaggelos): context timeout is passed over configuration, so that hook failures are propagated to the client
	ctx, cancel := context.WithCancel(s.Context)
	defer cancel()
	if t := utils.MicroclusterTimeoutFromConfig(initConfig); t != 0 {
		ctx, cancel = context.WithTimeout(ctx, t)
		defer cancel()
	}

	// make sure to cleanup in case of errors
	// the code can register cleanup hooks by appending to this slice
	var cleanups []func(context.Context) error
	defer func() {
		// do not cleanup if joining was successful
		if rerr == nil {
			log.Println("Joined cluster successfully")
			return
		}

		// annotate error with context cancellation
		if err := ctx.Err(); err != nil {
			rerr = fmt.Errorf("%w: %v", rerr, ctx.Err())
		}

		// start goroutine to cleanup on the background and return quickly
		go func() {
			log.Printf("Join cluster failed: %v", rerr)

			log.Printf("Waiting for node to finish microcluster join")
			control.WaitUntilReady(s.Context, func() (bool, error) {
				var notPending bool
				if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
					member, err := cluster.GetInternalClusterMember(ctx, tx, s.Name())
					if err != nil {
						log.Printf("Failed to get member: %v", err)
						return nil
					}
					notPending = member.Role != cluster.Pending
					return nil
				}); err != nil {
					log.Printf("Transaction to check cluster member role failed: %v", err)
				}
				return notPending, nil
			})

			log.Println("Cleaning up...")
			for i := len(cleanups) - 1; i >= 0; i-- {
				// run cleanup functions in reverse order
				if err := cleanups[i](s.Context); err != nil {
					log.Printf("Cleanup hook %d/%d failed: %v", i, len(cleanups), err)
				}
			}
			log.Println("All cleanup hooks finished, removing node from microcluster")

			// NOTE(neoaggelos): this also runs the pre-remove hook and resets the cluster member
			control.WaitUntilReady(s.Context, func() (bool, error) {
				client, err := s.Leader()
				if err != nil {
					log.Printf("Error: failed to create client to leader: %v", err)
					return false, nil
				}
				if err := client.DeleteClusterMember(s.Context, s.Name(), true); err != nil {
					log.Printf("Error: failed to DeleteClusterMember: %v", err)
					return false, nil
				}
				return true, nil
			})
		}()
	}()

	joinConfig, err := apiv1.ControlPlaneJoinConfigFromMicrocluster(initConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal control plane join config: %w", err)
	}

	cfg, err := databaseutil.GetClusterConfig(ctx, s)
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
			return fmt.Errorf("failed to initialize k8s-dqlite certificates: %w", err)
		}
		if _, err := setup.EnsureK8sDqlitePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write k8s-dqlite certificates: %w", err)
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

	// Certificates
	extraIPs, extraNames := utils.SplitIPAndDNSSANs(joinConfig.ExtraSANS)
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
	if err := snap.PreInitChecks(ctx, cfg); err != nil {
		return fmt.Errorf("pre-init checks failed for joining node: %w", err)
	}

	// Write certificates to disk
	cleanups = append(cleanups, func(ctx context.Context) error {
		log.Println("Cleaning up control plane certificates")
		if _, err := setup.EnsureControlPlanePKI(snap, &pki.ControlPlanePKI{}); err != nil {
			return fmt.Errorf("failed to cleanup control plane certificates: %w", err)
		}
		return nil
	})
	if _, err := setup.EnsureControlPlanePKI(snap, certificates); err != nil {
		return fmt.Errorf("failed to write control plane certificates: %w", err)
	}

	if err := setupKubeconfigs(s, snap.KubernetesConfigDir(), cfg.APIServer.GetSecurePort(), *certificates); err != nil {
		return fmt.Errorf("failed to generate kubeconfigs: %w", err)
	}

	// Configure datastore
	switch cfg.Datastore.GetType() {
	case "k8s-dqlite":
		// TODO(neoaggelos): use cluster.GetInternalClusterMembers() instead
		leader, err := s.Leader()
		if err != nil {
			return fmt.Errorf("failed to get dqlite leader: %w", err)
		}
		members, err := leader.GetClusterMembers(ctx)
		if err != nil {
			return fmt.Errorf("failed to get microcluster members: %w", err)
		}
		cluster := make([]string, len(members))
		for _, member := range members {
			cluster = append(cluster, fmt.Sprintf("%s:%d", member.Address.Addr(), cfg.Datastore.GetK8sDqlitePort()))
		}

		address := fmt.Sprintf("%s:%d", nodeIP.String(), cfg.Datastore.GetK8sDqlitePort())
		if err := setup.K8sDqlite(snap, address, cluster, joinConfig.ExtraNodeK8sDqliteArgs); err != nil {
			return fmt.Errorf("failed to configure k8s-dqlite with address=%s cluster=%v: %w", address, cluster, err)
		}
	case "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.Datastore.GetType(), setup.SupportedDatastores)
	}

	cleanups = append(cleanups, func(ctx context.Context) error {
		for _, dir := range []string{snap.ServiceArgumentsDir()} {
			log.Printf("Cleaning up config files from %v", dir)
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("failed to delete %v: %w", dir, err)
			}
		}
		return nil
	})

	// Configure services
	if err := setup.Containerd(snap, nil, joinConfig.ExtraNodeContainerdArgs); err != nil {
		return fmt.Errorf("failed to configure containerd: %w", err)
	}
	if err := setup.KubeletControlPlane(snap, s.Name(), nodeIP, cfg.Kubelet.GetClusterDNS(), cfg.Kubelet.GetClusterDomain(), cfg.Kubelet.GetCloudProvider(), cfg.Kubelet.GetControlPlaneTaints(), joinConfig.ExtraNodeKubeletArgs); err != nil {
		return fmt.Errorf("failed to configure kubelet: %w", err)
	}
	if err := setup.KubeProxy(ctx, snap, s.Name(), cfg.Network.GetPodCIDR(), joinConfig.ExtraNodeKubeProxyArgs); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}
	if err := setup.KubeControllerManager(snap, joinConfig.ExtraNodeKubeControllerManagerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-controller-manager: %w", err)
	}
	if err := setup.KubeScheduler(snap, joinConfig.ExtraNodeKubeSchedulerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-scheduler: %w", err)
	}
	if err := setup.KubeAPIServer(snap, cfg.Network.GetServiceCIDR(), s.Address().Path("1.0", "kubernetes", "auth", "webhook").String(), true, cfg.Datastore, cfg.APIServer.GetAuthorizationMode(), joinConfig.ExtraNodeKubeAPIServerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-apiserver: %w", err)
	}

	if err := setup.ExtraNodeConfigFiles(snap, joinConfig.ExtraNodeConfigFiles); err != nil {
		return fmt.Errorf("failed to write extra node config files: %w", err)
	}

	if err := snapdconfig.SetSnapdFromK8sd(ctx, cfg.ToUserFacing(), snap); err != nil {
		return fmt.Errorf("failed to set snapd configuration from k8sd: %w", err)
	}

	// Start services
	cleanups = append(cleanups, func(ctx context.Context) error {
		log.Println("Stopping control plane services")
		if err := stopControlPlaneServices(ctx, snap, cfg.Datastore.GetType()); err != nil {
			return fmt.Errorf("failed to stop services: %w", err)
		}
		return nil
	})
	if err := startControlPlaneServices(ctx, snap, cfg.Datastore.GetType()); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait until Kube-API server is ready
	if err := waitApiServerReady(ctx, snap); err != nil {
		return fmt.Errorf("failed to wait for kube-apiserver to become ready: %w", err)
	}

	return nil
}

func (a *App) onPreRemove(s *state.State, force bool) (rerr error) {
	snap := a.Snap()

	// NOTE(neoaggelos): When the pre-remove hook fails, the microcluster node will
	// be removed from the cluster members, but remains in the microcluster dqlite database.
	//
	// Log the error and proceed, such that the node is in fact removed.
	defer func() {
		if rerr != nil {
			log.Printf("WARNING: There was an error when running the pre-remove hook: %v", rerr)
		}
		rerr = nil
	}()

	cfg, err := databaseutil.GetClusterConfig(s.Context, s)
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

	c, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	if err := c.DeleteNode(s.Context, s.Name()); err != nil {
		return fmt.Errorf("failed to remove k8s node %q: %w", s.Name(), err)
	}

	return nil
}
