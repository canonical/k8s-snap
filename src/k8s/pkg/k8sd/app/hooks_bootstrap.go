package app

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/setup"
	"github.com/canonical/k8s/pkg/k8sd/database/clusterconfigs"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/cert"
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

func onBootstrapWorkerNode(state *state.State, encodedToken string) error {
	token := &types.InternalWorkerNodeToken{}
	if err := token.Decode(encodedToken); err != nil {
		return fmt.Errorf("failed to parse worker token: %w", err)
	}

	if len(token.JoinAddresses) == 0 {
		return fmt.Errorf("empty list of control plane addresses")
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

	requestBody, err := json.Marshal(apiv1.WorkerNodeInfoRequest{Hostname: state.Name()})
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

	s := snap.SnapFromContext(state.Context)
	if err := setup.InitFolders(s.DataPath("args")); err != nil {
		return fmt.Errorf("failed to setup folders: %w", err)
	}
	if err := setup.InitContainerd(s); err != nil {
		return fmt.Errorf("failed to configure containerd: %w", err)
	}
	if err := setup.InitContainerdArgs(s, nil, nil); err != nil {
		return fmt.Errorf("failed to configure containerd arguments: %w", err)
	}
	if err := setup.WriteCA(s, response.CA); err != nil {
		return fmt.Errorf("failed to write CA certificate: %w", err)
	}

	kubeletArgs := map[string]string{
		"--hostname-override": state.Name(),
		"--cluster-dns":       response.ClusterDNS,
		"--cluster-domain":    response.ClusterDomain,
		"--cloud-provider":    response.CloudProvider,
	}
	if err := setup.InitKubeletArgs(s, kubeletArgs, nil); err != nil {
		return fmt.Errorf("failed to configure kubelet: %w", err)
	}
	if err := setup.RenderKubeletKubeconfig(s, response.KubeletToken, response.CA); err != nil {
		return fmt.Errorf("failed to render kubelet kubeconfig: %w", err)
	}

	proxyArgs := map[string]string{
		"--hostname-override": state.Name(),
		"--cluster-cidr":      response.ClusterCIDR,
	}
	if err := setup.InitKubeProxyArgs(s, proxyArgs, nil); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}
	if err := setup.RenderKubeProxyKubeconfig(s, response.KubeProxyToken, response.CA); err != nil {
		return fmt.Errorf("failed to render kube-proxy kubeconfig: %w", err)
	}

	if err := setup.InitAPIServerProxy(s, response.APIServers); err != nil {
		return fmt.Errorf("failed to configure k8s-apiserver-proxy: %w", err)
	}

	// TODO: mark node as worker

	if err := snap.StartWorkerServices(state.Context, s); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	return nil
}

func onBootstrapControlPlane(s *state.State, initConfig map[string]string) error {
	snap := snap.SnapFromContext(s.Context)

	err := setup.InitFolders(snap.DataPath("args"))
	if err != nil {
		return fmt.Errorf("failed to setup folders: %w", err)
	}

	err = setup.InitServiceArgs(snap, nil)
	if err != nil {
		return fmt.Errorf("failed to setup service arguments: %w", err)
	}

	if err := setup.InitContainerd(snap); err != nil {
		return fmt.Errorf("failed to initialize containerd: %w", err)
	}

	certMan, err := setup.InitCertificates(nil)
	if err != nil {
		return fmt.Errorf("failed to setup certificates: %w", err)
	}

	err = setup.InitKubeconfigs(s.Context, s, certMan.CA, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to kubeconfig files: %w", err)
	}

	err = setup.InitKubeApiserver(snap.Path("k8s/config/apiserver-token-hook.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to initialize kube-apiserver: %w", err)
	}

	err = setup.InitPermissions(s.Context, snap)
	if err != nil {
		return fmt.Errorf("failed to setup permissions: %w", err)
	}

	// TODO(neoaggelos): these should be done with "database.SetClusterConfig()" at the end of the bootstrap
	err = cert.WriteCertKeyPairToK8sd(s.Context, s, "certificates-k8s-dqlite",
		snap.CommonPath(cert.K8sDqlitePkiPath, "cluster.crt"), snap.CommonPath(cert.K8sDqlitePkiPath, "cluster.key"))
	if err != nil {
		return fmt.Errorf("failed to write k8s-dqlite cert to k8sd: %w", err)
	}
	err = cert.WriteCertKeyPairToK8sd(s.Context, s, "certificates-ca",
		path.Join(cert.KubePkiPath, "ca.crt"), path.Join(cert.KubePkiPath, "ca.key"))
	if err != nil {
		return fmt.Errorf("failed to write CA to k8sd: %w", err)
	}

	// TODO(neoaggelos): configure k8s-dqlite init.yaml file, as it is currently only left to guess for defaults
	//                   - see "k8s::init::k8s_dqlite" in k8s/lib.sh for details.
	//                   - do not bind on 127.0.0.1, use configuration option or fallback to default address like microcluster.

	clusterConfig := clusterconfigs.Default()
	bootstrapConfig, err := apiv1.BootstrapConfigFromMap(initConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bootstrap config: %w", err)
	}

	clusterConfig, err = clusterconfigs.Merge(clusterConfig, utils.ConvertBootstrapToClusterConfig(bootstrapConfig))
	if err != nil {
		return fmt.Errorf("failed to merge cluster config with bootstrap config: %w", err)
	}

	// TODO(neoaggelos): first generate config then reconcile state
	s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		return clusterconfigs.SetClusterConfig(ctx, tx, clusterConfig)
	})

	err = snap.StartService(s.Context, "k8s")
	if err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	k8sClient, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	// The apiserver needs to be ready to start components.
	err = k8s.WaitApiServerReady(s.Context, k8sClient)
	if err != nil {
		return fmt.Errorf("k8s api server did not become ready in time: %w", err)
	}

	// TODO: start configured components.
	return nil
}
