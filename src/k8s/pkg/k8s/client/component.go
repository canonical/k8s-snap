package client

import (
	"context"
	"fmt"
	"time"

	api "github.com/canonical/k8s/api/v1"
	lxdApi "github.com/canonical/lxd/shared/api"
)

// ListComponents returns the k8s components.
func (c *k8sdClient) ListComponents(ctx context.Context) ([]api.Component, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.GetComponentsResponse
	err := c.Query(queryCtx, "GET", lxdApi.NewURL().Path("k8sd", "components"), nil, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return nil, fmt.Errorf("failed to query endpoint GET /k8sd/components on %q: %w", clientURL.String(), err)
	}
	return response.Components, nil
}

func (c *k8sdClient) UpdateDNSComponent(ctx context.Context, request api.UpdateDNSComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateDNSComponentResponse
	err := c.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "dns"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to enable dns component: %w", err)
	}
	return nil
}

func (c *k8sdClient) UpdateNetworkComponent(ctx context.Context, request api.UpdateNetworkComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateNetworkComponentResponse
	err := c.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "network"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to enable network component: %w", err)
	}
	return nil
}

func (c *k8sdClient) UpdateStorageComponent(ctx context.Context, request api.UpdateStorageComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateStorageComponentResponse
	err := c.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "storage"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to enable storage component: %w", err)
	}
	return nil
}

func (c *k8sdClient) UpdateIngressComponent(ctx context.Context, request api.UpdateIngressComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateIngressComponentResponse
	err := c.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "ingress"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to enable ingress component: %w", err)
	}
	return nil
}

func (c *k8sdClient) UpdateGatewayComponent(ctx context.Context, request api.UpdateGatewayComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateGatewayComponentResponse
	err := c.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "gateway"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to enable gateway component: %w", err)
	}
	return nil
}

func (c *k8sdClient) UpdateLoadBalancerComponent(ctx context.Context, request api.UpdateLoadBalancerComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateLoadBalancerComponentResponse
	if err := c.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "loadbalancer"), request, &response); err != nil {
		return fmt.Errorf("failed to enable loadbalancer component: %w", err)
	}
	return nil
}

func (c *k8sdClient) UpdateMetricsServerComponent(ctx context.Context, request api.UpdateMetricsServerComponentRequest) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.UpdateMetricsServerComponentResponse
	err := c.Query(queryCtx, "PUT", lxdApi.NewURL().Path("k8sd", "components", "metrics-server"), request, &response)
	if err != nil {
		return fmt.Errorf("failed to enable metrics-server component: %w", err)
	}
	return nil
}
