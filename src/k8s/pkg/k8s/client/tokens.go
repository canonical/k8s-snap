package client

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

func (c *Client) CreateJoinToken(ctx context.Context, name string, worker bool) (string, error) {
	request := apiv1.TokenRequest{
		Name:   name,
		Worker: worker,
	}
	response := apiv1.TokensResponse{}

	err := c.mc.Query(ctx, "POST", api.NewURL().Path("k8sd", "tokens"), request, &response)
	if err != nil {
		return "", fmt.Errorf("failed to query endpoint POST /k8sd/tokens: %w", err)
	}
	return response.EncodedToken, nil
}
