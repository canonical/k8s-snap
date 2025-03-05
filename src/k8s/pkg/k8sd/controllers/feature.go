package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/client/helm/loader"
	"github.com/canonical/k8s/pkg/k8sd/database"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/features/implementation"
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

func (c *FeatureController) Run(
	ctx context.Context,
	s state.State,
	notifyUpdateNodeConfigController func(),
) {
	c.waitReady()
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "feature"))

	go c.reconcileLoop(ctx, s, notifyUpdateNodeConfigController, features.Network, implementation.Implementation.NewNetworkReconciler, c.triggerNetworkCh, c.reconciledNetworkCh)

	go c.reconcileLoop(ctx, s, notifyUpdateNodeConfigController, features.Gateway, implementation.Implementation.NewGatewayReconciler, c.triggerGatewayCh, c.reconciledGatewayCh)

	go c.reconcileLoop(ctx, s, notifyUpdateNodeConfigController, features.Ingress, implementation.Implementation.NewIngressReconciler, c.triggerIngressCh, c.reconciledIngressCh)

	go c.reconcileLoop(ctx, s, notifyUpdateNodeConfigController, features.LoadBalancer, implementation.Implementation.NewLoadBalancerReconciler, c.triggerLoadBalancerCh, c.reconciledLoadBalancerCh)

	go c.reconcileLoop(ctx, s, notifyUpdateNodeConfigController, features.LocalStorage, implementation.Implementation.NewLocalStorageReconciler, c.triggerLocalStorageCh, c.reconciledLocalStorageCh)

	go c.reconcileLoop(ctx, s, notifyUpdateNodeConfigController, features.MetricsServer, implementation.Implementation.NewMetricsServerReconciler, c.triggerMetricsServerCh, c.reconciledMetricsServerCh)

	go c.reconcileLoop(ctx, s, notifyUpdateNodeConfigController, features.DNS, implementation.Implementation.NewDNSReconciler, c.triggerDNSCh, c.reconciledDNSCh)
}

func (c *FeatureController) reconcile(
	ctx context.Context,
	s state.State,
	notifyUpdateNodeConfigController func(),
	newReconciler func(base features.BaseReconciler) features.Reconciler,
	updateFeatureStatus func(context.Context, types.FeatureStatus) error,
) error {
	cfg, err := c.getClusterConfig(ctx, s)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster configuration: %w", err)
	}

	// helm client with database loader
	m := c.snap.HelmClient(loader.NewDatabaseLoader(s))

	base := features.NewReconciler(c.snap, m, s, notifyUpdateNodeConfigController)
	reconciler := newReconciler(base)

	status, applyErr := reconciler.Reconcile(ctx, cfg)
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
	s state.State,
	notifyUpdateNodeConfigController func(),
	featureName types.FeatureName,
	newReconciler func(base features.BaseReconciler) features.Reconciler,
	triggerCh chan struct{},
	reconciledCh chan<- struct{},
) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-triggerCh:
			if err := c.reconcile(ctx, s, notifyUpdateNodeConfigController, newReconciler, func(ctx context.Context, status types.FeatureStatus) error {
				return c.setFeatureStatus(ctx, s, featureName, status)
			}); err != nil {
				log.FromContext(ctx).WithValues("feature", featureName).Error(err, "Failed to apply feature configuration")

				// notify triggerCh after 5 seconds to retry
				time.AfterFunc(5*time.Second, func() { utils.MaybeNotify(triggerCh) })
			} else {
				utils.MaybeNotify(reconciledCh)
			}

		}
	}
}

func (c *FeatureController) getClusterConfig(ctx context.Context, s state.State) (types.ClusterConfig, error) {
	return databaseutil.GetClusterConfig(ctx, s)
}

func (c *FeatureController) setFeatureStatus(ctx context.Context, s state.State, name types.FeatureName, featureStatus types.FeatureStatus) error {
	if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// we set timestamp here in order to reduce the clutter. otherwise we will need to
		// set .UpdatedAt field in a lot of places for every event/error.
		// this is not 100% accurate but should be good enough
		featureStatus.UpdatedAt = time.Now()
		if err := database.SetFeatureStatus(ctx, tx, name, featureStatus); err != nil {
			return fmt.Errorf("failed to set feature status in db for %q: %w", name, err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database transaction to set feature status failed: %w", err)
	}
	return nil
}
