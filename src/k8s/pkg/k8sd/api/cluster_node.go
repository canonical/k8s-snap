package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/setup"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
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
	var req apiv1.AddNodeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request data: %w", err))
	}

	k8sdToken, err := impl.K8sdTokenFromBase64Token(req.Token)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse token information: %w", err))
	}
	host, err := types.ParseAddrPort(req.Address)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse host address %s: %w", req.Address, err))
	}

	err = setup.InitFolders()
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup folders: %w", err))
	}

	err = setup.InitServiceArgs()
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup service arguments: %w", err))
	}

	err = setup.InitContainerd()
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to initialize containerd: %w", err))
	}

	err = setup.InitPermissions(r.Context())
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup permissions: %w", err))
	}

	err = impl.JoinK8sDqliteCluster(r.Context(), s, k8sdToken.JoinAddresses, host.Addr().String())
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to join k8s-dqlite nodes: %w", err))
	}

	// TODO: Implement k8s joining stuff (e.g. get kubelet args etc.)

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
