package controllers

import (
	"context"
	"fmt"
	"sync"
	"time"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations"
	upgradesv1alpha "github.com/canonical/k8s/pkg/k8sd/crds/upgrades/v1alpha"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	timeutils "github.com/canonical/k8s/pkg/utils/time"
	"github.com/canonical/microcluster/v2/state"
)

// FeatureController manages the lifecycle of built-in Canonical Kubernetes features on a running cluster.
// The controller has separate trigger channels for each feature.
type FeatureController struct {
	snap      snap.Snap
	waitReady func()

	readyCh chan struct{}

	triggerNetworkCh       chan struct{}
	triggerGatewayCh       chan struct{}
	triggerIngressCh       chan struct{}
	triggerLoadBalancerCh  chan struct{}
	triggerDNSCh           chan struct{}
	triggerLocalStorageCh  chan struct{}
	triggerMetricsServerCh chan struct{}

	// TODO(Hue): (KU-3219) Change these with an atomic bool or something similar.
	// Because we don't close them when the feature is reconciled, we simply
	// put something into them. And that thing is going to be gone as soon as we
	// read from these channels. So "checking to see if a feature is reconciled",
	// will technically cause it to be considered "not-reconciled" immediately.
	reconciledNetworkCh       chan struct{}
	reconciledGatewayCh       chan struct{}
	reconciledIngressCh       chan struct{}
	reconciledLoadBalancerCh  chan struct{}
	reconciledDNSCh           chan struct{}
	reconciledLocalStorageCh  chan struct{}
	reconciledMetricsServerCh chan struct{}

	// reconcileLoopMaxRetryAttempts is the maximum number of retry attempts for the reconcile loop.
	// Zero or negative values mean unlimited retries.
	reconcileLoopMaxRetryAttempts int

	// ciliumLock is a mutex to ensure that only one reconciliation is
	// happening for Cilium-related features (that operate on the `ck-network` chart) at a time.
	// Currently, these features include:
	// - Network
	// - Gateway
	// - Ingress
	ciliumLock sync.Mutex
}

// ReadyCh returns a channel that is closed when the controller is ready.
// This is used to signal to other components that they can start using the controller.
func (c *FeatureController) ReadyCh() <-chan struct{} {
	return c.readyCh
}

func (c *FeatureController) ReconciledNetworkCh() <-chan struct{} {
	return c.reconciledNetworkCh
}

func (c *FeatureController) ReconciledGatewayCh() <-chan struct{} {
	return c.reconciledGatewayCh
}

func (c *FeatureController) ReconciledIngressCh() <-chan struct{} {
	return c.reconciledIngressCh
}

func (c *FeatureController) ReconciledLoadBalancerCh() <-chan struct{} {
	return c.reconciledLoadBalancerCh
}

func (c *FeatureController) ReconciledDNSCh() <-chan struct{} {
	return c.reconciledDNSCh
}

func (c *FeatureController) ReconciledLocalStorageCh() <-chan struct{} {
	return c.reconciledLocalStorageCh
}

func (c *FeatureController) ReconciledMetricsServerCh() <-chan struct{} {
	return c.reconciledMetricsServerCh
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

	// ReconcileLoopMaxRetryAttempts is the maximum number of retry attempts for the reconcile loop.
	// Zero or negative values mean unlimited retries.
	ReconcileLoopMaxRetryAttempts int
}

func NewFeatureController(opts FeatureControllerOpts) *FeatureController {
	return &FeatureController{
		snap:                          opts.Snap,
		waitReady:                     opts.WaitReady,
		readyCh:                       make(chan struct{}),
		triggerNetworkCh:              opts.TriggerNetworkCh,
		triggerGatewayCh:              opts.TriggerGatewayCh,
		triggerIngressCh:              opts.TriggerIngressCh,
		triggerLoadBalancerCh:         opts.TriggerLoadBalancerCh,
		triggerDNSCh:                  opts.TriggerDNSCh,
		triggerLocalStorageCh:         opts.TriggerLocalStorageCh,
		triggerMetricsServerCh:        opts.TriggerMetricsServerCh,
		reconciledNetworkCh:           make(chan struct{}, 1),
		reconciledGatewayCh:           make(chan struct{}, 1),
		reconciledIngressCh:           make(chan struct{}, 1),
		reconciledLoadBalancerCh:      make(chan struct{}, 1),
		reconciledDNSCh:               make(chan struct{}, 1),
		reconciledLocalStorageCh:      make(chan struct{}, 1),
		reconciledMetricsServerCh:     make(chan struct{}, 1),
		reconcileLoopMaxRetryAttempts: opts.ReconcileLoopMaxRetryAttempts,
	}
}

func (c *FeatureController) Run(
	ctx context.Context,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
	getState func() state.State,
	notifyDNSChangedIP func(ctx context.Context, dnsIP string) error,
	setFeatureStatus func(ctx context.Context, name types.FeatureName, featureStatus types.FeatureStatus) error,
) {
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "feature"))
	log := log.FromContext(ctx)
	c.waitReady()

	isWorker, err := snaputil.IsWorker(c.snap)
	if err != nil {
		log.Error(err, "Failed to determine if snap is running as worker")
	}
	if isWorker {
		log.Info("Skipping feature controller on worker node")
		return
	}

	log.Info("Starting feature controller")

	s := getState()

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.Network, c.triggerNetworkCh, c.reconciledNetworkCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		c.ciliumLock.Lock()
		defer c.ciliumLock.Unlock()
		return features.Implementation.ApplyNetwork(ctx, c.snap, s, cfg.APIServer, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.Gateway, c.triggerGatewayCh, c.reconciledGatewayCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		c.ciliumLock.Lock()
		defer c.ciliumLock.Unlock()
		return features.Implementation.ApplyGateway(ctx, c.snap, cfg.Gateway, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.Ingress, c.triggerIngressCh, c.reconciledIngressCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		c.ciliumLock.Lock()
		defer c.ciliumLock.Unlock()
		return features.Implementation.ApplyIngress(ctx, c.snap, cfg.Ingress, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.LoadBalancer, c.triggerLoadBalancerCh, c.reconciledLoadBalancerCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return features.Implementation.ApplyLoadBalancer(ctx, c.snap, cfg.LoadBalancer, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.LocalStorage, c.triggerLocalStorageCh, c.reconciledLocalStorageCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return features.Implementation.ApplyLocalStorage(ctx, c.snap, cfg.LocalStorage, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.MetricsServer, c.triggerMetricsServerCh, c.reconciledMetricsServerCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return features.Implementation.ApplyMetricsServer(ctx, c.snap, cfg.MetricsServer, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, getClusterConfig, setFeatureStatus, features.DNS, c.triggerDNSCh, c.reconciledDNSCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
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

	close(c.readyCh)
	log.Info("Feature controller ready")
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
	var attempts int

	for {
		select {
		case <-ctx.Done():
			return
		case <-triggerCh:
			log := log.FromContext(ctx).WithValues("feature", featureName)

			// reset "reconciled" state before reconciling
			utils.MaybeReceive(reconciledCh)

			blocked, err := c.isBlocked(ctx, getClusterConfig)
			if err != nil {
				log.Error(err, "Failed to check if feature controller is blocked")
				// notify triggerCh after 5 seconds to retry
				time.AfterFunc(5*time.Second, func() { utils.MaybeNotify(triggerCh) })
				continue
			}

			if blocked {
				continue
			}

			if err := c.reconcile(ctx, getClusterConfig, apply, func(ctx context.Context, status types.FeatureStatus) error {
				return setFeatureStatus(ctx, featureName, status)
			}); err != nil {
				log.Error(err, "Failed to apply feature configuration")
				attempts++

				maxAttempts := fmt.Sprintf("%d", c.reconcileLoopMaxRetryAttempts)
				if c.reconcileLoopMaxRetryAttempts <= 0 {
					maxAttempts = "unlimited"
				}

				if attempts >= c.reconcileLoopMaxRetryAttempts && c.reconcileLoopMaxRetryAttempts > 0 {
					log.Error(err, "Failed to apply feature configuration after maximum retry attempts", "attempts", fmt.Sprintf("%d/%s", attempts, maxAttempts))
					// NOTE(Hue): we don't notify the triggerCh here, because we want to stop retrying
					// We also set the attempts to 0, so that the next time we receive a trigger,
					// we start from 0 again.
					attempts = 0
					continue
				}

				log.Info("Retrying feature reconciliation", "attempts", fmt.Sprintf("%d/%s", attempts, maxAttempts))
				// notify triggerCh after 3-15 seconds to retry
				time.AfterFunc(timeutils.ExponentialBackoff(attempts, 3*time.Second, 5*time.Minute), func() { utils.MaybeNotify(triggerCh) })
			} else {
				utils.MaybeNotify(reconciledCh)
				attempts = 0
			}

		}
	}
}

// isBlocked checks if the feature controller is blocked by an in-progress upgrade.
// If an upgrade is in progress, the feature controller will not apply any configuration changes.
func (c *FeatureController) isBlocked(ctx context.Context, getClusterConfig func(context.Context) (types.ClusterConfig, error)) (bool, error) {
	log := log.FromContext(ctx)

	clusterConfig, err := getClusterConfig(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve cluster configuration: %w", err)
	}
	// Skip feature reconciliation while an upgrade is in progress to avoid conflicting cluster
	// configuration changes.
	if _, ok := clusterConfig.Annotations.Get(apiv1_annotations.AnnotationDisableSeparateFeatureUpgrades); !ok {
		k8sClient, err := c.snap.KubernetesClient("")
		if err != nil {
			return false, fmt.Errorf("failed to get Kubernetes client: %w", err)
		}

		upgrade, err := k8sClient.GetInProgressUpgrade(ctx)
		if err != nil {
			return false, fmt.Errorf("failed to check for in-progress upgrade: %w", err)
		}

		if upgrade == nil {
			return false, nil
		}

		if upgrade.Status.Phase == upgradesv1alpha.UpgradePhaseFeatureUpgrade {
			log.Info("Upgrade in progress - but in feature upgrade phase - applying configuration", "upgrade", upgrade.Name, "phase", upgrade.Status.Phase)
			return false, nil
		}

		log.Info("Upgrade in progress - feature controller blocked", "upgrade", upgrade.Name, "phase", upgrade.Status.Phase)
		return true, nil
	}

	return false, nil
}
