package api

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
	"gopkg.in/yaml.v2"
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
	// Check if the cluster is already bootstrapped
	status, err := e.provider.MicroCluster().Status(r.Context())
	if err != nil {
		return response.BadRequest(fmt.Errorf("failed to get microcluster status: %w", err))
	}

	if status.Ready {
		return NodeInUse(fmt.Errorf("node %q is part of the cluster", hostname))
	}

	joinConfig := struct {
		// We only care about this field from the entire join config.
		ContainerdBaseDir string `yaml:"containerd-base-dir,omitempty"`
	}{}

	if err := yaml.Unmarshal([]byte(req.Config), &joinConfig); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request config: %w", err))
	}

	if joinConfig.ContainerdBaseDir != "" {
		// append k8s-containerd to the given base dir, so we don't flood it with our own folders.
		e.provider.Snap().SetContainerdBaseDir(filepath.Join(joinConfig.ContainerdBaseDir, "k8s-containerd"))
	}

	config := map[string]string{}

	// NOTE(neoaggelos): microcluster adds an implicit 30 second timeout if no context deadline is set.
	ctx, cancel := context.WithTimeout(r.Context(), time.Hour)
	defer cancel()

	// NOTE(neoaggelos): pass the timeout as a config option, so that the context cancel will propagate errors.
	config = utils.MicroclusterMapWithTimeout(config, req.Timeout)

	internalToken := types.InternalWorkerNodeToken{}
	// Check if token is worker token
	if internalToken.Decode(req.Token) == nil {
		// valid worker node token - let's join the cluster
		// The validation of the token is done when fetching the cluster information.
		config = utils.MicroclusterMapWithWorkerJoinConfig(config, req.Token, req.Config)
		if err := e.provider.MicroCluster().NewCluster(ctx, hostname, req.Address, config); err != nil {
			return response.InternalError(fmt.Errorf("failed to join k8sd cluster as worker: %w", err))
		}
	} else {
		// Is not a worker token. let microcluster check if it is a valid control-plane token.
		config = utils.MicroclusterMapWithControlPlaneJoinConfig(config, req.Config)
		if err := e.provider.MicroCluster().JoinCluster(ctx, hostname, req.Address, req.Token, config); err != nil {
			return response.InternalError(fmt.Errorf("failed to join k8sd cluster as control plane: %w", err))
		}
	}

	return response.SyncResponse(true, &apiv1.JoinClusterResponse{})
}
