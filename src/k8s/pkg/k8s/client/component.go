package client

import (
	"context"
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	lxdApi "github.com/canonical/lxd/shared/api"
)

// ListComponents returns the k8s components.
func (c *k8sdClient) ListComponents(ctx context.Context) ([]api.Component, error) {
	var response api.GetComponentsResponse
	err := c.Query(ctx, "GET", lxdApi.NewURL().Path("k8sd", "components"), nil, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return nil, fmt.Errorf("failed to query endpoint GET /k8sd/components on %q: %w", clientURL.String(), err)
	}
	return response.Components, nil
}
