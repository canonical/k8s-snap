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
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/k8s/pkg/utils/vals"
	"github.com/canonical/microcluster/state"
)

// onBootstrap is called after we bootstrap the first cluster node.
// onBootstrap configures local services then writes the cluster config on the database.
func onBootstrap(s *state.State, initConfig map[string]string) error {
	if workerToken, ok := initConfig["workerToken"]; ok {
		return onBootstrapWorkerNode(s, workerToken)
	}

	return onBootstrapControlPlane(s, initConfig)
}

func onBootstrapWorkerNode(s *state.State, encodedToken string) error {
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

	snap := snap.SnapFromContext(s.Context)

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
		return fmt.Errorf("failed to initialize cluster certificates: %w", err)
	}
	if err := setup.EnsureWorkerPKI(snap, certificates); err != nil {
		return fmt.Errorf("failed to write cluster certificates: %w", err)
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

func onBootstrapControlPlane(s *state.State, initConfig map[string]string) error {
	snap := snap.SnapFromContext(s.Context)

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
	if err := setup.EnsureAllDirectories(snap); err != nil {
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
			return fmt.Errorf("failed to initialize cluster certificates: %w", err)
		}
		if err := setup.EnsureK8sDqlitePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write cluster certificates: %w", err)
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
			return fmt.Errorf("failed to initialize cluster certificates: %w", err)
		}
		if err := setup.EnsureExtDatastorePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write cluster certificates: %w", err)
		}
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.APIServer.Datastore, setup.SupportedDatastores)
	}

	// Certificates
	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:                  s.Name(),
		IPSANs:                    append([]net.IP{nodeIP}, serviceIPs...),
		Years:                     20,
		AllowSelfSignedCA:         true,
		IncludeMachineAddressSANs: true,
	})
	if err := certificates.CompleteCertificates(); err != nil {
		return fmt.Errorf("failed to initialize cluster certificates: %w", err)
	}
	if err := setup.EnsureControlPlanePKI(snap, certificates); err != nil {
		return fmt.Errorf("failed to write cluster certificates: %w", err)
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
		if err := setup.K8sDqlite(snap, fmt.Sprintf("%s:%d", nodeIP.String(), cfg.K8sDqlite.Port), nil); err != nil {
			return fmt.Errorf("failed to configure k8s-dqlite: %w", err)
		}
	case "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.APIServer.Datastore, setup.SupportedDatastores)
	}

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
	switch cfg.APIServer.Datastore {
	case "k8s-dqlite":
		if err := snaputil.StartK8sDqliteServices(s.Context, snap); err != nil {
			return fmt.Errorf("failed to start control plane services: %w", err)
		}
	case "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.APIServer.Datastore, setup.SupportedDatastores)
	}

	if err := snaputil.StartControlPlaneServices(s.Context, snap); err != nil {
		return fmt.Errorf("failed to start control plane services: %w", err)
	}

	// Wait for API server to come up
	client, err := k8s.NewClient(snap)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	err = client.WaitApiServerReady(s.Context)
	if err != nil {
		return fmt.Errorf("k8s api server did not become ready in time: %w", err)
	}

	if cfg.Network.Enabled != nil {
		err := component.ReconcileNetworkComponent(s.Context, snap, vals.Pointer(false), cfg.Network.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile network: %w", err)
		}
	}

	if cfg.DNS.Enabled != nil {
		dnsIP, _, err := component.ReconcileDNSComponent(s.Context, snap, vals.Pointer(false), cfg.DNS.Enabled, cfg)
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
	}

	if cfg.LocalStorage.Enabled != nil {
		err := component.ReconcileLocalStorageComponent(s.Context, snap, vals.Pointer(false), cfg.LocalStorage.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile local-storage: %w", err)
		}
	}

	if cfg.Gateway.Enabled != nil {
		err := component.ReconcileGatewayComponent(s.Context, snap, vals.Pointer(false), cfg.Gateway.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile gateway: %w", err)
		}
	}

	if cfg.Ingress.Enabled != nil {
		err := component.ReconcileIngressComponent(s.Context, snap, vals.Pointer(false), cfg.Ingress.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile ingress: %w", err)
		}
	}

	if cfg.LoadBalancer.Enabled != nil {
		err := component.ReconcileLoadBalancerComponent(s.Context, snap, vals.Pointer(false), cfg.LoadBalancer.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile load-balancer: %w", err)
		}
	}

	if cfg.MetricsServer.Enabled != nil {
		err := component.ReconcileMetricsServerComponent(s.Context, snap, vals.Pointer(false), cfg.MetricsServer.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile metrics-server: %w", err)
		}
	}

	return nil
}
