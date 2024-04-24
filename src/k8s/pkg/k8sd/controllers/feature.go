package controllers

import (
	"context"
	"fmt"
	"log"

	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// FeatureController manages the lifecycle of built-in Canonical Kubernetes features on a running cluster.
// The controller has separate trigger channels for each feature.
type FeatureController struct {
	snap      snap.Snap
	waitReady func()

	triggerNetworkCh       <-chan struct{}
	triggerDNSCh           <-chan struct{}
	triggerLocalStorageCh  <-chan struct{}
	triggerMetricsServerCh <-chan struct{}

	reconciledNetworkCh       chan struct{}
	reconciledDNSCh           chan struct{}
	reconciledLocalStorageCh  chan struct{}
	reconciledMetricsServerCh chan struct{}
}

func NewFeatureController(
	snap snap.Snap,
	waitReady func(),
	triggerNetworkCh <-chan struct{},
	triggerDNSCh <-chan struct{},
	triggerLocalStorageCh <-chan struct{},
	triggerMetricsServerCh <-chan struct{},
) *FeatureController {
	return &FeatureController{
		snap:                      snap,
		waitReady:                 waitReady,
		triggerNetworkCh:          triggerNetworkCh,
		triggerDNSCh:              triggerDNSCh,
		triggerLocalStorageCh:     triggerLocalStorageCh,
		triggerMetricsServerCh:    triggerMetricsServerCh,
		reconciledNetworkCh:       make(chan struct{}, 1),
		reconciledDNSCh:           make(chan struct{}, 1),
		reconciledLocalStorageCh:  make(chan struct{}, 1),
		reconciledMetricsServerCh: make(chan struct{}, 1),
	}
}

func (c *FeatureController) Run(ctx context.Context, getClusterConfig func(context.Context) (types.ClusterConfig, error), notifyDNSChangedIP func(ctx context.Context, dnsIP string) error) {
	// TODO: split the network components into separate reconcileLoops (and use separate channels) after implementing retry logic in case of failures
	go c.reconcileLoop(ctx, getClusterConfig, "network", c.triggerNetworkCh, c.reconciledNetworkCh, func(cfg types.ClusterConfig) error {
		if err := features.ApplyNetwork(ctx, c.snap, cfg.Network); err != nil {
			return fmt.Errorf("failed to apply CNI configuration: %w", err)
		}
		if err := features.ApplyGateway(ctx, c.snap, cfg.Gateway, cfg.Network); err != nil {
			return fmt.Errorf("failed to apply gateway configuration: %w", err)
		}
		if err := features.ApplyIngress(ctx, c.snap, cfg.Ingress, cfg.Network); err != nil {
			return fmt.Errorf("failed to apply ingress configuration: %w", err)
		}
		if err := features.ApplyLoadBalancer(ctx, c.snap, cfg.LoadBalancer, cfg.Network); err != nil {
			return fmt.Errorf("failed to apply load balancer configuration: %w", err)
		}
		return nil
	})

	go c.reconcileLoop(ctx, getClusterConfig, "local storage", c.triggerLocalStorageCh, c.reconciledLocalStorageCh, func(cfg types.ClusterConfig) error {
		return features.ApplyLocalStorage(ctx, c.snap, cfg.LocalStorage)
	})

	go c.reconcileLoop(ctx, getClusterConfig, "metrics server", c.triggerMetricsServerCh, c.reconciledMetricsServerCh, func(cfg types.ClusterConfig) error {
		return features.ApplyMetricsServer(ctx, c.snap, cfg.MetricsServer)
	})

	go c.reconcileLoop(ctx, getClusterConfig, "DNS", c.triggerDNSCh, c.reconciledDNSCh, func(cfg types.ClusterConfig) error {
		if dnsIP, err := features.ApplyDNS(ctx, c.snap, cfg.DNS, cfg.Kubelet); err != nil {
			return fmt.Errorf("failed to apply DNS configuration: %w", err)
		} else if dnsIP != "" {
			if err := notifyDNSChangedIP(ctx, dnsIP); err != nil {
				return fmt.Errorf("failed to update DNS IP address to %s: %w", dnsIP, err)
			}
		}
		return nil
	})
}

func (c *FeatureController) reconcile(ctx context.Context, getClusterConfig func(context.Context) (types.ClusterConfig, error), apply func(cfg types.ClusterConfig) error) error {
	cfg, err := getClusterConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster configuration: %w", err)
	}

	if err := apply(cfg); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}
	return nil
}

func (c *FeatureController) reconcileLoop(ctx context.Context, getClusterConfig func(context.Context) (types.ClusterConfig, error), componentName string, triggerCh <-chan struct{}, reconciledCh chan<- struct{}, apply func(cfg types.ClusterConfig) error) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-triggerCh:
			// TODO: add retry logic in case of error
			if err := c.reconcile(ctx, getClusterConfig, apply); err != nil {
				log.Println(fmt.Errorf("failed to reconcile %s configuration: %w", componentName, err))
			}

			select {
			case reconciledCh <- struct{}{}:
			default:
			}
		}
	}
}
