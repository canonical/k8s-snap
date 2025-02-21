package api

import (
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v3/state"
)

func (e *Endpoints) postSnapRefresh(s state.State, r *http.Request) response.Response {
	req := apiv1.SnapRefreshRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	refreshOpts, err := types.RefreshOptsFromAPI(req)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid refresh options: %w", err))
	}

	id, err := e.provider.Snap().Refresh(e.Context(), refreshOpts)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to refresh snap: %w", err))
	}

	return response.SyncResponse(true, apiv1.SnapRefreshResponse{ChangeID: id})
}
