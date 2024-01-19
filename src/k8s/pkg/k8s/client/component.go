package client

import (
	"context"
	"fmt"
	"time"

	api "github.com/canonical/k8s/api/v1"
	lxdApi "github.com/canonical/lxd/shared/api"
)

// ListComponents returns the k8s components.
func (c *Client) ListComponents(ctx context.Context) ([]api.Component, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.GetComponentsResponse
	err := c.mc.Query(queryCtx, "GET", lxdApi.NewURL().Path("k8sd", "components"), nil, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return nil, fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}
	return response.Components, nil
}

func (c *Client) UpdateDNSComponent(ctx context.Context, request api.UpdateDNSComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateDNSComponentResponse
	// TODO: This URL is a temporary measure to prevent collisions with the /k8sd/components/{name} path
	err := c.mc.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "dns:enable"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to enable dns component: %w", err)
	}
	return nil
}

// UpdateComponent updates the state of a component.
func (c *Client) UpdateComponent(ctx context.Context, name string, status api.ComponentStatus) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	request := api.UpdateComponentRequest{
		Status: status,
	}
	var response api.UpdateComponentResponse
	err := c.mc.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", name), request, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}
	return nil
}
