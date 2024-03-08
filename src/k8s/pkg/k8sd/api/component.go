package api

import (
	"fmt"
	"net/http"

	api "github.com/canonical/k8s/api/v1"

	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func getComponents(s *state.State, r *http.Request) response.Response {
	snap := snap.SnapFromContext(s.Context)

	components, err := impl.GetComponents(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get components: %w", err))
	}

	result := api.GetComponentsResponse{
		Components: components,
	}
	return response.SyncResponse(true, &result)
}
