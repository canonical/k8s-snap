package app

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/component"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/k8s/pkg/utils/vals"
	"github.com/canonical/microcluster/state"
	"github.com/mitchellh/mapstructure"
)

// onBootstrap is called after we bootstrap the first cluster node.
// onBootstrap configures local services then writes the cluster config on the database.
func (a *App) onBootstrap(s *state.State, initConfig map[string]string) error {
	if workerToken, ok := initConfig["workerToken"]; ok {
		return a.onBootstrapWorkerNode(s, workerToken)
	}

	return a.onBootstrapControlPlane(s, initConfig)
}

func (a *App) onBootstrapWorkerNode(s *state.State, encodedToken string) error {
	token := &types.InternalWorkerNodeToken{}
	if err := token.Decode(encodedToken); err != nil {
		return fmt.Errorf("failed to parse worker token: %w", err)
	}

	if len(token.JoinAddresses) == 0 {
		return fmt.Errorf("empty list of control plane addresses")
	}
	nodeIP := net.ParseIP(s.Address().Hostname())
	if nodeIP == nil {
		return fmt.Errorf("failed to parse node IP address %s", s.Address().Hostname())
	}

	// TODO(neoaggelos): figure out how to use the microcluster client instead

	// create an HTTP client that ignores https
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	type wrappedResponse struct {
		Error    string                       `json:"error"`
		Metadata apiv1.WorkerNodeInfoResponse `json:"metadata"`
	}

	requestBody, err := json.Marshal(apiv1.WorkerNodeInfoRequest{Address: nodeIP.String()})
	if err != nil {
		return fmt.Errorf("failed to prepare worker info request: %w", err)
	}

	httpRequest, err := http.NewRequest("POST", fmt.Sprintf("https://%s/1.0/k8sd/worker/info", token.JoinAddresses[0]), bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to prepare HTTP request: %w", err)
	}
	httpRequest.Header.Add("worker-name", s.Name())
	httpRequest.Header.Add("worker-token", token.Secret)

	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("failed to POST %s: %w", httpRequest.URL.String(), err)
	}
	defer httpResponse.Body.Close()
	var wrappedResp wrappedResponse
	if err := json.NewDecoder(httpResponse.Body).Decode(&wrappedResp); err != nil {
		return fmt.Errorf("failed to parse HTTP response: %w", err)
	}
	if httpResponse.StatusCode != 200 {
		return fmt.Errorf("HTTP request for worker node info failed: %s", wrappedResp.Error)
	}
	response := wrappedResp.Metadata

	// Create directories
	if err := setup.EnsureAllDirectories(a.Snap()); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Certificates
	certificates := &pki.WorkerNodePKI{
		CACert:      response.CA,
		KubeletCert: response.KubeletCert,
		KubeletKey:  response.KubeletKey,
	}
	if err := certificates.CompleteCertificates(); err != nil {
		return fmt.Errorf("failed to initialize worker node certificates: %w", err)
	}
	if err := setup.EnsureWorkerPKI(a.Snap(), certificates); err != nil {
		return fmt.Errorf("failed to write worker node certificates: %w", err)
	}

	// Kubeconfigs
	if err := setup.Kubeconfig(path.Join(a.Snap().KubernetesConfigDir(), "kubelet.conf"), response.KubeletToken, "127.0.0.1:6443", certificates.CACert); err != nil {
		return fmt.Errorf("failed to generate kubelet kubeconfig: %w", err)
	}
	if err := setup.Kubeconfig(path.Join(a.Snap().KubernetesConfigDir(), "proxy.conf"), response.KubeProxyToken, "127.0.0.1:6443", certificates.CACert); err != nil {
		return fmt.Errorf("failed to generate kube-proxy kubeconfig: %w", err)
	}

	// Worker node services
	if err := setup.Containerd(a.Snap(), nil); err != nil {
		return fmt.Errorf("failed to configure containerd: %w", err)
	}
	if err := setup.KubeletWorker(a.Snap(), s.Name(), nodeIP, response.ClusterDNS, response.ClusterDomain, response.CloudProvider); err != nil {
		return fmt.Errorf("failed to configure kubelet: %w", err)
	}
	if err := setup.KubeProxy(s.Context, a.Snap(), s.Name(), response.PodCIDR); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}
	if err := setup.K8sAPIServerProxy(a.Snap(), response.APIServers); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}

	// TODO(berkayoz): remove the lock on cleanup
	if err := snaputil.MarkAsWorkerNode(a.Snap(), true); err != nil {
		return fmt.Errorf("failed to mark node as worker: %w", err)
	}

	// Start services
	if err := snaputil.StartWorkerServices(s.Context, a.Snap()); err != nil {
		return fmt.Errorf("failed to start worker services: %w", err)
	}

	return nil
}

func (a *App) onBootstrapControlPlane(s *state.State, initConfig map[string]string) error {
	bootstrapConfig, err := apiv1.BootstrapConfigFromMap(initConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bootstrap config: %w", err)
	}

	cfg := types.ClusterConfigFromBootstrapConfig(bootstrapConfig)
	cfg.SetDefaults()
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid cluster configuration: %w", err)
	}

	nodeIP := net.ParseIP(s.Address().Hostname())
	if nodeIP == nil {
		return fmt.Errorf("failed to parse node IP address %q", s.Address().Hostname())
	}

	// Create directories
	if err := setup.EnsureAllDirectories(a.Snap()); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// cfg.Network.ServiceCIDR may be "IPv4CIDR[,IPv6CIDR]". get the first ip from CIDR(s).
	serviceIPs, err := utils.GetKubernetesServiceIPsFromServiceCIDRs(cfg.Network.ServiceCIDR)
	if err != nil {
		return fmt.Errorf("failed to get IP address(es) from ServiceCIDR %q: %w", cfg.Network.ServiceCIDR, err)
	}

	switch cfg.APIServer.Datastore {
	case "k8s-dqlite":
		certificates := pki.NewK8sDqlitePKI(pki.K8sDqlitePKIOpts{
			Hostname:          s.Name(),
			IPSANs:            []net.IP{{127, 0, 0, 1}},
			Years:             20,
			AllowSelfSignedCA: true,
		})
		if err := certificates.CompleteCertificates(); err != nil {
			return fmt.Errorf("failed to initialize k8s-dqlite certificates: %w", err)
		}
		if err := setup.EnsureK8sDqlitePKI(a.Snap(), certificates); err != nil {
			return fmt.Errorf("failed to write k8s-dqlite certificates: %w", err)
		}

		cfg.Certificates.K8sDqliteCert = certificates.K8sDqliteCert
		cfg.Certificates.K8sDqliteKey = certificates.K8sDqliteKey

	case "external":
		certificates := &pki.ExternalDatastorePKI{
			DatastoreCACert:     cfg.Certificates.DatastoreCACert,
			DatastoreClientCert: cfg.Certificates.DatastoreClientCert,
			DatastoreClientKey:  cfg.Certificates.DatastoreClientKey,
		}
		if err := certificates.CheckCertificates(); err != nil {
			return fmt.Errorf("failed to initialize external datastore certificates: %w", err)
		}
		if err := setup.EnsureExtDatastorePKI(a.Snap(), certificates); err != nil {
			return fmt.Errorf("failed to write external datastore certificates: %w", err)
		}
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.APIServer.Datastore, setup.SupportedDatastores)
	}

	extraIPs, extraNames := utils.SplitIPAndDNSSANs(bootstrapConfig.ExtraSANs)

	IPSANs := append(append([]net.IP{nodeIP}, serviceIPs...), extraIPs...)

	// Certificates
	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:                  s.Name(),
		IPSANs:                    IPSANs,
		DNSSANs:                   extraNames,
		Years:                     20,
		AllowSelfSignedCA:         true,
		IncludeMachineAddressSANs: true,
	})
	if err := certificates.CompleteCertificates(); err != nil {
		return fmt.Errorf("failed to initialize control plane certificates: %w", err)
	}
	if err := setup.EnsureControlPlanePKI(a.Snap(), certificates); err != nil {
		return fmt.Errorf("failed to write control plane certificates: %w", err)
	}

	// Add certificates to the cluster config
	cfg.Certificates.CACert = certificates.CACert
	cfg.Certificates.CAKey = certificates.CAKey
	cfg.Certificates.FrontProxyCACert = certificates.FrontProxyCACert
	cfg.Certificates.FrontProxyCAKey = certificates.FrontProxyCAKey
	cfg.Certificates.APIServerKubeletClientCert = certificates.APIServerKubeletClientCert
	cfg.Certificates.APIServerKubeletClientKey = certificates.APIServerKubeletClientKey
	cfg.APIServer.ServiceAccountKey = certificates.ServiceAccountKey

	// Generate kubeconfigs
	if err := setupKubeconfigs(s, a.Snap().KubernetesConfigDir(), cfg.APIServer.SecurePort, cfg.Certificates.CACert); err != nil {
		return fmt.Errorf("failed to generate kubeconfigs: %w", err)
	}

	// Configure datastore
	switch cfg.APIServer.Datastore {
	case "k8s-dqlite":
		if err := setup.K8sDqlite(a.Snap(), fmt.Sprintf("%s:%d", nodeIP.String(), cfg.K8sDqlite.Port), nil); err != nil {
			return fmt.Errorf("failed to configure k8s-dqlite: %w", err)
		}
	case "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.APIServer.Datastore, setup.SupportedDatastores)
	}

	// Configure services
	if err := setupControlPlaneServices(a.Snap(), s, cfg, nodeIP); err != nil {
		return fmt.Errorf("failed to configure services: %w", err)
	}

	// Write cluster configuration to dqlite
	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		if err := database.SetClusterConfig(ctx, tx, cfg); err != nil {
			return fmt.Errorf("failed to write cluster configuration: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database transaction to update cluster configuration failed: %w", err)
	}

	// Start services
	if err := startControlPlaneServices(s.Context, a.Snap(), cfg.APIServer.Datastore); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait until Kube-API server is ready
	if err := waitApiServerReady(s.Context, a.Snap()); err != nil {
		return fmt.Errorf("kube-apiserver did not become ready in time: %w", err)
	}

	if cfg.Network.Enabled != nil {
		err := component.ReconcileNetworkComponent(s.Context, a.Snap(), vals.Pointer(false), cfg.Network.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile network: %w", err)
		}
	}

	if cfg.DNS.Enabled != nil {
		dnsIP, _, err := component.ReconcileDNSComponent(s.Context, a.Snap(), vals.Pointer(false), cfg.DNS.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile dns: %w", err)
		}
		if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
			if err := database.SetClusterConfig(ctx, tx, types.ClusterConfig{
				Kubelet: types.Kubelet{
					ClusterDNS: dnsIP,
				},
			}); err != nil {
				return fmt.Errorf("failed to update cluster configuration for dns=%s: %w", dnsIP, err)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("database transaction to update cluster configuration failed: %w", err)
		}

		var data map[string]string
		if err := mapstructure.Decode(types.NodeConfig{
			ClusterDNS:    dnsIP,
			ClusterDomain: cfg.Kubelet.ClusterDomain,
		}, &data); err != nil {
			return fmt.Errorf("failed to encode node config: %w", err)
		}

		client, err := k8s.NewClient(a.Snap().KubernetesRESTClientGetter(""))
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", err)
		}

		if _, err := client.UpdateConfigMap(s.Context, "kube-system", "k8sd-config", data); err != nil {
			return fmt.Errorf("failed to update node configs: %w", err)
		}
	}

	if cfg.LocalStorage.Enabled != nil {
		err := component.ReconcileLocalStorageComponent(s.Context, a.Snap(), vals.Pointer(false), cfg.LocalStorage.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile local-storage: %w", err)
		}
	}

	if cfg.Gateway.Enabled != nil {
		err := component.ReconcileGatewayComponent(s.Context, a.Snap(), vals.Pointer(false), cfg.Gateway.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile gateway: %w", err)
		}
	}

	if cfg.Ingress.Enabled != nil {
		err := component.ReconcileIngressComponent(s.Context, a.Snap(), vals.Pointer(false), cfg.Ingress.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile ingress: %w", err)
		}
	}

	if cfg.LoadBalancer.Enabled != nil {
		err := component.ReconcileLoadBalancerComponent(s.Context, a.Snap(), vals.Pointer(false), cfg.LoadBalancer.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile load-balancer: %w", err)
		}
	}

	if cfg.MetricsServer.Enabled != nil {
		err := component.ReconcileMetricsServerComponent(s.Context, a.Snap(), vals.Pointer(false), cfg.MetricsServer.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile metrics-server: %w", err)
		}
	}

	return nil
}
