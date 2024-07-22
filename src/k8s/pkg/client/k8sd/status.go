package k8sd

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/lxd/shared/api"
)

func (c *k8sd) NodeStatus(ctx context.Context) (apiv1.NodeStatus, error) {
	var response apiv1.GetNodeStatusResponse
	if err := c.client.Query(ctx, "GET", apiv1.K8sdVersionPrefix, api.NewURL().Path("k8sd", "node"), nil, &response); err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("failed to GET /k8sd/node: %w", err)
	}
	return response.NodeStatus, nil
}

func (c *k8sd) ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error) {
	var response apiv1.GetClusterStatusResponse
	if err := control.WaitUntilReady(ctx, func() (bool, error) {
		if err := c.client.Query(ctx, "GET", apiv1.K8sdVersionPrefix, api.NewURL().Path("k8sd", "cluster"), nil, &response); err != nil {
			return false, fmt.Errorf("failed to GET /k8sd/cluster: %w", err)
		}
		return !waitReady || response.ClusterStatus.Ready, nil
	}); err != nil {
		return apiv1.ClusterStatus{}, err
	}
	return response.ClusterStatus, nil
}
