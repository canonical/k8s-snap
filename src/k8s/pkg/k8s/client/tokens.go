package client

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

func (c *k8sdClient) GetJoinToken(ctx context.Context, request apiv1.GetJoinTokenRequest) (string, error) {
	response := apiv1.GetJoinTokenResponse{}
	if err := c.mc.Query(ctx, "POST", api.NewURL().Path("k8sd", "cluster", "tokens"), request, &response); err != nil {
		return "", fmt.Errorf("failed to POST /k8sd/cluster/tokens: %w", err)
	}
	return response.EncodedToken, nil
}
