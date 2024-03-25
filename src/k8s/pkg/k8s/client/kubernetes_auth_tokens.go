package client

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

// GenerateAuthToken calls "POST 1.0/kubernetes/auth/tokens".
func (c *k8sdClient) GenerateAuthToken(ctx context.Context, request apiv1.GenerateKubernetesAuthTokenRequest) (string, error) {
	response := apiv1.CreateKubernetesAuthTokenResponse{}

	if err := c.mc.Query(ctx, "POST", api.NewURL().Path("kubernetes", "auth", "tokens"), request, &response); err != nil {
		return "", fmt.Errorf("failed to POST /kubernetes/auth/tokens: %w", err)
	}

	return response.Token, nil
}

// RevokeAuthToken calls "DELETE 1.0/kubernetes/auth/tokens".
func (c *k8sdClient) RevokeAuthToken(ctx context.Context, request apiv1.RevokeKubernetesAuthTokenRequest) error {
	if err := c.mc.Query(ctx, "DELETE", api.NewURL().Path("kubernetes", "auth", "tokens"), request, nil); err != nil {
		return fmt.Errorf("failed to DELETE /kubernetes/auth/tokens: %w", err)
	}

	return nil
}
