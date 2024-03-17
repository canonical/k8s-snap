package client

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

// NodeStatus queries the local node status
func (c *k8sdClient) LocalNodeStatus(ctx context.Context) (apiv1.NodeStatus, error) {
	var response apiv1.GetNodeStatusResponse

	if err := c.mc.Query(ctx, "GET", api.NewURL().Path("k8sd", "node"), nil, &response); err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("failed to GET /k8sd/node: %w", err)
	}
	return response.NodeStatus, nil
}
