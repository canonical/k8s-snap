package k8sd

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

func (c *k8sd) KubeConfig(ctx context.Context, request apiv1.GetKubeConfigRequest) (string, error) {
	var response apiv1.GetKubeConfigResponse
	if err := c.client.Query(ctx, "GET", apiv1.K8sdAPIVersion, api.NewURL().Path("k8sd", "kubeconfig"), request, &response); err != nil {
		return "", fmt.Errorf("failed to GET /k8sd/kubeconfig: %w", err)
	}
	return response.KubeConfig, nil
}
