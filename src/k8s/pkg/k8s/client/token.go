package client

import (
	"context"
	"fmt"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

// CreateJoinToken calls "POST 1.0/k8sd/tokens"
func (c *Client) CreateJoinToken(ctx context.Context, name string) (string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	request := apiv1.CreateJoinTokenRequest{Name: name}
	var response apiv1.CreateJoinTokenResponse
	err := c.mc.Query(queryCtx, "POST", api.NewURL().Path("k8sd", "tokens"), request, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return "", fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}
	return response.Token, nil
}
