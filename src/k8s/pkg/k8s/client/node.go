package client

import (
	"context"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

// NodeStatus queries the local node status
func (c *k8sdClient) NodeStatus(ctx context.Context) (apiv1.NodeStatus, error) {
	var response apiv1.GetNodeStatusResponse
	err := c.Query(ctx, "GET", api.NewURL().Path("k8sd", "node"), nil, &response)
	if err != nil {
		return apiv1.NodeStatus{}, err
	}
	return response.NodeStatus, nil
}
