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
	err := c.m.JoinCluster(name, address, token, nil, time.Second*30)
	if err != nil {
		return fmt.Errorf("failed to join k8sd cluster: %w", err)
	}

	err = c.m.Ready(30)
	if err != nil {
		return fmt.Errorf("cluster did not come up in time: %w", err)
	}

	// Joining a node takes some time since services need to be restarted.
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*180)
	defer cancel()

	request := apiv1.AddNodeRequest{
		Address: address,
		Token:   token,
	}
	var response apiv1.AddNodeResponse
	err = c.mc.Query(queryCtx, "POST", api.NewURL().Path("k8sd", "cluster", name), request, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
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
		return fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}
	return nil
}
