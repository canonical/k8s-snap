package k8sd

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

func (c *k8sd) SetClusterAPIAuthToken(ctx context.Context, request apiv1.SetClusterAPIAuthTokenRequest) error {
	if err := c.client.Query(ctx, "POST", apiv1.K8sdVersionPrefix, api.NewURL().Path("x", "capi", "set-auth-token"), request, nil); err != nil {
		return fmt.Errorf("failed to POST /x/capi/set-auth-token: %w", err)
	}
	return nil
}
