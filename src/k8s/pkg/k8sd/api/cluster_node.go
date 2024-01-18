package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/setup"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/cert"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/rest/types"
	"github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var k8sdClusterNode = rest.Endpoint{
	Path:   "k8sd/cluster/{node}",
	Post:   rest.EndpointAction{Handler: clusterNodePost, AllowUntrusted: false},
	Delete: rest.EndpointAction{Handler: clusterNodeDelete, AllowUntrusted: false},
}

func clusterNodePost(s *state.State, r *http.Request) response.Response {
	snap := snap.SnapFromContext(s.Context)

	var req apiv1.AddNodeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request data: %w", err))
	}

	// Retrieve the cluster configuration from the master node by querying the /k8sd/cluster/join endpoint.
	// The k8sd token is used to authenticate this request.
	clusterConfig, err := impl.GetClusterConfiguration(r.Context(), s, req.Token)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	k8sdToken, err := impl.K8sdTokenFromBase64Token(req.Token)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse token information: %w", err))
	}
	host, err := types.ParseAddrPort(req.Address)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse host address %s: %w", req.Address, err))
	}

	err = setup.InitFolders(snap.DataPath("args"))
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup folders: %w", err))
	}

	err = setup.InitServiceArgs(snap, clusterConfig.ExtraServiceArgs)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup service arguments: %w", err))
	}

	err = setup.InitContainerd(snap.Path("k8s/config/containerd/config.toml"), snap.Path("opt/cni/bin/"))
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to initialize containerd: %w", err))
	}

	// Get the CA certificate and key from the k8sd database and store it locally.
	if err := cert.StoreCertKeyPair(r.Context(), s, "ca", path.Join(cert.KubePkiPath, "ca.crt"), path.Join(cert.KubePkiPath, "ca.key")); err != nil {
		return response.SmartError(fmt.Errorf("failed to store CA certificate: %w", err))
	}

	// Use the CA from the cluster to sign the certificates
	ca, err := cert.LoadCertKeyPair(path.Join(cert.KubePkiPath, "ca.key"), path.Join(cert.KubePkiPath, "ca.crt"))
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to read CA: %w", err))
	}
	certMan, err := setup.InitCertificates(ca)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup certificates: %w", err))
	}

	apiServerPort, err := strconv.Atoi(clusterConfig.ExtraServiceArgs["kube-apiserver"]["--secure-port"])
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse apiserver port override: %w", err))
	}

	apiServerIp, _, err := net.SplitHostPort(k8sdToken.JoinAddresses[0])
	if err != nil {
		return response.SmartError(fmt.Errorf(
			"failed to parse IP from join address %s: %w", k8sdToken.JoinAddresses[0], err))
	}
	err = setup.InitKubeconfigs(r.Context(), s, certMan.CA, &apiServerIp, &apiServerPort)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to generate kubeconfig files: %w", err))
	}

	err = setup.InitKubeApiserver(snap.Path("k8s/config/apiserver-token-hook.tmpl"))
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to initialize kube-apiserver: %w", err))
	}

	err = setup.InitPermissions(r.Context(), snap)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup permissions: %w", err))
	}

	err = impl.JoinK8sDqliteCluster(r.Context(), s, snap, k8sdToken.JoinAddresses, host.Addr().String())
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to join k8s-dqlite nodes: %w", err))
	}

	err = snap.StartService(r.Context(), "k8s")
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to start services: %w", err))
	}

	result := apiv1.AddNodeResponse{}
	return response.SyncResponse(true, &result)
}

func clusterNodeDelete(s *state.State, r *http.Request) response.Response {
	// Get node name from URL.
	nodeName, err := url.PathUnescape(mux.Vars(r)["node"])
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse node name from URL '%s': %w", r.URL, err))
	}

	var req apiv1.RemoveNodeRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse request data: %w", err))
	}

	logrus.WithField("name", nodeName).Info("Delete cluster member")
	err = impl.DeleteClusterMember(r.Context(), s, nodeName, req.Force)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to delete cluster member: %w", err))
	}
	result := apiv1.AddNodeResponse{}
	return response.SyncResponse(true, &result)
}
