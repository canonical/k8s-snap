package cluster

import (
	"context"
	"fmt"

	v1 "github.com/canonical/k8s/api/v1"
)

// GenerateAuthToken calls "POST 1.0/k8sd/tokens".
func (c *Client) GenerateAuthToken(ctx context.Context, username string, groups []string) (string, error) {
	request := v1.CreateTokenRequest{Username: username, Groups: groups}
	response := v1.CreateTokenResponse{}
	if err := c.doHTTP(ctx, "POST", "1.0/k8sd/tokens", request, &response); err != nil {
		return "", fmt.Errorf("POST 1.0/k8sd/tokens failed: %w", err)
	}

	return response.Token, nil
}
