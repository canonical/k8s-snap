package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/state"
)

func postClusterRemoveNode(m *microcluster.MicroCluster, s *state.State, r *http.Request) response.Response {
	snap := snap.SnapFromContext(s.Context)

	req := apiv1.RemoveNodeRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	isControlPlane, err := utils.IsControlPlaneNode(r.Context(), s, req.Name)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is control-plane: %w", err))
	}
	if isControlPlane {
		// Remove control plane via microcluster API.
		// The postRemove hook will take care of cleaning up kubernetes.
		c, err := m.LocalClient()
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to create local client: %w", err))
		}
		if err := c.DeleteClusterMember(r.Context(), req.Name, req.Force); err != nil {
			return response.InternalError(fmt.Errorf("failed to delete cluster member %s: %w", req.Name, err))
		}
	}

	isWorker, err := utils.IsWorkerNode(r.Context(), s, req.Name)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is control-plane: %w", err))
	}
	if isWorker {
		// For worker nodes, we need to manually cleanup the kubernetes node and db entry.
		c, err := k8s.NewClient(snap)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to create k8s client: %w", err))
		}

		if err := c.DeleteNode(s.Context, req.Name); err != nil {
			return response.InternalError(fmt.Errorf("failed to remove k8s node %q: %w", req.Name, err))
		}

		if err := utils.DeleteWorkerNodeEntry(r.Context(), s, req.Name); err != nil {
			return response.InternalError(fmt.Errorf("failed to remove worker entry %q: %w", req.Name, err))
		}
	}

	if !isWorker && !isControlPlane {
		return response.InternalError(fmt.Errorf("Node %q is not part of the cluster", req.Name))
	}
	return response.SyncResponse(true, nil)
}
