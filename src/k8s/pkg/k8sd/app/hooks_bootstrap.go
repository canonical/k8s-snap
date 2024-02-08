package app

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/k8s"
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

	requestBody, err := json.Marshal(apiv1.WorkerNodeInfoRequest{Hostname: s.Name(), Address: nodeIP.String()})
	if err != nil {
		return fmt.Errorf("failed to prepare worker info request: %w", err)
	}

	httpRequest, err := http.NewRequest("POST", fmt.Sprintf("https://%s/1.0/k8sd/worker/info", token.JoinAddresses[0]), bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to prepare HTTP request: %w", err)
	}
	httpRequest.Header.Add("k8sd-token", token.Token)

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

	certificates := &pki.WorkerNodePKI{
		CACert:      response.CA,
		KubeletCert: response.KubeletCert,
		KubeletKey:  response.KubeletKey,
	}
	for action, f := range map[string]func() error{
		"ensure cluster certificates": func() error { return certificates.CompleteCertificates() },
		"create cluster directories":  func() error { return setup.EnsureAllDirectories(snap) },
		"write cluster certificates":  func() error { return setup.EnsureWorkerPKI(snap, certificates) },
		"generate kubelet kubeconfig": func() error {
			return setup.Kubeconfig(path.Join(snap.KubernetesConfigDir(), "kubelet.conf"), response.KubeletToken, "127.0.0.1:6443", certificates.CACert)
		},
		"generate kube-proxy kubeconfig": func() error {
			return setup.Kubeconfig(path.Join(snap.KubernetesConfigDir(), "proxy.conf"), response.KubeProxyToken, "127.0.0.1:6443", certificates.CACert)
		},
		"configure containerd": func() error { return setup.Containerd(snap) },
		"configure kubelet": func() error {
			return setup.Kubelet(snap, s.Name(), nodeIP, response.ClusterDNS, response.ClusterDomain, response.CloudProvider)
		},
		"configure kube-proxy":          func() error { return setup.KubeProxy(snap, s.Name(), response.PodCIDR) },
		"configure k8s-apiserver-proxy": func() error { return setup.K8sAPIServerProxy(snap, response.APIServers) },
		"start worker node services":    func() error { return setup.KubeProxy(snap, s.Name(), response.PodCIDR) },
	} {
		if err := f(); err != nil {
			return fmt.Errorf("failed to %s: %w", action, err)
		}
	}

	return nil
}

func onBootstrapControlPlane(s *state.State, initConfig map[string]string) error {
	snap := snap.SnapFromContext(s.Context)

	bootstrapConfig, err := apiv1.BootstrapConfigFromMap(initConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bootstrap config: %w", err)
	}
	cfg, err := types.MergeClusterConfig(types.DefaultClusterConfig(), types.ClusterConfigFromBootstrapConfig(bootstrapConfig))
	if err != nil {
		return fmt.Errorf("failed initialize cluster config from bootstrap config: %w", err)
	}
	nodeIP := net.ParseIP(s.Address().Hostname())
	if nodeIP == nil {
		return fmt.Errorf("failed to parse node IP address %q", s.Address().Hostname())
	}
	certificates := pki.NewControlPlanePKI(s.Name(), nil, []net.IP{nodeIP}, 10, true)

	type step struct {
		name string
		f    func() error
	}

	for _, step := range []step{
		{"initialize cluster certificates", func() error { return certificates.CompleteCertificates() }},
		{"create cluster directories", func() error { return setup.EnsureAllDirectories(snap) }},
		{"write cluster certificates", func() error { return setup.EnsureControlPlanePKI(snap, certificates) }},
	} {
		if err := step.f(); err != nil {
			return fmt.Errorf("failed to %s: %w", step.name, err)
		}
	}

	cfg.Certificates.CACert = certificates.CACert
	cfg.Certificates.CAKey = certificates.CAKey
	cfg.Certificates.FrontProxyCACert = certificates.FrontProxyCACert
	cfg.Certificates.FrontProxyCAKey = certificates.FrontProxyCAKey
	cfg.Certificates.APIServerKubeletClientCert = certificates.APIServerKubeletClientCert
	cfg.Certificates.APIServerKubeletClientKey = certificates.APIServerKubeletClientKey
	cfg.Certificates.K8sDqliteCert = certificates.K8sDqliteCert
	cfg.Certificates.K8sDqliteKey = certificates.K8sDqliteKey
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

	for _, step := range []step{
		{"configure containerd", func() error { return setup.Containerd(snap) }},
		{"configure k8s-dqlite", func() error {
			return setup.K8sDqlite(snap, fmt.Sprintf("%s:%d", nodeIP.String(), cfg.K8sDqlite.Port), nil)
		}},
		{"configure kubelet", func() error {
			return setup.Kubelet(snap, s.Name(), nodeIP, cfg.Kubelet.ClusterDNS, cfg.Kubelet.ClusterDomain, cfg.Kubelet.CloudProvider)
		}},
		{"configure kube-proxy", func() error { return setup.KubeProxy(snap, s.Name(), cfg.Network.PodCIDR) }},
		{"configure kube-controller-manager", func() error { return setup.KubeControllerManager(snap) }},
		{"configure kube-scheduler", func() error { return setup.KubeScheduler(snap) }},
		{"configure kube-apiserver", func() error {
			return setup.KubeAPIServer(snap, cfg.Network.ServiceCIDR, s.Address().Path("1.0", "kubernetes", "auth", "webhook").String(), true, cfg.APIServer.Datastore, cfg.APIServer.AuthorizationMode)
		}},
		{"start control plane services", func() error { return snaputil.StartControlPlaneServices(s.Context, snap) }},
	} {
		if err := step.f(); err != nil {
			return fmt.Errorf("failed to %s: %w", step.name, err)
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
