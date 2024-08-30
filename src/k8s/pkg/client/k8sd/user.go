package k8sd

import (
	"context"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
)

func (c *k8sd) KubeConfig(ctx context.Context, request apiv1.KubeConfigRequest) (apiv1.KubeConfigResponse, error) {
	return query(ctx, c, "GET", apiv1.KubeConfigRPC, request, &apiv1.KubeConfigResponse{})
}
