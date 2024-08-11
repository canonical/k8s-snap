package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) postClusterBootstrap(_ state.State, r *http.Request) response.Response {
	req := apiv1.BootstrapClusterRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	// Convert Bootstrap config to map
	config, err := utils.MicroclusterMapWithBootstrapConfig(nil, req.Config)
	if err != nil {
		return response.BadRequest(fmt.Errorf("failed to prepare bootstrap config: %w", err))
	}

	// Clean hostname
	hostname, err := utils.CleanHostname(req.Name)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid hostname %q: %w", req.Name, err))
	}

	// Check if the cluster is already bootstrapped
	_, err = e.provider.MicroCluster().Status(r.Context())
	if err == nil {
		return response.BadRequest(fmt.Errorf("cluster is already bootstrapped"))
	}

	// NOTE(neoaggelos): microcluster adds an implicit 30 second timeout if no context deadline is set.
	ctx, cancel := context.WithTimeout(r.Context(), time.Hour)
	defer cancel()

	// NOTE(neoaggelos): pass the timeout as a config option, so that the context cancel will propagate errors.
	config = utils.MicroclusterMapWithTimeout(config, req.Timeout)

	// Bootstrap the cluster
	if err := e.provider.MicroCluster().NewCluster(ctx, hostname, req.Address, config); err != nil {
		return response.BadRequest(fmt.Errorf("failed to bootstrap new cluster: %w", err))
	}

	return response.SyncResponse(true, &apiv1.BootstrapClusterResponse{
		Name:    hostname,
		Address: req.Address,
	})
}
