package client

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

// GenerateAuthToken calls "POST 1.0/kubernetes/auth/tokens".
func (c *k8sdClient) GenerateAuthToken(ctx context.Context, username string, groups []string) (string, error) {
	request := apiv1.CreateKubernetesAuthTokenRequest{Username: username, Groups: groups}
	response := apiv1.CreateKubernetesAuthTokenResponse{}

	err := c.Query(ctx, "POST", api.NewURL().Path("kubernetes", "auth", "tokens"), request, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return "", fmt.Errorf("failed to query endpoint POST /kubernetes/auth/tokens on %q: %w", clientURL.String(), err)
	}

	return response.Token, nil
}

// RevokeAuthToken calls "DELETE 1.0/kubernetes/auth/tokens".
func (c *k8sdClient) RevokeAuthToken(ctx context.Context, token string) error {
	request := apiv1.RevokeKubernetesAuthTokenRequest{Token: token}

	err := c.Query(ctx, "DELETE", api.NewURL().Path("kubernetes", "auth", "tokens"), request, nil)
	if err != nil {
		clientURL := c.mc.URL()
		return fmt.Errorf("failed to query endpoint DELETE /kubernetes/auth/tokens on %q: %w", clientURL.String(), err)
	}

	return nil
}
