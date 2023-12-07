package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
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
	// Get node name from URL.
	nodeName, err := url.PathUnescape(mux.Vars(r)["node"])
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse node name from URL '%s': %w", r.URL, err))
	}

	var req api.AddNodeRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(err)
	}
	logrus.Info(req)
	logrus.Info(nodeName)
	// TODO: Implement k8s joining stuff (e.g. get kubelet args etc.)

	result := api.AddNodeResponse{}
	return response.SyncResponse(true, &result)
}

func clusterNodeDelete(s *state.State, r *http.Request) response.Response {
	// Get node name from URL.
	nodeName, err := url.PathUnescape(mux.Vars(r)["node"])
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse node name from URL '%s': %w", r.URL, err))
	}

	var req api.RemoveNodeRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse request data: %w", err))
	}

	logrus.WithField("name", nodeName).Info("Delete cluster member")
	err = utils.DeleteClusterMember(r.Context(), s, nodeName, req.Force)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to delete cluster member: %w", err))
	}
	result := api.AddNodeResponse{}
	return response.SyncResponse(true, &result)
}
