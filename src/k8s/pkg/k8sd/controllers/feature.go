package controllers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

// FeatureController manages the lifecycle of built-in Canonical Kubernetes features on a running cluster.
// The controller has separate trigger channels for each feature.
type FeatureController struct {
	snap      snap.Snap
	waitReady func()

	triggerNetworkCh       chan struct{}
	triggerGatewayCh       chan struct{}
	triggerIngressCh       chan struct{}
	triggerLoadBalancerCh  chan struct{}
	triggerDNSCh           chan struct{}
	triggerLocalStorageCh  chan struct{}
	triggerMetricsServerCh chan struct{}

	reconciledNetworkCh       chan struct{}
	reconciledGatewayCh       chan struct{}
	reconciledIngressCh       chan struct{}
	reconciledLoadBalancerCh  chan struct{}
	reconciledDNSCh           chan struct{}
	reconciledLocalStorageCh  chan struct{}
	reconciledMetricsServerCh chan struct{}
}

type FeatureControllerOpts struct {
	Snap      snap.Snap
	WaitReady func()

	TriggerNetworkCh       chan struct{}
	TriggerGatewayCh       chan struct{}
	TriggerIngressCh       chan struct{}
	TriggerLoadBalancerCh  chan struct{}
	TriggerDNSCh           chan struct{}
	TriggerLocalStorageCh  chan struct{}
	TriggerMetricsServerCh chan struct{}
}

func NewFeatureController(opts FeatureControllerOpts) *FeatureController {
	return &FeatureController{
		snap:                      opts.Snap,
		waitReady:                 opts.WaitReady,
		triggerNetworkCh:          opts.TriggerNetworkCh,
		triggerGatewayCh:          opts.TriggerGatewayCh,
		triggerIngressCh:          opts.TriggerIngressCh,
		triggerLoadBalancerCh:     opts.TriggerLoadBalancerCh,
		triggerDNSCh:              opts.TriggerDNSCh,
		triggerLocalStorageCh:     opts.TriggerLocalStorageCh,
		triggerMetricsServerCh:    opts.TriggerMetricsServerCh,
		reconciledNetworkCh:       make(chan struct{}, 1),
		reconciledGatewayCh:       make(chan struct{}, 1),
		reconciledIngressCh:       make(chan struct{}, 1),
		reconciledLoadBalancerCh:  make(chan struct{}, 1),
		reconciledDNSCh:           make(chan struct{}, 1),
		reconciledLocalStorageCh:  make(chan struct{}, 1),
		reconciledMetricsServerCh: make(chan struct{}, 1),
	}
}

func (c *FeatureController) Run(ctx context.Context, getClusterConfig func(context.Context) (types.ClusterConfig, error), notifyDNSChangedIP func(ctx context.Context, dnsIP string) error) {
	c.waitReady()

	go c.reconcileLoop(ctx, getClusterConfig, "network", c.triggerNetworkCh, c.reconciledNetworkCh, func(cfg types.ClusterConfig) error {
		return features.Implementation.ApplyNetwork(ctx, c.snap, cfg.Network)
	})

	go c.reconcileLoop(ctx, getClusterConfig, "gateway", c.triggerGatewayCh, c.reconciledGatewayCh, func(cfg types.ClusterConfig) error {
		return features.Implementation.ApplyGateway(ctx, c.snap, cfg.Gateway, cfg.Network)
	})

	go c.reconcileLoop(ctx, getClusterConfig, "ingress", c.triggerIngressCh, c.reconciledIngressCh, func(cfg types.ClusterConfig) error {
		return features.Implementation.ApplyIngress(ctx, c.snap, cfg.Ingress, cfg.Network)
	})

	go c.reconcileLoop(ctx, getClusterConfig, "load balancer", c.triggerLoadBalancerCh, c.reconciledLoadBalancerCh, func(cfg types.ClusterConfig) error {
		return features.Implementation.ApplyLoadBalancer(ctx, c.snap, cfg.LoadBalancer, cfg.Network)
	})

	go c.reconcileLoop(ctx, getClusterConfig, "local storage", c.triggerLocalStorageCh, c.reconciledLocalStorageCh, func(cfg types.ClusterConfig) error {
		return features.Implementation.ApplyLocalStorage(ctx, c.snap, cfg.LocalStorage)
	})

	go c.reconcileLoop(ctx, getClusterConfig, "metrics server", c.triggerMetricsServerCh, c.reconciledMetricsServerCh, func(cfg types.ClusterConfig) error {
		return features.Implementation.ApplyMetricsServer(ctx, c.snap, cfg.MetricsServer)
	})

	go c.reconcileLoop(ctx, getClusterConfig, "DNS", c.triggerDNSCh, c.reconciledDNSCh, func(cfg types.ClusterConfig) error {
		if dnsIP, err := features.Implementation.ApplyDNS(ctx, c.snap, cfg.DNS, cfg.Kubelet); err != nil {
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

func (c *FeatureController) reconcileLoop(ctx context.Context, getClusterConfig func(context.Context) (types.ClusterConfig, error), componentName string, triggerCh chan struct{}, reconciledCh chan<- struct{}, apply func(cfg types.ClusterConfig) error) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-triggerCh:
			if err := c.reconcile(ctx, getClusterConfig, apply); err != nil {
				log.Printf("failed to reconcile %s configuration, will retry in 5 seconds: %v", componentName, err)

				// notify triggerCh after 5 seconds to retry
				time.AfterFunc(5*time.Second, func() { utils.MaybeNotify(triggerCh) })
			} else {
				utils.MaybeNotify(reconciledCh)
			}

		}
	}
}
