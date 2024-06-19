package client

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

// SetClusterAPIAuthToken calls "POST 1.0/x/capi/set-auth-token".
func (c *k8sdClient) SetClusterAPIAuthToken(ctx context.Context, request apiv1.SetClusterAPIAuthTokenRequest) error {
	if err := c.mc.Query(ctx, "POST", api.NewURL().Path("x", "capi", "set-auth-token"), request, nil); err != nil {
		return fmt.Errorf("failed to POST /x/capi/set-auth-token: %w", err)
	}
	return nil
}
