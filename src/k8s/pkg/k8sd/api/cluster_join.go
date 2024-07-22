package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) postClusterJoin(s state.State, r *http.Request) response.Response {
	req := apiv1.JoinClusterRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	hostname, err := utils.CleanHostname(req.Name)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid hostname %q: %w", req.Name, err))
	}

	if _, err := e.provider.MicroCluster().Status(r.Context()); err == nil {
		return NodeInUse(fmt.Errorf("node %q is part of the cluster", hostname))
	}

	config := map[string]string{}

	// NOTE(neoaggelos): microcluster adds an implicit 30 second timeout if no context deadline is set.
	ctx, cancel := context.WithTimeout(r.Context(), time.Hour)
	defer cancel()

	// NOTE(neoaggelos): pass the timeout as a config option, so that the context cancel will propagate errors.
	config = utils.MicroclusterConfigWithTimeout(config, req.Timeout)

	internalToken := types.InternalWorkerNodeToken{}
	// Check if token is worker token
	if internalToken.Decode(req.Token) == nil {
		// valid worker node token - let's join the cluster
		// The validation of the token is done when fetching the cluster information.
		config["workerToken"] = req.Token
		config["workerJoinConfig"] = req.Config
		if err := e.provider.MicroCluster().NewCluster(ctx, hostname, req.Address, config); err != nil {
			return response.InternalError(fmt.Errorf("failed to join k8sd cluster as worker: %w", err))
		}
	} else {
		// Is not a worker token. let microcluster check if it is a valid control-plane token.
		config["controlPlaneJoinConfig"] = req.Config
		if err := e.provider.MicroCluster().JoinCluster(ctx, hostname, req.Address, req.Token, config); err != nil {
			return response.InternalError(fmt.Errorf("failed to join k8sd cluster as control plane: %w", err))
		}
	}

	return response.SyncResponse(true, nil)
}
