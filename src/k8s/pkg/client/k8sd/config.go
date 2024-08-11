package k8sd

import (
	"context"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
)

func (c *k8sd) SetClusterConfig(ctx context.Context, request apiv1.SetClusterConfigRequest) error {
	_, err := query[any](ctx, c, "PUT", apiv1.SetClusterConfigRPC, request, nil)
	return err
}

func (c *k8sd) GetClusterConfig(ctx context.Context) (apiv1.GetClusterConfigResponse, error) {
	return query(ctx, c, "GET", apiv1.GetClusterConfigRPC, nil, &apiv1.GetClusterConfigResponse{})
}
