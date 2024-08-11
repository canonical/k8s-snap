package k8sd

import (
	"context"
	"fmt"
	"strings"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/lxd/shared/api"
)

func (c *k8sd) SetClusterConfig(ctx context.Context, request apiv1.SetClusterConfigRequest) error {
	if err := c.client.Query(ctx, "PUT", apiv1.K8sdAPIVersion, api.NewURL().Path(strings.Split(apiv1.SetClusterConfigRPC, "/")...), request, nil); err != nil {
		return fmt.Errorf("failed to PUT /k8sd/cluster/config: %w", err)
	}
	return nil
}

func (c *k8sd) GetClusterConfig(ctx context.Context) (apiv1.UserFacingClusterConfig, error) {
	var response apiv1.GetClusterConfigResponse
	if err := c.client.Query(ctx, "GET", apiv1.K8sdAPIVersion, api.NewURL().Path(strings.Split(apiv1.GetClusterConfigRPC, "/")...), nil, &response); err != nil {
		return apiv1.UserFacingClusterConfig{}, fmt.Errorf("failed to GET /k8sd/cluster/config: %w", err)
	}

	return response.Config, nil
}
