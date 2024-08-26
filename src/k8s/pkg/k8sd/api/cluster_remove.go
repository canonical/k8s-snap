package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
	nodeutil "github.com/canonical/k8s/pkg/utils/node"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v3/cluster"
	"github.com/canonical/microcluster/v3/state"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (e *Endpoints) postClusterRemove(s state.State, r *http.Request) response.Response {
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

	log := log.FromContext(ctx).WithValues("name", req.Name)

	isControlPlane, err := nodeutil.IsControlPlaneNode(ctx, s, req.Name)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is control-plane: %w", err))
	}
	if isControlPlane {
		log.Info("Waiting for node to not be pending")
		control.WaitUntilReady(ctx, func() (bool, error) {
			var notPending bool
			if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				member, err := cluster.GetCoreClusterMember(ctx, tx, req.Name)
				if err != nil {
					log.Error(err, "Failed to get member")
					return nil
				}
				log.WithValues("role", member.Role).Info("Current node role")
				notPending = member.Role != cluster.Pending
				return nil
			}); err != nil {
				log.Error(err, "Transaction to check cluster member role failed")
			}
			return notPending, nil
		})

		log.Info("Starting node deletion")

		// Remove control plane via microcluster API.
		// The postRemove hook will take care of cleaning up kubernetes.
		c, err := s.Leader()
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to create client to cluster leader: %w", err))
		}
		if err := c.DeleteClusterMember(ctx, req.Name, req.Force); err != nil {
			return response.InternalError(fmt.Errorf("failed to delete cluster member %s: %w", req.Name, err))
		}

		return response.SyncResponse(true, &apiv1.RemoveNodeResponse{})
	}

	cfg, err := databaseutil.GetClusterConfig(ctx, s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	if _, ok := cfg.Annotations[apiv1.AnnotationSkipCleanupKubernetesNodeOnRemove]; ok {
		// Explicitly skip removing the node from Kubernetes.
		log.Info("Skipping Kubernetes worker node removal")
		return response.SyncResponse(true, nil)
	}

	client, err := snap.KubernetesClient("")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create k8s client: %w", err))
	}
	if node, err := client.CoreV1().Nodes().Get(ctx, req.Name, metav1.GetOptions{}); err != nil {
		return NodeUnavailable(fmt.Errorf("node %q is not part of the cluster: %w", req.Name, err))
	} else if v, ok := node.Labels["k8sd.io/role"]; !ok || v != "worker" {
		return NodeUnavailable(fmt.Errorf("node %q is missing k8sd.io/role=worker label", req.Name))
	}

	if err := client.DeleteNode(ctx, req.Name); err != nil {
		return response.InternalError(fmt.Errorf("failed to remove k8s node %q: %w", req.Name, err))
	}

	return response.SyncResponse(true, &apiv1.RemoveNodeResponse{})
}
