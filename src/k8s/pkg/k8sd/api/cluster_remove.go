package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
	nodeutil "github.com/canonical/k8s/pkg/utils/node"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/cluster"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) postClusterRemove(s *state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()

	req := apiv1.RemoveNodeRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	if req.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	isControlPlane, err := nodeutil.IsControlPlaneNode(ctx, s, req.Name)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is control-plane: %w", err))
	}
	if isControlPlane {
		log.Printf("Waiting for node to not be pending")
		control.WaitUntilReady(ctx, func() (bool, error) {
			var notPending bool
			if err := s.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				member, err := cluster.GetInternalClusterMember(ctx, tx, req.Name)
				if err != nil {
					log.Printf("Failed to get member: %v", err)
					return nil
				}
				log.Printf("Node %s is %s", member.Name, member.Role)
				notPending = member.Role != cluster.Pending
				return nil
			}); err != nil {
				log.Printf("Transaction to check cluster member role failed: %v", err)
			}
			return notPending, nil
		})
		log.Printf("Starting node deletion")

		// Remove control plane via microcluster API.
		// The postRemove hook will take care of cleaning up kubernetes.
		c, err := s.Leader()
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to create client to cluster leader: %w", err))
		}
		if err := c.DeleteClusterMember(ctx, req.Name, req.Force); err != nil {
			return response.InternalError(fmt.Errorf("failed to delete cluster member %s: %w", req.Name, err))
		}
	}

	isWorker, err := databaseutil.IsWorkerNode(ctx, s, req.Name)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is worker: %w", err))
	}
	if isWorker {
		// For worker nodes, we need to manually clean up the kubernetes node and db entry.
		c, err := snap.KubernetesClient("")
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to create k8s client: %w", err))
		}

		if err := c.DeleteNode(ctx, req.Name); err != nil {
			return response.InternalError(fmt.Errorf("failed to remove k8s node %q: %w", req.Name, err))
		}

		if err := databaseutil.DeleteWorkerNodeEntry(ctx, s, req.Name); err != nil {
			return response.InternalError(fmt.Errorf("failed to remove worker entry %q: %w", req.Name, err))
		}
	}

	if !isWorker && !isControlPlane {
		return NodeUnavailable(fmt.Errorf("node %q is not part of the cluster", req.Name))
	}
	return response.SyncResponse(true, nil)
}
