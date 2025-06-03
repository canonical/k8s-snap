package upgrade

import (
	"context"
	"fmt"
	"time"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/state"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	defaultFeatureControllerReadyTimeout     = 30 * time.Second
	defaultFeatureControllerReconcileTimeout = 30 * time.Second
)

type Controller struct {
	snap                              snap.Snap
	waitReady                         func()
	featureControllerReadyCh          <-chan struct{}
	notifyFeatureController           func(network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns bool)
	featureToReconciledCh             map[string]<-chan struct{}
	featureControllerReadyTimeout     time.Duration
	featureControllerReconcileTimeout time.Duration

	getState func() state.State
	manager  manager.Manager
	logger   logr.Logger
}

type ControllerOptions struct {
	// Snap is the snap instance.
	Snap snap.Snap
	// WaitReady is a function that waits for the Microcluster to be ready.
	WaitReady func()
	// FeatureControllerReadyCh is a channel that is closed when the feature controller is ready.
	FeatureControllerReadyCh <-chan struct{}
	// NotifyFeatureController is a function that notifies the feature controller to reconcile.
	NotifyFeatureController func(network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns bool)
	// FeatureToReconciledCh is a map of feature names to channels that are full
	// when the feature controller has reconciled the feature.
	FeatureToReconciledCh map[string]<-chan struct{}
	// FeatureControllerReadyTimeout is the timeout for the feature controller to be ready.
	FeatureControllerReadyTimeout time.Duration
	// FeatureControllerReconcileTimeout is the timeout for each feature to get reconciled by the feature controller.
	FeatureControllerReconcileTimeout time.Duration
}

func NewController(opts ControllerOptions) *Controller {
	return &Controller{
		snap:                              opts.Snap,
		waitReady:                         opts.WaitReady,
		featureControllerReadyCh:          opts.FeatureControllerReadyCh,
		notifyFeatureController:           opts.NotifyFeatureController,
		featureToReconciledCh:             opts.FeatureToReconciledCh,
		featureControllerReadyTimeout:     opts.FeatureControllerReadyTimeout,
		featureControllerReconcileTimeout: opts.FeatureControllerReconcileTimeout,
	}
}

func (c *Controller) Run(
	ctx context.Context,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
	getState func() state.State,
) error {
	logger := log.FromContext(ctx).WithName("upgrade-controller")
	ctx = log.NewContext(ctx, logger)

	c.waitReady()

	clusterConfig, err := getClusterConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster configuration: %w", err)
	}

	if featureUpgradesDisabled(clusterConfig) {
		logger.Info("Feature upgrades are disabled. Skipping upgrade controller.")
		return nil
	}

	config, err := c.getRESTConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes REST config: %w", err)
	}

	scheme := runtime.NewScheme()
	if err := kubernetes.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add scheme: %w", err)
	}

	// TODO(Hue): (KU-3216) use a single manager for upgrade and csrsigning controllers.
	mgr, err := manager.New(config, manager.Options{
		Scheme:                  scheme,
		Logger:                  logger,
		LeaderElection:          true,
		LeaderElectionID:        "a27980c4.k8sd-upgrade-controller",
		LeaderElectionNamespace: "kube-system",
		BaseContext:             func() context.Context { return ctx },
		Cache: cache.Options{
			SyncPeriod: utils.Pointer(10 * time.Minute),
		},
		Metrics: server.Options{
			BindAddress: "0",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	c.getState = getState
	c.manager = mgr
	c.logger = mgr.GetLogger()

	if err := c.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup controller with manager: %w", err)
	}

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start manager: %w", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubernetes.Upgrade{}).
		WithOptions(controller.Options{
			RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
				time.Second,
				5*time.Minute,
			),
		}).
		Complete(c)
}

func (c *Controller) getRESTConfig(ctx context.Context) (*rest.Config, error) {
	for {
		client, err := c.snap.KubernetesClient("")
		if err == nil {
			return client.RESTConfig(), nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

// featureUpgradesDisabled checks if feature upgrades are disabled in the cluster configuration.
func featureUpgradesDisabled(clusterConfig types.ClusterConfig) bool {
	_, ok := clusterConfig.Annotations.Get(apiv1_annotations.AnnotationDisableSeparateFeatureUpgrades)
	return ok
}
