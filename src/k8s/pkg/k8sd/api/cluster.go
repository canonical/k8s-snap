package api

import (
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
)

var k8sdCluster = rest.Endpoint{
	Path: "k8sd/cluster",
	Get:  rest.EndpointAction{Handler: clusterGet, AllowUntrusted: false},
}

func clusterGet(s *state.State, r *http.Request) response.Response {
	status, err := impl.GetClusterStatus(r.Context(), s)
	if err != nil {
		response.InternalError(err)
	}

	result := apiv1.GetClusterStatusResponse{
		ClusterStatus: status,
	}

	return response.SyncResponse(true, &result)
}
