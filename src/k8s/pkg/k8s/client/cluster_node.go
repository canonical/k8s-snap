package client

import (
	"context"
	"fmt"
	"os"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/lxd/shared/api"
)

func (c *k8sdClient) JoinCluster(ctx context.Context, request apiv1.JoinClusterRequest) error {
	timeout := utils.TimeoutFromCtx(ctx)

	if err := c.m.Ready(int(timeout / time.Second)); err != nil {
		return fmt.Errorf("k8sd API is not ready: %w", err)
	}

	if err := c.mc.Query(ctx, "POST", api.NewURL().Path("k8sd", "cluster", "join"), request, nil); err != nil {
		// TODO(neoaggelos): only return error that join cluster failed
		fmt.Fprintln(os.Stderr, "Cleaning up, error was", err)
		c.CleanupNode(ctx, request.Name)
		return fmt.Errorf("failed to POST /k8sd/cluster/join: %w", err)
	}

	c.WaitForDqliteNodeToBeReady(ctx, request.Name)
	return nil
}

func (c *k8sdClient) RemoveNode(ctx context.Context, request apiv1.RemoveNodeRequest) error {
	if err := c.mc.Query(ctx, "POST", api.NewURL().Path("k8sd", "cluster", "remove"), request, nil); err != nil {
		return fmt.Errorf("failed to DELETE /k8sd/cluster/remove: %w", err)
	}
	return nil
}

func (c *k8sdClient) ResetNode(ctx context.Context, name string, force bool) error {
	if err := c.mc.ResetClusterMember(ctx, name, force); err != nil {
		return fmt.Errorf("failed to ResetClusterMember: %w", err)
	}
	return nil
}

// WaitForDqliteNodeToBeReady waits until the underlying dqlite node of the microcluster is not in PENDING state.
// While microcluster checkReady will validate that the nodes API server is ready, it will not check if the
// dqlite node is properly setup yet.
func (c *k8sdClient) WaitForDqliteNodeToBeReady(ctx context.Context, nodeName string) error {
	return control.WaitUntilReady(ctx, func() (bool, error) {
		clusterStatus, err := c.ClusterStatus(ctx, false)
		if err != nil {
			return false, fmt.Errorf("failed to get the cluster status: %w", err)
		}

		for _, member := range clusterStatus.Members {
			if member.Name == nodeName {
				if member.DatastoreRole == apiv1.DatastoreRolePending {
					return false, nil
				}
				return true, nil
			}
		}
		return false, fmt.Errorf("cluster does not contain node %s", nodeName)
	})
}

// CleanupNode resets the nodes configuration and cluster state.
// The cleanup will happen on a best-effort base. Any error that occurs will be ignored.
func (c *k8sdClient) CleanupNode(ctx context.Context, nodeName string) {

	// For self-removal, microcluster expects the dqlite node to not be in pending state.
	c.WaitForDqliteNodeToBeReady(ctx, nodeName)

	// Delete the node from the cluster.
	// This will fail if this is the only member in the cluster.
	c.RemoveNode(ctx, apiv1.RemoveNodeRequest{Name: nodeName, Force: false})
	// Reset the local state and daemon.
	// This is required to reset a bootstrapped node before
	// joining another cluster.
	c.ResetNode(ctx, nodeName, true)

	snaputil.StopControlPlaneServices(ctx, c.snap)
	snaputil.StopK8sDqliteServices(ctx, c.snap)

	snaputil.MarkAsWorkerNode(c.snap, false)
}
