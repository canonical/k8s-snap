package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/database"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/features/featureset"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	microclusterutil "github.com/canonical/k8s/pkg/utils/microcluster"
	"github.com/canonical/microcluster/v2/state"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	go c.reconcileLoop(ctx, s, features.Network, c.triggerNetworkCh, c.reconciledNetworkCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return featureset.Reconcile.ApplyNetwork(ctx, s, c.snap, cfg.APIServer, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, s, features.Gateway, c.triggerGatewayCh, c.reconciledGatewayCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return featureset.Reconcile.ApplyGateway(ctx, s, c.snap, cfg.Gateway, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, s, features.Ingress, c.triggerIngressCh, c.reconciledIngressCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return featureset.Reconcile.ApplyIngress(ctx, s, c.snap, cfg.Ingress, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, s, features.LoadBalancer, c.triggerLoadBalancerCh, c.reconciledLoadBalancerCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return featureset.Reconcile.ApplyLoadBalancer(ctx, s, c.snap, cfg.LoadBalancer, cfg.Network, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, s, features.LocalStorage, c.triggerLocalStorageCh, c.reconciledLocalStorageCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return featureset.Reconcile.ApplyLocalStorage(ctx, s, c.snap, cfg.LocalStorage, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, s, features.MetricsServer, c.triggerMetricsServerCh, c.reconciledMetricsServerCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return featureset.Reconcile.ApplyMetricsServer(ctx, s, c.snap, cfg.MetricsServer, cfg.Annotations)
	})

	go c.reconcileLoop(ctx, s, features.DNS, c.triggerDNSCh, c.reconciledDNSCh, func(cfg types.ClusterConfig) (types.FeatureStatus, error) {
		return featureset.Reconcile.ApplyDNS(ctx, s, c.snap, cfg.DNS, cfg.Kubelet, cfg.Annotations)
	})

	go c.reconcileClusterDNS(ctx, s, notifyUpdateNodeConfigController)
}

func (c *FeatureController) reconcile(
	ctx context.Context,
	s state.State,
	apply func(cfg types.ClusterConfig) (types.FeatureStatus, error),
	updateFeatureStatus func(context.Context, types.FeatureStatus) error,
) error {
	cfg, err := c.getClusterConfig(ctx, s)
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
	s state.State,
	featureName types.FeatureName,
	triggerCh chan struct{},
	reconciledCh chan<- struct{},
	apply func(cfg types.ClusterConfig) (types.FeatureStatus, error),
) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-triggerCh:
			log := log.FromContext(ctx).WithValues("feature", featureName)

			isLeader, err := microclusterutil.IsLeader(s)
			if err != nil {
				log.Error(err, "Failed to check if node is leader")
				continue
			}

			if !isLeader {
				log.Info("Skipping feature reconcilation on non-leader node")
				continue
			}

			if err := c.reconcile(ctx, s, apply, func(ctx context.Context, status types.FeatureStatus) error {
				return c.setFeatureStatus(ctx, s, featureName, status)
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

func (c *FeatureController) getClusterConfig(ctx context.Context, s state.State) (types.ClusterConfig, error) {
	return databaseutil.GetClusterConfig(ctx, s)
}

func (c *FeatureController) notifyClusterDNSChange(ctx context.Context, s state.State, notifyUpdateNodeConfigController func()) {
	log := log.FromContext(ctx)
	for {
		cfg, err := c.getClusterConfig(ctx, s)
		if err != nil {
			log.Error(err, "Failed to retrieve cluster configuration")
			continue
		}

		client, err := getNewK8sClientWithRetries(ctx, c.snap, false)
		if err != nil {
			log.Error(err, "Failed to create a Kubernetes client")
			continue
		}

		configMap, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "k8sd-config", metav1.GetOptions{})
		if err != nil {
			log.Error(err, "Failed to get k8sd-config configmap")
			continue
		}

		current := ""
		if cfg.Kubelet.ClusterDNS != nil {
			current = *cfg.Kubelet.ClusterDNS
		}

		existing, ok := configMap.Data["cluster-dns"]
		if !ok {
			existing = ""
		}

		if existing != current {
			log.Info("Cluster DNS changed, notifying update node configuration controller")
			notifyUpdateNodeConfigController()
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (c *FeatureController) reconcileClusterDNS(ctx context.Context, s state.State, notifyUpdateNodeConfigController func()) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.reconciledDNSCh:
			c.notifyClusterDNSChange(ctx, s, notifyUpdateNodeConfigController)
		}
	}
}
