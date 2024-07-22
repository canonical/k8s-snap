package k8sd

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

func (c *k8sd) SetClusterConfig(ctx context.Context, request apiv1.UpdateClusterConfigRequest) error {
	if err := c.client.Query(ctx, "PUT", apiv1.K8sdAPIVersion, api.NewURL().Path("k8sd", "cluster", "config"), request, nil); err != nil {
		return fmt.Errorf("failed to PUT /k8sd/cluster/config: %w", err)
	}
	return nil
}

func (c *k8sd) GetClusterConfig(ctx context.Context) (apiv1.UserFacingClusterConfig, error) {
	var response apiv1.GetClusterConfigResponse
	if err := c.client.Query(ctx, "GET", apiv1.K8sdAPIVersion, api.NewURL().Path("k8sd", "cluster", "config"), nil, &response); err != nil {
		return apiv1.UserFacingClusterConfig{}, fmt.Errorf("failed to GET /k8sd/cluster/config: %w", err)
	}

	return response.Config, nil
}
