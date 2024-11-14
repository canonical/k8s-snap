package api

import (
	"net/http"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) getNodeStatus(s state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()

	status, err := impl.GetLocalNodeStatus(r.Context(), s, snap)
	if err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, &apiv1.NodeStatusResponse{
		NodeStatus: status,
	})
}
