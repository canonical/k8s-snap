package client

import (
	"context"
	"fmt"
	"time"
)

func (c *Client) JoinNode(ctx context.Context, name string, address string, token string) error {
	if err := c.m.Ready(30); err != nil {
		return fmt.Errorf("cluster did not come up in time: %w", err)
	}
	if err := c.m.JoinCluster(name, address, token, nil, time.Second*180); err != nil {
		return fmt.Errorf("failed to join k8sd cluster: %w", err)
	}
	return nil
}

func (c *Client) RemoveNode(ctx context.Context, name string, force bool) error {
	return c.mc.DeleteClusterMember(ctx, name, force)
}
