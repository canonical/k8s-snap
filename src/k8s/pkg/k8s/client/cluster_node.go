package client

import (
	"context"
	"fmt"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

// JoinNode calls "POST 1.0/k8sd/cluster/<node>"
func (c *Client) JoinNode(ctx context.Context, name string, address string, token string) error {
	if err := c.m.Ready(30); err != nil {
		return fmt.Errorf("cluster did not come up in time: %w", err)
	}
	if err := c.m.JoinCluster(name, address, token, nil, time.Second*180); err != nil {
		return fmt.Errorf("failed to join k8sd cluster: %w", err)
	}
	return nil
}

// RemoveNode calls "DELETE 1.0/k8sd/cluster/<node>"
func (c *Client) RemoveNode(ctx context.Context, name string, force bool) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	request := apiv1.RemoveNodeRequest{
		Force: force,
	}
	var response apiv1.RemoveNodeResponse
	err := c.mc.Query(queryCtx, "DELETE", api.NewURL().Path("k8sd", "cluster", name), request, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return fmt.Errorf("failed to query DELETE k8sd/cluster/{name} endpoint on %q: %w", clientURL.String(), err)
	}
	return nil
}
