package client

import (
	"context"
	"fmt"
	"time"

	api "github.com/canonical/k8s/api/v1"
	lxdApi "github.com/canonical/lxd/shared/api"
)

// CreateJoinToken calls "POST 1.0/k8sd/tokens"
func (c *Client) CreateJoinToken(ctx context.Context, name string) (string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	request := api.CreateJoinTokenRequest{Name: name}
	var response api.CreateJoinTokenResponse
	err := c.mc.Query(queryCtx, "POST", lxdApi.NewURL().Path("k8sd", "tokens"), request, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return "", fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}
	return response.Token, nil
}
