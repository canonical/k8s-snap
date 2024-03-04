package client

import (
	"context"
	"fmt"
	"time"

	api "github.com/canonical/k8s/api/v1"
	lxdApi "github.com/canonical/lxd/shared/api"
)

func (c *k8sdClient) UpdateClusterConfig(ctx context.Context, request api.UpdateClusterConfigRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateClusterConfigResponse
	err := c.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "cluster", "config"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to update cluster configuration: %w", err)
	}
	return nil
}

func (c *k8sdClient) GetClusterConfig(ctx context.Context, request api.GetClusterConfigRequest) (api.UserFacingClusterConfig, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.GetClusterConfigResponse

	err := c.Query(queryCtx, "GET", lxdApi.NewURL().Path("k8sd", "cluster", "config"), nil, &response)
	if err != nil {
		return api.UserFacingClusterConfig{}, err
	}

	return response.Config, nil
}
