package client

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
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
			return fmt.Errorf("failed to join k8sd cluster as worker: %w", err)
		}
	} else {
		if err := c.joinControlPlaneNode(ctx, name, address, token); err != nil {
			return fmt.Errorf("failed to join k8sd cluster as control plane: %w", err)
		}
	}
	return nil
}

func (c *Client) RemoveNode(ctx context.Context, name string, force bool) error {
	return c.mc.DeleteClusterMember(ctx, name, force)
}
