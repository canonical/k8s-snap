package api

import (
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/utils"
	nodeutil "github.com/canonical/k8s/pkg/utils/node"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) postClusterRemove(s *state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()

	req := apiv1.RemoveNodeRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	isControlPlane, err := nodeutil.IsControlPlaneNode(r.Context(), s, req.Name)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is control-plane: %w", err))
	}
	if isControlPlane {
		// Remove control plane via microcluster API DeleteClusterMember which removes cluster member from dqlite.
		// DeleteClusterMember also calls ResetClusterMember which removes the cluster member from the microcluster state.
		// The preRemove (in ResetClusterMember) hook will take care of cleaning up kubernetes.
		c, err := e.provider.MicroCluster().LocalClient()
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to create local client: %w", err))
		}
		if err := c.DeleteClusterMember(r.Context(), req.Name, req.Force); err != nil {
			return response.InternalError(fmt.Errorf("failed to delete cluster member %s: %w", req.Name, err))
		}
	}

	isWorker, err := databaseutil.IsWorkerNode(r.Context(), s, req.Name)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is control-plane: %w", err))
	}
	if isWorker {
		// For worker nodes, we need to manually clean up the kubernetes node and db entry.
		c, err := snap.KubernetesClient("")
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to create k8s client: %w", err))
		}

		if err := c.DeleteNode(s.Context, req.Name); err != nil {
			return response.InternalError(fmt.Errorf("failed to remove k8s node %q: %w", req.Name, err))
		}

		if err := databaseutil.DeleteWorkerNodeEntry(r.Context(), s, req.Name); err != nil {
			return response.InternalError(fmt.Errorf("failed to remove worker entry %q: %w", req.Name, err))
		}
	}

	if !isWorker && !isControlPlane {
		return response.InternalError(fmt.Errorf("node %q is not part of the cluster", req.Name))
	}
	return response.SyncResponse(true, nil)
}
