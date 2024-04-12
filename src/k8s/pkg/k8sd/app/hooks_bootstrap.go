package app

import (
	"bytes"
	"context"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path"

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
	snap := a.Snap()

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

	// Get remote certificate from the cluster member
	cert, err := utils.GetRemoteCertificate(token.JoinAddresses[0])
	if err != nil {
		return fmt.Errorf("failed to get certificate of cluster member: %w", err)
	}

	// verify that the fingerprint of the certificate matches the fingerprint of the token
	fingerprint := utils.CertFingerprint(cert)
	if fingerprint != token.Fingerprint {
		return fmt.Errorf("fingerprint from token (%q) does not match fingerprint of node %q (%q)", token.Fingerprint, token.JoinAddresses[0], fingerprint)
	}

	// Create the http client with trusted certificate
	tlsConfig, err := utils.TLSClientConfigWithTrustedCertificate(cert, x509.NewCertPool())
	if err != nil {
		return fmt.Errorf("failed to get TLS configuration for trusted certificate: %w", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
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
	if err := setup.EnsureAllDirectories(snap); err != nil {
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
	if err := setup.EnsureWorkerPKI(snap, certificates); err != nil {
		return fmt.Errorf("failed to write worker node certificates: %w", err)
	}

	// Kubeconfigs
	if err := setup.Kubeconfig(path.Join(snap.KubernetesConfigDir(), "kubelet.conf"), response.KubeletToken, "127.0.0.1:6443", certificates.CACert); err != nil {
		return fmt.Errorf("failed to generate kubelet kubeconfig: %w", err)
	}
	if err := setup.Kubeconfig(path.Join(snap.KubernetesConfigDir(), "proxy.conf"), response.KubeProxyToken, "127.0.0.1:6443", certificates.CACert); err != nil {
		return fmt.Errorf("failed to generate kube-proxy kubeconfig: %w", err)
	}

	// Worker node services
	if err := setup.Containerd(snap, nil); err != nil {
		return fmt.Errorf("failed to configure containerd: %w", err)
	}
	if err := setup.KubeletWorker(snap, s.Name(), nodeIP, response.ClusterDNS, response.ClusterDomain, response.CloudProvider); err != nil {
		return fmt.Errorf("failed to configure kubelet: %w", err)
	}
	if err := setup.KubeProxy(s.Context, snap, s.Name(), response.PodCIDR); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}
	if err := setup.K8sAPIServerProxy(snap, response.APIServers); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}

	// TODO(berkayoz): remove the lock on cleanup
	if err := snaputil.MarkAsWorkerNode(snap, true); err != nil {
		return fmt.Errorf("failed to mark node as worker: %w", err)
	}

	// Start services
	if err := snaputil.StartWorkerServices(s.Context, snap); err != nil {
		return fmt.Errorf("failed to start worker services: %w", err)
	}

	return nil
}

func (a *App) onBootstrapControlPlane(s *state.State, initConfig map[string]string) error {
	snap := a.Snap()

	bootstrapConfig, err := apiv1.BootstrapConfigFromMicrocluster(initConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bootstrap config: %w", err)
	}

	cfg, err := types.ClusterConfigFromBootstrapConfig(bootstrapConfig)
	if err != nil {
		return fmt.Errorf("invalid bootstrap config: %w", err)
	}
	cfg.SetDefaults()
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid cluster configuration: %w", err)
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
			Hostname:          s.Name(),
			IPSANs:            []net.IP{{127, 0, 0, 1}},
			Years:             20,
			AllowSelfSignedCA: true,
		})
		if err := certificates.CompleteCertificates(); err != nil {
			return fmt.Errorf("failed to initialize k8s-dqlite certificates: %w", err)
		}
		if err := setup.EnsureK8sDqlitePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write k8s-dqlite certificates: %w", err)
		}

		cfg.Datastore.K8sDqliteCert = vals.Pointer(certificates.K8sDqliteCert)
		cfg.Datastore.K8sDqliteKey = vals.Pointer(certificates.K8sDqliteKey)
	case "external":
		certificates := &pki.ExternalDatastorePKI{
			DatastoreCACert:     cfg.Datastore.GetExternalCACert(),
			DatastoreClientCert: cfg.Datastore.GetExternalClientCert(),
			DatastoreClientKey:  cfg.Datastore.GetExternalClientKey(),
		}
		if err := certificates.CheckCertificates(); err != nil {
			return fmt.Errorf("failed to initialize external datastore certificates: %w", err)
		}
		if err := setup.EnsureExtDatastorePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write external datastore certificates: %w", err)
		}
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.Datastore.GetType(), setup.SupportedDatastores)
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
	if err := setup.EnsureControlPlanePKI(snap, certificates); err != nil {
		return fmt.Errorf("failed to write control plane certificates: %w", err)
	}

	// Add certificates to the cluster config
	cfg.Certificates.CACert = vals.Pointer(certificates.CACert)
	cfg.Certificates.CAKey = vals.Pointer(certificates.CAKey)
	cfg.Certificates.FrontProxyCACert = vals.Pointer(certificates.FrontProxyCACert)
	cfg.Certificates.FrontProxyCAKey = vals.Pointer(certificates.FrontProxyCAKey)
	cfg.Certificates.APIServerKubeletClientCert = vals.Pointer(certificates.APIServerKubeletClientCert)
	cfg.Certificates.APIServerKubeletClientKey = vals.Pointer(certificates.APIServerKubeletClientKey)
	cfg.Certificates.ServiceAccountKey = vals.Pointer(certificates.ServiceAccountKey)

	// Generate kubeconfigs
	if err := setupKubeconfigs(s, snap.KubernetesConfigDir(), cfg.APIServer.GetSecurePort(), cfg.Certificates.GetCACert()); err != nil {
		return fmt.Errorf("failed to generate kubeconfigs: %w", err)
	}

	// Configure datastore
	switch cfg.Datastore.GetType() {
	case "k8s-dqlite":
		if err := setup.K8sDqlite(snap, fmt.Sprintf("%s:%d", nodeIP.String(), cfg.Datastore.GetK8sDqlitePort()), nil); err != nil {
			return fmt.Errorf("failed to configure k8s-dqlite: %w", err)
		}
	case "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.Datastore.GetType(), setup.SupportedDatastores)
	}

	// Configure services
	if err := setupControlPlaneServices(snap, s, cfg, nodeIP); err != nil {
		return fmt.Errorf("failed to configure services: %w", err)
	}

	// Write cluster configuration to dqlite
	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := database.SetClusterConfig(ctx, tx, cfg); err != nil {
			return fmt.Errorf("failed to write cluster configuration: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database transaction to update cluster configuration failed: %w", err)
	}

	// Start services
	if err := startControlPlaneServices(s.Context, snap, cfg.Datastore.GetType()); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait until Kube-API server is ready
	if err := waitApiServerReady(s.Context, snap); err != nil {
		return fmt.Errorf("kube-apiserver did not become ready in time: %w", err)
	}

	// TODO(neoaggelos): the work below should be a POST /cluster/config

	if cfg.Network.GetEnabled() {
		if err := component.ReconcileNetworkComponent(s.Context, snap, vals.Pointer(false), vals.Pointer(true), cfg); err != nil {
			return fmt.Errorf("failed to reconcile network: %w", err)
		}
	}

	if cfg.DNS.GetEnabled() {
		dnsIP, _, err := component.ReconcileDNSComponent(s.Context, snap, vals.Pointer(false), vals.Pointer(true), cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile dns: %w", err)
		}

		// If DNS IP is not empty, update cluster configuration
		if dnsIP != "" {
			if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
				if cfg, err = database.SetClusterConfig(ctx, tx, types.ClusterConfig{
					Kubelet: types.Kubelet{
						ClusterDNS: vals.Pointer(dnsIP),
					},
				}); err != nil {
					return fmt.Errorf("failed to update cluster configuration for dns=%s: %w", dnsIP, err)
				}
				return nil
			}); err != nil {
				return fmt.Errorf("database transaction to update cluster configuration failed: %w", err)
			}
		}
	}

	cmData, err := cfg.Kubelet.ToConfigMap()
	if err != nil {
		return fmt.Errorf("failed to format kubelet configmap data: %w", err)
	}

	client, err := k8s.NewClient(snap.KubernetesRESTClientGetter(""))
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if _, err := client.UpdateConfigMap(s.Context, "kube-system", "k8sd-config", cmData); err != nil {
		return fmt.Errorf("failed to update node configs: %w", err)
	}

	if cfg.LocalStorage.GetEnabled() {
		if err := component.ReconcileLocalStorageComponent(s.Context, snap, vals.Pointer(false), vals.Pointer(true), cfg); err != nil {
			return fmt.Errorf("failed to reconcile local-storage: %w", err)
		}
	}

	if cfg.Gateway.GetEnabled() {
		if err := component.ReconcileGatewayComponent(s.Context, snap, vals.Pointer(false), vals.Pointer(true), cfg); err != nil {
			return fmt.Errorf("failed to reconcile gateway: %w", err)
		}
	}

	if cfg.Ingress.GetEnabled() {
		if err := component.ReconcileIngressComponent(s.Context, snap, vals.Pointer(false), vals.Pointer(true), cfg); err != nil {
			return fmt.Errorf("failed to reconcile ingress: %w", err)
		}
	}

	if cfg.LoadBalancer.GetEnabled() {
		if err := component.ReconcileLoadBalancerComponent(s.Context, snap, vals.Pointer(false), vals.Pointer(true), cfg); err != nil {
			return fmt.Errorf("failed to reconcile load-balancer: %w", err)
		}
	}

	if cfg.MetricsServer.GetEnabled() {
		if err := component.ReconcileMetricsServerComponent(s.Context, snap, vals.Pointer(false), vals.Pointer(true), cfg); err != nil {
			return fmt.Errorf("failed to reconcile metrics-server: %w", err)
		}
	}

	return nil
}
