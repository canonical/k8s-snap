package client

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
)

func (c *Client) joinWorkerNode(ctx context.Context, name, address, token string) error {
	return c.m.NewCluster(name, address, map[string]string{"workerToken": token}, time.Second*180)
}

func (c *Client) joinControlPlaneNode(ctx context.Context, name, address, token string) error {
	return c.m.JoinCluster(name, address, token, nil, time.Second*180)
}

func (c *Client) JoinNode(ctx context.Context, name string, address string, token string) error {
	if err := c.m.Ready(30); err != nil {
		return fmt.Errorf("cluster did not come up in time: %w", err)
	}

	// differentiate between control plane and worker node tokens
	info := &types.InternalWorkerNodeToken{}
	if info.Decode(token) == nil {
		// valid worker node token
		if err := c.joinWorkerNode(ctx, name, address, token); err != nil {

			// TODO: Handle worker node specific cleanup.
			// If the node setup unrecoverably fails after the worker has
			// registered itself to the cluster, the worker needs to remove itself again.
			// For that:
			//  - we need an endpoint on the control-plane with which workers can remove themselves.
			//  - we need unique worker tokens (right now, all workers share the same one) so that
			//    each worker kann only remove itself and not other workers.
			return fmt.Errorf("failed to join k8sd cluster as worker: %w", err)
		}
	} else {
		if err := c.joinControlPlaneNode(ctx, name, address, token); err != nil {
			c.CleanupNode(ctx, name)
			return fmt.Errorf("failed to join k8sd cluster as control plane: %w", err)
		}
	}

	c.WaitForDqliteNodeToBeReady(ctx, name)
	return nil
}

func (c *Client) RemoveNode(ctx context.Context, name string, force bool) error {
	return c.mc.DeleteClusterMember(ctx, name, force)
}

func (c *Client) ResetNode(ctx context.Context, name string, force bool) error {
	return c.mc.ResetClusterMember(ctx, name, force)
}

// WaitForDqliteNodeToBeReady waits until the underlying dqlite node of the microcluster is not in PENDING state.
// While microcluster checkReady will validate that the nodes API server is ready, it will not check if the
// dqlite node is properly setup yet.
func (c *Client) WaitForDqliteNodeToBeReady(ctx context.Context, nodeName string) error {
	return control.WaitUntilReady(ctx, func() (bool, error) {
		clusterStatus, err := c.ClusterStatus(ctx, false)
		if err != nil {
			return false, fmt.Errorf("failed to get cluster status: %w", err)
		}

		for _, member := range clusterStatus.Members {
			if member.Name == nodeName {
				if member.Role == "PENDING" {
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
func (c *Client) CleanupNode(ctx context.Context, nodeName string) {

	// For self-removal, microcluster expects the dqlite node to not be in pending state.
	c.WaitForDqliteNodeToBeReady(ctx, nodeName)

	// Delete the node from the cluster.
	// This will fail if this is the only member in the cluster.
	c.RemoveNode(ctx, nodeName, false)
	// Reset the local state and daemon.
	// This is required to reset a bootstrapped node before
	// joining another cluster.
	c.ResetNode(ctx, nodeName, true)

	snap.StopControlPlaneServices(ctx, snap.NewDefaultSnap())
}
