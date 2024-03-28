package api

import (
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) getNodeStatus(s *state.State, r *http.Request) response.Response {
	status, err := impl.GetLocalNodeStatus(r.Context(), s, e.provider.Snap())
	if err != nil {
		response.InternalError(err)
	}

	result := apiv1.GetNodeStatusResponse{
		NodeStatus: status,
	}

	return response.SyncResponse(true, &result)
}
