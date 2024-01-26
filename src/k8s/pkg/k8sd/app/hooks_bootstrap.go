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
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/cert"
	"github.com/canonical/microcluster/state"
)

// onBootstrap is called after we bootstrap the first cluster node.
// onBootstrap configures local services then writes the cluster config on the database.
func onBootstrap(s *state.State, initConfig map[string]string) error {
	if workerToken, ok := initConfig["workerToken"]; ok {
		return onBootstrapWorkerNode(s, workerToken)
	}

	return onBootstrapControlPlane(s)
}

func onBootstrapWorkerNode(s *state.State, encodedToken string) error {
	token := &types.InternalWorkerNodeToken{}
	if err := token.Decode(encodedToken); err != nil {
		return fmt.Errorf("failed to parse worker token: %w", err)
	}

	if len(token.JoinAddresses) == 0 {
		return fmt.Errorf("empty list of control plane addresses")
	}

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

	requestBody, err := json.Marshal(apiv1.WorkerNodeInfoRequest{Hostname: s.Name()})
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
	if err := setup.InitFolders(snap.DataPath("args")); err != nil {
		return fmt.Errorf("failed to setup folders: %w", err)
	}
	if err := setup.InitContainerd(snap.Path("k8s/config/containerd/config.toml"), snap.Path("opt/cni/bin/")); err != nil {
		return fmt.Errorf("failed to configure containerd: %w", err)
	}
	if err := setup.InitContainerdArgs(snap, nil, nil); err != nil {
		return fmt.Errorf("failed to configure containerd arguments: %w", err)
	}
	if err := setup.WriteCA(snap, response.CA); err != nil {
		return fmt.Errorf("failed to write CA certificate: %w", err)
	}

	kubeletArgs := map[string]string{
		"--hostname-override": s.Name(),
		"--cluster-dns":       response.ClusterDNS,
		"--cluster-domain":    response.ClusterDomain,
		"--cloud-provider":    response.CloudProvider,
	}
	if err := setup.InitKubeletArgs(snap, kubeletArgs, nil); err != nil {
		return fmt.Errorf("failed to configure kubelet: %w", err)
	}
	if err := setup.RenderKubeletKubeconfig(snap, response.KubeletToken, response.CA); err != nil {
		return fmt.Errorf("failed to render kubelet kubeconfig: %w", err)
	}

	proxyArgs := map[string]string{
		"--hostname-override": s.Name(),
		"--cluster-cidr":      response.ClusterCIDR,
	}
	if err := setup.InitKubeProxyArgs(snap, proxyArgs, nil); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}
	if err := setup.RenderKubeProxyKubeconfig(snap, response.KubeProxyToken, response.CA); err != nil {
		return fmt.Errorf("failed to render kube-proxy kubeconfig: %w", err)
	}

	if err := setup.InitAPIServerProxy(snap, response.APIServers); err != nil {
		return fmt.Errorf("failed to configure k8s-apiserver-proxy: %w", err)
	}

	// TODO: mark node as worker

	for _, service := range []string{"containerd", "k8s-apiserver-proxy", "kubelet", "kube-proxy"} {
		if err := snap.StartService(s.Context, fmt.Sprintf("k8s.%s", service)); err != nil {
			return fmt.Errorf("failed to start service %s: %w", service, err)
		}
	}

	return nil
}

func onBootstrapControlPlane(s *state.State) error {
	snap := snap.SnapFromContext(s.Context)

	err := setup.InitFolders(snap.DataPath("args"))
	if err != nil {
		return fmt.Errorf("failed to setup folders: %w", err)
	}

	err = setup.InitServiceArgs(snap, nil)
	if err != nil {
		return fmt.Errorf("failed to setup service arguments: %w", err)
	}

	err = setup.InitContainerd(snap.Path("k8s/config/containerd/config.toml"), snap.Path("opt/cni/bin/"))
	if err != nil {
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
		path.Join(cert.K8sDqlitePkiPath, "cluster.crt"), path.Join(cert.K8sDqlitePkiPath, "cluster.key"))
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

	// TODO(neoaggelos): first generate config then reconcile state
	s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		return database.SetClusterConfig(ctx, tx, database.ClusterConfig{
			Cluster: database.ClusterConfigCluster{
				CIDR: "10.1.0.0/16",
			},
			APIServer: database.ClusterConfigAPIServer{
				AuthorizationMode: "Node,RBAC",
				SecurePort:        6443,
			},
			Kubelet: database.ClusterConfigKubelet{
				ClusterDomain: "cluster.local",
			},
		})
	})

	err = snap.StartService(s.Context, "k8s")
	if err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	return nil
}
