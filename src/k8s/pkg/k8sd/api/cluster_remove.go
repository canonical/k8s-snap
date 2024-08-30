package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

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
	log := log.FromContext(ctx).WithValues("name", req.Name)
	defer func() {
		select {
		case <-ctx.Done():
			log.Info("k8sd - context already cancelled, how?", "cause", context.Cause(ctx), "err", ctx.Err())
		default:
			log.Info("k8sd - running defer and ctx cancel")
		}
		cancel()
	}()

	if deadline, ok := ctx.Deadline(); ok {
		log.Info("k8sd - ctx has deadline", "deadline", deadline, "until", time.Until(deadline))
	}

	if req.Timeout > 0 {
		log.Info("k8sd - setting timeout for context", "timeout", req.Timeout)
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer func() {
			select {
			case <-ctx.Done():
				log.Info("k8sd - context already cancelled, how? -- timeout")
			default:
				log.Info("k8sd - running defer and ctx cancel -- timeout")
			}
			cancel()
		}()
	}

	if deadline, ok := ctx.Deadline(); ok {
		log.Info("k8sd - ctx has deadline 2", "deadline", deadline, "until", time.Until(deadline))
	}

	isControlPlane, err := nodeutil.IsControlPlaneNode(ctx, s, req.Name)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is control-plane: %w", err))
	}
	if isControlPlane {
		log.Info("k8sd - Waiting for node to not be pending")
		control.WaitUntilReady(ctx, func() (bool, error) {
			var notPending bool
			if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				member, err := cluster.GetCoreClusterMember(ctx, tx, req.Name)
				if err != nil {
					log.Error(err, "Failed to get member")
					return nil
				}
				log.WithValues("role", member.Role).Info("k8sd - Current node role")
				notPending = member.Role != cluster.Pending
				return nil
			}); err != nil {
				log.Error(err, "Transaction to check cluster member role failed")
			}
			return notPending, nil
		})

		log.Info("k8sd - Starting node deletion")

		// Remove control plane via microcluster API.
		// The postRemove hook will take care of cleaning up kubernetes.
		c, err := s.Leader()
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to create client to cluster leader: %w", err))
		}

		log.Info("k8sd - got the leader")

		// !!!IMPORTANT!!! this is the workaround, instead of passing `ctx`, we make another independant ctx
		// so that if the original `ctx` is canceled, `DeleteClusterMember` does not get canceled.
		// NOTE(hue): also use the https://github.com/HomayoonAlimohammadi/microcluster/tree/capi-test branch
		// in order to build the k8s snap with added microcluster logs.
		deleteCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer func() {
			select {
			case <-deleteCtx.Done():
				log.Info("k8sd - delete context already cancelled, how?", "cause", context.Cause(deleteCtx), "err", deleteCtx.Err())
			default:
				log.Info("k8sd - running defer and deleteCtx cancel")
			}
			cancel()
		}()
		if err := c.DeleteClusterMember(deleteCtx, req.Name, req.Force); err != nil {
			log.Info("k8sd - failed to delete cluster member", "err", err, "name", req.Name)
			log.Info("k8sd - checking context", "err", ctx.Err(), "cause", context.Cause(ctx))
			return response.InternalError(fmt.Errorf("failed to delete cluster member %s: %w", req.Name, err))
		}

		log.Info("k8sd - done deleting the cluster member", "err", err)

		return response.SyncResponse(true, &apiv1.RemoveNodeResponse{})
	}

	cfg, err := databaseutil.GetClusterConfig(ctx, s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	log.Info("k8sd - got cluster config")

	if _, ok := cfg.Annotations[apiv1.AnnotationSkipCleanupKubernetesNodeOnRemove]; ok {
		// Explicitly skip removing the node from Kubernetes.
		log.Info("k8sd - Skipping Kubernetes worker node removal")
		return response.SyncResponse(true, nil)
	}

	log.Info("k8sd - did not find SkipCleanup annotation")

	client, err := snap.KubernetesClient("")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create k8s client: %w", err))
	}

	log.Info("k8sd - got k8s clien")

	if node, err := client.CoreV1().Nodes().Get(ctx, req.Name, metav1.GetOptions{}); err != nil {
		return NodeUnavailable(fmt.Errorf("node %q is not part of the cluster: %w", req.Name, err))
	} else if v, ok := node.Labels["k8sd.io/role"]; !ok || v != "worker" {
		return NodeUnavailable(fmt.Errorf("node %q is missing k8sd.io/role=worker label", req.Name))
	}

	log.Info("k8sd - got node")

	if err := client.DeleteNode(ctx, req.Name); err != nil {
		return response.InternalError(fmt.Errorf("failed to remove k8s node %q: %w", req.Name, err))
	}

	log.Info("k8sd - removed node")

	return response.SyncResponse(true, &apiv1.RemoveNodeResponse{})
}
