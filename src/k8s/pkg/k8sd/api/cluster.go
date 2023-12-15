package api

import (
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
)

var k8sdCluster = rest.Endpoint{
	Path: "k8sd/cluster",
	Get:  rest.EndpointAction{Handler: clusterGet, AllowUntrusted: false},
	Post: rest.EndpointAction{Handler: clusterPost, AllowUntrusted: false},
}

func clusterGet(s *state.State, r *http.Request) response.Response {
	status, err := utils.GetClusterStatus(r.Context(), s)
	if err != nil {
		response.InternalError(err)
	}

	result := apiv1.GetClusterStatusResponse{
		ClusterStatus: status,
	}

	return response.SyncResponse(true, &result)
}

func clusterPost(s *state.State, r *http.Request) response.Response {
	// The `k8s init` command will be move here eventually - right now this only writes the k8s-dqlite
	// certificate to the cluster so that k8s-dqlite joining works.
	err := utils.WriteK8sDqliteCertInfoToK8sd(r.Context(), s)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to write k8s-dqlite cert to k8sd: %w", err))
	}
	return response.SyncResponse(true, &apiv1.InitClusterResponse{})
}
