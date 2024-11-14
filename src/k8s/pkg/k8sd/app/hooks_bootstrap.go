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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	"github.com/canonical/microcluster/v2/state"
)

// onBootstrap is called after we bootstrap the first cluster node.
// onBootstrap configures local services then writes the cluster config on the database.
func (a *App) onBootstrap(ctx context.Context, s state.State, initConfig map[string]string) error {
	// NOTE(neoaggelos): context timeout is passed over configuration, so that hook failures are propagated to the client
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if t := utils.MicroclusterTimeoutFromMap(initConfig); t != 0 {
		ctx, cancel = context.WithTimeout(ctx, t)
		defer cancel()
	}

	if workerToken, ok := initConfig["workerToken"]; ok {
		workerConfig, err := utils.MicroclusterWorkerJoinConfigFromMap(initConfig)
		if err != nil {
			return fmt.Errorf("failed to unmarshal worker join config: %w", err)
		}
		return a.onBootstrapWorkerNode(ctx, s, workerToken, workerConfig)
	}

	bootstrapConfig, err := utils.MicroclusterBootstrapConfigFromMap(initConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bootstrap config: %w", err)
	}

	return a.onBootstrapControlPlane(ctx, s, bootstrapConfig)
}

func (a *App) onBootstrapWorkerNode(ctx context.Context, s state.State, encodedToken string, joinConfig apiv1.WorkerJoinConfig) (rerr error) {
	snap := a.Snap()

	log := log.FromContext(ctx).WithValues("hook", "join")

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

	// Get remote certificate from the cluster member. We only need one node to be reachable for this.
	// One might fail because the node is not part of the cluster anymore but was at the time the token was created.
	var cert *x509.Certificate
	var address string
	var err error
	for _, address = range token.JoinAddresses {
		cert, err = utils.GetRemoteCertificate(address)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("failed to get certificate of cluster member: %w", err)
	}

	// verify that the fingerprint of the certificate matches the fingerprint of the token
	fingerprint := utils.CertFingerprint(cert)
	if fingerprint != token.Fingerprint {
		return fmt.Errorf("fingerprint from token (%q) does not match fingerprint of node %q (%q)", token.Fingerprint, address, fingerprint)
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
		Error    string                          `json:"error"`
		Metadata apiv1.GetWorkerJoinInfoResponse `json:"metadata"`
	}

	requestBody, err := json.Marshal(apiv1.GetWorkerJoinInfoRequest{Address: nodeIP.String()})
	if err != nil {
		return fmt.Errorf("failed to prepare worker info request: %w", err)
	}

	httpRequest, err := http.NewRequest("POST", fmt.Sprintf("https://%s/1.0/k8sd/worker/info", address), bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to prepare HTTP request: %w", err)
	}
	httpRequest.Header.Add("Worker-Name", s.Name())
	httpRequest.Header.Add("Worker-Token", token.Secret)

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
		CACert:              response.CACert,
		ClientCACert:        response.ClientCACert,
		KubeletCert:         response.KubeletCert,
		KubeletKey:          response.KubeletKey,
		KubeletClientCert:   response.KubeletClientCert,
		KubeletClientKey:    response.KubeletClientKey,
		KubeProxyClientCert: response.KubeProxyClientCert,
		KubeProxyClientKey:  response.KubeProxyClientKey,
	}

	// override certificates from JoinConfig
	for _, i := range []struct {
		target   *string
		override string
	}{
		{target: &certificates.KubeletCert, override: joinConfig.GetKubeletCert()},
		{target: &certificates.KubeletKey, override: joinConfig.GetKubeletKey()},
		{target: &certificates.KubeletClientCert, override: joinConfig.GetKubeletClientCert()},
		{target: &certificates.KubeletClientKey, override: joinConfig.GetKubeletClientKey()},
		{target: &certificates.KubeProxyClientCert, override: joinConfig.GetKubeProxyClientCert()},
		{target: &certificates.KubeProxyClientKey, override: joinConfig.GetKubeProxyClientKey()},
	} {
		if i.override != "" {
			*i.target = i.override
		}
	}

	if err := certificates.CompleteCertificates(); err != nil {
		return fmt.Errorf("failed to initialize worker node certificates: %w", err)
	}

	if _, err := setup.EnsureWorkerPKI(snap, certificates); err != nil {
		return fmt.Errorf("failed to write worker node certificates: %w", err)
	}

	var localhostAddress string
	if nodeIP.To4() == nil {
		localhostAddress = "[::1]"
	} else {
		localhostAddress = "127.0.0.1"
	}

	port := "6443"
	if len(response.APIServers) == 0 {
		return fmt.Errorf("no APIServers found in worker node info")
	}
	// Get the secure port from the first APIServer since they should all be the same.
	port = response.APIServers[0][strings.LastIndex(response.APIServers[0], ":")+1:]
	securePort, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("failed to parse apiserver secure port: %w", err)
	}

	// Kubeconfigs
	if err := setup.Kubeconfig(filepath.Join(snap.KubernetesConfigDir(), "kubelet.conf"), fmt.Sprintf("%s:%d", localhostAddress, securePort), certificates.CACert, certificates.KubeletClientCert, certificates.KubeletClientKey); err != nil {
		return fmt.Errorf("failed to generate kubelet kubeconfig: %w", err)
	}
	if err := setup.Kubeconfig(filepath.Join(snap.KubernetesConfigDir(), "proxy.conf"), fmt.Sprintf("%s:%d", localhostAddress, securePort), certificates.CACert, certificates.KubeProxyClientCert, certificates.KubeProxyClientKey); err != nil {
		return fmt.Errorf("failed to generate kube-proxy kubeconfig: %w", err)
	}

	// Write worker node configuration to dqlite
	//
	// Worker nodes only use a subset of the ClusterConfig struct. At the moment, these are:
	// - Network.PodCIDR and Network.ClusterCIDR: informative
	// - Certificates.K8sdPublicKey: used to verify the signature of the k8sd-config configmap.
	// - Certificates.CACert: kubernetes CA certificate.
	// - Certificates.ClientCACert: kubernetes client CA certificate.
	//
	// TODO(neoaggelos): We should be explicit here and try to avoid having worker nodes use
	// or set other cluster configuration keys by accident.
	cfg := types.ClusterConfig{
		APIServer: types.APIServer{
			SecurePort: utils.Pointer(securePort),
		},
		Network: types.Network{
			PodCIDR:     utils.Pointer(response.PodCIDR),
			ServiceCIDR: utils.Pointer(response.ServiceCIDR),
		},
		Certificates: types.Certificates{
			K8sdPublicKey: utils.Pointer(response.K8sdPublicKey),
			CACert:        utils.Pointer(response.CACert),
			ClientCACert:  utils.Pointer(response.ClientCACert),
		},
		Annotations: response.Annotations,
	}

	serviceConfigs := types.K8sServiceConfigs{
		ExtraNodeKubeletArgs:   joinConfig.ExtraNodeKubeletArgs,
		ExtraNodeKubeProxyArgs: joinConfig.ExtraNodeKubeProxyArgs,
	}

	// Pre-init checks
	if err := snap.PreInitChecks(ctx, cfg, serviceConfigs, false); err != nil {
		return fmt.Errorf("pre-init checks failed for worker node: %w", err)
	}

	if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := database.SetClusterConfig(ctx, tx, cfg); err != nil {
			return fmt.Errorf("failed to write cluster configuration: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database transaction to set cluster configuration failed: %w", err)
	}

	// Worker node services
	if err := setup.Containerd(snap, joinConfig.ExtraNodeContainerdConfig, joinConfig.ExtraNodeContainerdArgs); err != nil {
		return fmt.Errorf("failed to configure containerd: %w", err)
	}
	if err := setup.KubeletWorker(snap, s.Name(), nodeIP, response.ClusterDNS, response.ClusterDomain, response.CloudProvider, joinConfig.ExtraNodeKubeletArgs); err != nil {
		return fmt.Errorf("failed to configure kubelet: %w", err)
	}
	if err := setup.KubeProxy(ctx, snap, s.Name(), response.PodCIDR, localhostAddress, joinConfig.ExtraNodeKubeProxyArgs); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}
	if err := setup.K8sAPIServerProxy(snap, response.APIServers, securePort, joinConfig.ExtraNodeK8sAPIServerProxyArgs); err != nil {
		return fmt.Errorf("failed to configure k8s-apiserver-proxy: %w", err)
	}
	if err := setup.ExtraNodeConfigFiles(snap, joinConfig.ExtraNodeConfigFiles); err != nil {
		return fmt.Errorf("failed to write extra node config files: %w", err)
	}

	if err := snaputil.MarkAsWorkerNode(snap, true); err != nil {
		return fmt.Errorf("failed to mark node as worker: %w", err)
	}

	// Start services
	// This may fail if the node controllers try to restart the services at the same time, hence the retry.
	log.Info("Starting worker services")
	if err := control.RetryFor(ctx, 5, 5*time.Second, func() error {
		if err := snaputil.StartWorkerServices(ctx, snap); err != nil {
			return fmt.Errorf("failed to start worker services: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed after retry: %w", err)
	}

	return nil
}

func (a *App) onBootstrapControlPlane(ctx context.Context, s state.State, bootstrapConfig apiv1.BootstrapConfig) (rerr error) {
	snap := a.Snap()

	log := log.FromContext(ctx).WithValues("hook", "bootstrap")

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

	var localhostAddress string
	if nodeIP.To4() == nil {
		localhostAddress = "[::1]"
	} else {
		localhostAddress = "127.0.0.1"
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

	// NOTE: Set the notBefore certificate time to the current time.
	notBefore := time.Now()

	switch cfg.Datastore.GetType() {
	case "k8s-dqlite":
		// NOTE: Default certificate expiration is set to 20 years.
		certificates := pki.NewK8sDqlitePKI(pki.K8sDqlitePKIOpts{
			Hostname:          s.Name(),
			IPSANs:            []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
			NotBefore:         notBefore,
			NotAfter:          notBefore.AddDate(20, 0, 0),
			AllowSelfSignedCA: true,
		})
		if err := certificates.CompleteCertificates(); err != nil {
			return fmt.Errorf("failed to initialize k8s-dqlite certificates: %w", err)
		}
		if _, err := setup.EnsureK8sDqlitePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write k8s-dqlite certificates: %w", err)
		}

		cfg.Datastore.K8sDqliteCert = utils.Pointer(certificates.K8sDqliteCert)
		cfg.Datastore.K8sDqliteKey = utils.Pointer(certificates.K8sDqliteKey)
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
	// NOTE: Default certificate expiration is set to 20 years.
	extraIPs, extraNames := utils.SplitIPAndDNSSANs(bootstrapConfig.ExtraSANs)
	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:                  s.Name(),
		IPSANs:                    append(append([]net.IP{nodeIP}, serviceIPs...), extraIPs...),
		DNSSANs:                   extraNames,
		NotBefore:                 notBefore,
		NotAfter:                  notBefore.AddDate(20, 0, 0),
		AllowSelfSignedCA:         true,
		IncludeMachineAddressSANs: true,
	})

	certificates.CACert = bootstrapConfig.GetCACert()
	certificates.CAKey = bootstrapConfig.GetCAKey()
	certificates.ClientCACert = bootstrapConfig.GetClientCACert()
	certificates.ClientCAKey = bootstrapConfig.GetClientCAKey()
	certificates.FrontProxyCACert = bootstrapConfig.GetFrontProxyCACert()
	certificates.FrontProxyCAKey = bootstrapConfig.GetFrontProxyCAKey()
	certificates.FrontProxyClientCert = bootstrapConfig.GetFrontProxyClientCert()
	certificates.FrontProxyClientKey = bootstrapConfig.GetFrontProxyClientKey()
	certificates.ServiceAccountKey = bootstrapConfig.GetServiceAccountKey()
	certificates.APIServerKubeletClientCert = bootstrapConfig.GetAPIServerKubeletClientCert()
	certificates.APIServerKubeletClientKey = bootstrapConfig.GetAPIServerKubeletClientKey()
	certificates.AdminClientCert = bootstrapConfig.GetAdminClientCert()
	certificates.AdminClientKey = bootstrapConfig.GetAdminClientKey()
	certificates.KubeControllerManagerClientCert = bootstrapConfig.GetKubeControllerManagerClientCert()
	certificates.KubeControllerManagerClientKey = bootstrapConfig.GetKubeControllerManagerClientKey()
	certificates.KubeSchedulerClientCert = bootstrapConfig.GetKubeSchedulerClientCert()
	certificates.KubeSchedulerClientKey = bootstrapConfig.GetKubeSchedulerClientKey()
	certificates.KubeProxyClientCert = bootstrapConfig.GetKubeProxyClientCert()
	certificates.KubeProxyClientKey = bootstrapConfig.GetKubeProxyClientKey()

	certificates.APIServerCert = bootstrapConfig.GetAPIServerCert()
	certificates.APIServerKey = bootstrapConfig.GetAPIServerKey()
	certificates.KubeletCert = bootstrapConfig.GetKubeletCert()
	certificates.KubeletKey = bootstrapConfig.GetKubeletKey()
	certificates.KubeletClientCert = bootstrapConfig.GetKubeletClientCert()
	certificates.KubeletClientKey = bootstrapConfig.GetKubeletClientKey()

	if err := certificates.CompleteCertificates(); err != nil {
		return fmt.Errorf("failed to initialize control plane certificates: %w", err)
	}

	if _, err := setup.EnsureControlPlanePKI(snap, certificates); err != nil {
		return fmt.Errorf("failed to write control plane certificates: %w", err)
	}

	// Add certificates to the cluster config
	cfg.Certificates.CACert = utils.Pointer(certificates.CACert)
	cfg.Certificates.CAKey = utils.Pointer(certificates.CAKey)
	cfg.Certificates.ClientCACert = utils.Pointer(certificates.ClientCACert)
	cfg.Certificates.ClientCAKey = utils.Pointer(certificates.ClientCAKey)
	cfg.Certificates.FrontProxyCACert = utils.Pointer(certificates.FrontProxyCACert)
	cfg.Certificates.FrontProxyCAKey = utils.Pointer(certificates.FrontProxyCAKey)
	cfg.Certificates.APIServerKubeletClientCert = utils.Pointer(certificates.APIServerKubeletClientCert)
	cfg.Certificates.APIServerKubeletClientKey = utils.Pointer(certificates.APIServerKubeletClientKey)
	cfg.Certificates.ServiceAccountKey = utils.Pointer(certificates.ServiceAccountKey)
	cfg.Certificates.AdminClientCert = utils.Pointer(certificates.AdminClientCert)
	cfg.Certificates.AdminClientKey = utils.Pointer(certificates.AdminClientKey)
	cfg.Certificates.K8sdPublicKey = utils.Pointer(certificates.K8sdPublicKey)
	cfg.Certificates.K8sdPrivateKey = utils.Pointer(certificates.K8sdPrivateKey)

	serviceConfigs := types.K8sServiceConfigs{
		ExtraNodeKubeSchedulerArgs:         bootstrapConfig.ExtraNodeKubeSchedulerArgs,
		ExtraNodeKubeControllerManagerArgs: bootstrapConfig.ExtraNodeKubeControllerManagerArgs,
		ExtraNodeKubeletArgs:               bootstrapConfig.ExtraNodeKubeletArgs,
		ExtraNodeKubeProxyArgs:             bootstrapConfig.ExtraNodeKubeProxyArgs,
	}

	// Pre-init checks
	if err := snap.PreInitChecks(ctx, cfg, serviceConfigs, true); err != nil {
		return fmt.Errorf("pre-init checks failed for bootstrap node: %w", err)
	}

	// Generate kubeconfigs
	if err := setup.SetupControlPlaneKubeconfigs(snap.KubernetesConfigDir(), localhostAddress, cfg.APIServer.GetSecurePort(), *certificates); err != nil {
		return fmt.Errorf("failed to generate kubeconfigs: %w", err)
	}

	// Configure datastore
	switch cfg.Datastore.GetType() {
	case "k8s-dqlite":
		address := fmt.Sprintf("%s:%d", utils.ToIPString(nodeIP), cfg.Datastore.GetK8sDqlitePort())
		if err := setup.K8sDqlite(snap, address, nil, bootstrapConfig.ExtraNodeK8sDqliteArgs); err != nil {
			return fmt.Errorf("failed to configure k8s-dqlite: %w", err)
		}
	case "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.Datastore.GetType(), setup.SupportedDatastores)
	}

	// Configure services
	if err := setup.Containerd(snap, bootstrapConfig.ExtraNodeContainerdConfig, bootstrapConfig.ExtraNodeContainerdArgs); err != nil {
		return fmt.Errorf("failed to configure containerd: %w", err)
	}
	if err := setup.KubeletControlPlane(snap, s.Name(), nodeIP, cfg.Kubelet.GetClusterDNS(), cfg.Kubelet.GetClusterDomain(), cfg.Kubelet.GetCloudProvider(), cfg.Kubelet.GetControlPlaneTaints(), bootstrapConfig.ExtraNodeKubeletArgs); err != nil {
		return fmt.Errorf("failed to configure kubelet: %w", err)
	}
	if err := setup.KubeProxy(ctx, snap, s.Name(), cfg.Network.GetPodCIDR(), localhostAddress, bootstrapConfig.ExtraNodeKubeProxyArgs); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}
	if err := setup.KubeControllerManager(snap, bootstrapConfig.ExtraNodeKubeControllerManagerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-controller-manager: %w", err)
	}
	if err := setup.KubeScheduler(snap, bootstrapConfig.ExtraNodeKubeSchedulerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-scheduler: %w", err)
	}
	if err := setup.KubeAPIServer(snap, cfg.APIServer.GetSecurePort(), nodeIP, cfg.Network.GetServiceCIDR(), s.Address().Path("1.0", "kubernetes", "auth", "webhook").String(), true, cfg.Datastore, cfg.APIServer.GetAuthorizationMode(), bootstrapConfig.ExtraNodeKubeAPIServerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-apiserver: %w", err)
	}

	if err := setup.ExtraNodeConfigFiles(snap, bootstrapConfig.ExtraNodeConfigFiles); err != nil {
		return fmt.Errorf("failed to write extra node config files: %w", err)
	}

	// Write cluster configuration to dqlite
	if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := database.SetClusterConfig(ctx, tx, cfg); err != nil {
			return fmt.Errorf("failed to write cluster configuration: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database transaction to update cluster configuration failed: %w", err)
	}

	if err := snapdconfig.SetSnapdFromK8sd(ctx, cfg.ToUserFacing(), snap); err != nil {
		return fmt.Errorf("failed to set snapd configuration from k8sd: %w", err)
	}

	// Start services
	// This may fail if the node controllers try to restart the services at the same time, hence the retry.
	log.Info("Starting control-plane services")
	if err := control.RetryFor(ctx, 5, 5*time.Second, func() error {
		if err := startControlPlaneServices(ctx, snap, cfg.Datastore.GetType()); err != nil {
			return fmt.Errorf("failed to start services: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed after retry: %w", err)
	}

	// Wait until Kube-API server is ready
	log.Info("Waiting for kube-apiserver to become ready")
	if err := waitApiServerReady(ctx, snap); err != nil {
		return fmt.Errorf("kube-apiserver did not become ready in time: %w", err)
	}
	log.Info("API server is ready - notify controllers")

	a.NotifyFeatureController(
		cfg.Network.GetEnabled(),
		cfg.Gateway.GetEnabled(),
		cfg.Ingress.GetEnabled(),
		cfg.LoadBalancer.GetEnabled(),
		cfg.LocalStorage.GetEnabled(),
		cfg.MetricsServer.GetEnabled(),
		cfg.DNS.GetEnabled(),
	)
	a.NotifyUpdateNodeConfigController()
	return nil
}
