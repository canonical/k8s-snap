package api

import (
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	microclusterutil "github.com/canonical/k8s/pkg/utils/microcluster"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) postFeaturesReconcile(s state.State, r *http.Request) response.Response {
	req := apiv1.ReconcileFeaturesRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	isLeader, err := microclusterutil.IsLeader(s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is leader: %w", err))
	}

	if !isLeader {
		return response.InternalError(fmt.Errorf("feature reconciliation is not allowed on non-leader nodes"))
	}

	e.provider.NotifyFeatureController(
		req.Network,
		req.Gateway,
		req.Ingress,
		req.LoadBalancer,
		req.LocalStorage,
		req.MetricsServer,
		req.DNS,
	)

	return response.SyncResponse(true, &apiv1.ReconcileFeaturesResponse{})
}
