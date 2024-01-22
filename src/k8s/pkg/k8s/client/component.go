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
		return nil, fmt.Errorf("failed to query endpoint GET /k8sd/components on %q: %w", clientURL.String(), err)
	}
	return response.Components, nil
}

func (c *Client) UpdateDNSComponent(ctx context.Context, request api.UpdateDNSComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateDNSComponentResponse
	err := c.mc.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "dns"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to enable dns component: %w", err)
	}
	return nil
}

func (c *Client) UpdateNetworkComponent(ctx context.Context, request api.UpdateNetworkComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateNetworkComponentResponse
	err := c.mc.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "network"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to enable network component: %w", err)
	}
	return nil
}

func (c *Client) UpdateStorageComponent(ctx context.Context, request api.UpdateStorageComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateStorageComponentResponse
	err := c.mc.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "storage"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to enable storage component: %w", err)
	}
	return nil
}

func (c *Client) UpdateIngressComponent(ctx context.Context, request api.UpdateIngressComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateIngressComponentResponse
	err := c.mc.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "ingress"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to enable ingress component: %w", err)
	}
	return nil
}
