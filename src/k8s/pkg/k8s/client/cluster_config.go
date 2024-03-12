package client

import (
	"context"
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	lxdApi "github.com/canonical/lxd/shared/api"
)

func (c *k8sdClient) UpdateClusterConfig(ctx context.Context, request api.UpdateClusterConfigRequest) error {
	var response api.UpdateClusterConfigResponse
	err := c.Query(ctx, "PUT", lxdApi.NewURL().Path("k8sd", "cluster", "config"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to update cluster configuration: %w", err)
	}
	return nil
}

func (c *k8sdClient) GetClusterConfig(ctx context.Context, request api.GetClusterConfigRequest) (api.UserFacingClusterConfig, error) {
	var response api.GetClusterConfigResponse

	err := c.Query(ctx, "GET", lxdApi.NewURL().Path("k8sd", "cluster", "config"), nil, &response)
	if err != nil {
		return api.UserFacingClusterConfig{}, err
	}

	return response.Config, nil
}
