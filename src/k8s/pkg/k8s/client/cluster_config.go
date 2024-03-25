package client

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	api "github.com/canonical/lxd/shared/api"
)

func (c *k8sdClient) UpdateClusterConfig(ctx context.Context, request apiv1.UpdateClusterConfigRequest) error {
	var response apiv1.UpdateClusterConfigResponse
	if err := c.mc.Query(ctx, "PUT", api.NewURL().Path("k8sd", "cluster", "config"), request, &response); err != nil {
		return fmt.Errorf("failed to PUT /k8sd/cluster/config: %w", err)
	}
	return nil
}

func (c *k8sdClient) GetClusterConfig(ctx context.Context, request apiv1.GetClusterConfigRequest) (apiv1.UserFacingClusterConfig, error) {
	var response apiv1.GetClusterConfigResponse

	if err := c.mc.Query(ctx, "GET", api.NewURL().Path("k8sd", "cluster", "config"), nil, &response); err != nil {
		return apiv1.UserFacingClusterConfig{}, fmt.Errorf("failed to GET /k8sd/cluster/config: %w", err)
	}

	return response.Config, nil
}
