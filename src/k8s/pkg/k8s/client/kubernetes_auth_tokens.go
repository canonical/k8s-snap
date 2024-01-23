package client

import (
	"context"
	"fmt"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

// GenerateAuthToken calls "POST 1.0/kubernetes/auth/tokens".
func (c *Client) GenerateAuthToken(ctx context.Context, username string, groups []string) (string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	request := apiv1.CreateKubernetesAuthTokenRequest{Username: username, Groups: groups}
	response := apiv1.CreateKubernetesAuthTokenResponse{}

	err := c.mc.Query(queryCtx, "POST", api.NewURL().Path("kubernetes", "auth", "tokens"), request, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return "", fmt.Errorf("failed to query endpoint POST /kubernetes/auth/tokens on %q: %w", clientURL.String(), err)
	}

	return response.Token, nil
}
