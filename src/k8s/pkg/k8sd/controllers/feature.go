package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/state"
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

	ReconciledNetworkCh       chan struct{}
	ReconciledGatewayCh       chan struct{}
	ReconciledIngressCh       chan struct{}
	ReconciledLoadBalancerCh  chan struct{}
	ReconciledDNSCh           chan struct{}
	ReconciledLocalStorageCh  chan struct{}
	ReconciledMetricsServerCh chan struct{}
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
		ReconciledNetworkCh:       make(chan struct{}, 1),
		ReconciledGatewayCh:       make(chan struct{}, 1),
		ReconciledIngressCh:       make(chan struct{}, 1),
		ReconciledLoadBalancerCh:  make(chan struct{}, 1),
		ReconciledDNSCh:           make(chan struct{}, 1),
		ReconciledLocalStorageCh:  make(chan struct{}, 1),
		ReconciledMetricsServerCh: make(chan struct{}, 1),
	}
}

func (c *FeatureController) Run(
	ctx context.Context,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
	getState func() state.State,
	notifyDNSChangedIP func(ctx context.Context, dnsIP string) error,
	setFeatureStatus func(ctx context.Context, name types.FeatureName, featureStatus types.FeatureStatus) error,
) {
	c.waitReady()
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "feature"))

	s := getState()

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.Network, c.triggerNetworkCh, c.ReconciledNetworkCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return features.Implementation.ApplyNetwork(ctx, c.snap, s, cfg.APIServer, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.Gateway, c.triggerGatewayCh, c.ReconciledGatewayCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return features.Implementation.ApplyGateway(ctx, c.snap, cfg.Gateway, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.Ingress, c.triggerIngressCh, c.ReconciledIngressCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return features.Implementation.ApplyIngress(ctx, c.snap, cfg.Ingress, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.LoadBalancer, c.triggerLoadBalancerCh, c.ReconciledLoadBalancerCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return features.Implementation.ApplyLoadBalancer(ctx, c.snap, cfg.LoadBalancer, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.LocalStorage, c.triggerLocalStorageCh, c.ReconciledLocalStorageCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return features.Implementation.ApplyLocalStorage(ctx, c.snap, cfg.LocalStorage, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.MetricsServer, c.triggerMetricsServerCh, c.ReconciledMetricsServerCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return features.Implementation.ApplyMetricsServer(ctx, c.snap, cfg.MetricsServer, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.DNS, c.triggerDNSCh, c.ReconciledDNSCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		featureStatus, dnsIP, err := features.Implementation.ApplyDNS(ctx, c.snap, cfg.DNS, cfg.Kubelet, cfg.Annotations)

		if err != nil {
			return featureStatus, fmt.Errorf("failed to apply DNS configuration: %w", err)
		} else if dnsIP != "" {
			if err := notifyDNSChangedIP(ctx, dnsIP); err != nil {
				// we already have featureStatus.Message which contains wrapped error of the Apply<Feature>
				// (or empty if no error occurs). we further wrap the error to add the DNS IP change error to the message
				changeErr := fmt.Errorf("failed to update DNS IP address to %s: %w", dnsIP, err)
				featureStatus.Message = fmt.Sprintf("%s: %v", featureStatus.Message, changeErr)
				return featureStatus, changeErr
			}
		}
		return featureStatus, nil
	})
}

func (c *FeatureController) reconcile(
	ctx context.Context,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
	apply func(cfg types.ClusterConfig) (types.FeatureStatus, error),
	updateFeatureStatus func(context.Context, types.FeatureStatus) error,
) error {
	cfg, err := getClusterConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster configuration: %w", err)
	}

	status, applyErr := apply(cfg)
	if err := updateFeatureStatus(ctx, status); err != nil {
		// NOTE (hue): status update errors are not returned but only logged. we might need some retry logic in the future.
		log.FromContext(ctx).WithValues("message", status.Message, "applied-successfully", applyErr == nil).Error(err, "Failed to update feature status")
	}

	if applyErr != nil {
		return fmt.Errorf("failed to apply configuration: %w", applyErr)
	}

	return nil
}

func (c *FeatureController) reconcileLoop(
	ctx context.Context,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
	setFeatureStatus func(ctx context.Context, name types.FeatureName, status types.FeatureStatus) error,
	featureName types.FeatureName,
	triggerCh chan struct{},
	reconciledCh chan struct{},
	apply func(cfg types.ClusterConfig) (types.FeatureStatus, error),
) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-triggerCh:
			log := log.FromContext(ctx).WithValues("feature", featureName)

			// reset "reconciled" state before reconciling
			utils.MaybeReceive(reconciledCh)

			k8sClient, err := c.snap.KubernetesClient("")
			if err != nil {
				log.Error(err, "failed to get Kubernetes client")
				continue
			}

			upgrade, err := k8sClient.GetInProgressUpgrade(ctx)
			if err != nil {
				log.Error(err, "failed to check for in-progress upgrade")
				continue
			}
			log.Info("Upgrade in progress", "upgrade", upgrade)
			if upgrade != nil {
				if upgrade.Status.Phase != kubernetes.UpgradePhaseFeatureUpgrade {
					log.Info("Upgrade in progress - feature controller blocked", "upgrade", upgrade.Metadata.Name, "phase", upgrade.Status.Phase)
					continue
				}
				log.Info("Upgrade in progress - but in feature upgrade phase - applying configuration", "upgrade", upgrade.Metadata.Name, "phase", upgrade.Status.Phase)
			}

			if err := c.reconcile(ctx, getClusterConfig, apply, func(ctx context.Context, status types.FeatureStatus) error {
				return setFeatureStatus(ctx, featureName, status)
			}); err != nil {
				log.Error(err, "Failed to apply feature configuration")

				// notify triggerCh after 5 seconds to retry
				time.AfterFunc(5*time.Second, func() { utils.MaybeNotify(triggerCh) })
			} else {
				utils.MaybeNotify(reconciledCh)
			}

		}
	}
}
