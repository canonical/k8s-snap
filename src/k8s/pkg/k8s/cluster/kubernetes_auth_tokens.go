package cluster

import (
	"context"
	"fmt"

	v1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

// GenerateAuthToken calls "POST 1.0/k8sd/tokens".
func (c *Client) GenerateAuthToken(ctx context.Context, username string, groups []string) (string, error) {
	request := v1.CreateKubernetesAuthTokenRequest{Username: username, Groups: groups}
	response := v1.CreateKubernetesAuthTokenResponse{}
	client, err := c.microClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Query(ctx, "POST", api.NewURL().Path("kubernetes", "auth", "tokens"), request, &response); err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}

	return response.Token, nil
}
