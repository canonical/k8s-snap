package upgrade

import (
	"time"

	upgradesv1alpha1 "github.com/canonical/k8s/pkg/k8sd/crds/upgrades/v1alpha"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/microcluster/v2/state"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	getState                          func() state.State
	logger                            logr.Logger
	client                            client.Client
	featureControllerReadyCh          <-chan struct{}
	notifyFeatureController           func(network, gateway, ingress, dns, loadBalancer, localStorage, metricsServer bool)
	featureToReconciledCh             map[types.FeatureName]<-chan struct{}
	featureControllerReadyTimeout     time.Duration
	featureControllerReconcileTimeout time.Duration
}

type ControllerOptions struct {
	// FeatureControllerReadyCh is a channel that is closed when the feature controller is ready.
	FeatureControllerReadyCh <-chan struct{}
	// NotifyFeatureController is a function that notifies the feature controller to reconcile.
	NotifyFeatureController func(network, gateway, ingress, dns, loadBalancer, localStorage, metricsServer bool)
	// FeatureToReconciledCh is a map of feature names to channels that are full
	// when the feature controller has reconciled the feature.
	FeatureToReconciledCh map[types.FeatureName]<-chan struct{}
	// FeatureControllerReadyTimeout is the timeout for the feature controller to be ready.
	FeatureControllerReadyTimeout time.Duration
	// FeatureControllerReconcileTimeout is the timeout for the feature controller to reconcile.
	FeatureControllerReconcileTimeout time.Duration
}

func NewController(
	getState func() state.State,
	logger logr.Logger,
	client client.Client,
	opts ControllerOptions,
) *Controller {
	return &Controller{
		getState:                          getState,
		logger:                            logger,
		client:                            client,
		featureControllerReadyCh:          opts.FeatureControllerReadyCh,
		notifyFeatureController:           opts.NotifyFeatureController,
		featureToReconciledCh:             opts.FeatureToReconciledCh,
		featureControllerReadyTimeout:     opts.FeatureControllerReadyTimeout,
		featureControllerReconcileTimeout: opts.FeatureControllerReconcileTimeout,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&upgradesv1alpha1.Upgrade{}).
		Complete(c)
}
