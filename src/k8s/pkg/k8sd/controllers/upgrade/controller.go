package upgrade

import (
	"time"

	upgradesv1alpha1 "github.com/canonical/k8s/pkg/k8sd/crds/upgrades/v1alpha"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Controller struct {
	logger                            logr.Logger
	client                            client.Client
	featureControllerReadyCh          <-chan struct{}
	notifyNetworkFeature              func()
	notifyGatewayFeature              func()
	notifyIngressFeature              func()
	notifyLoadBalancerFeature         func()
	notifyLocalStorageFeature         func()
	notifyMetricsServerFeature        func()
	notifyDNSFeature                  func()
	featureToReconciledCh             map[types.FeatureName]<-chan struct{}
	featureControllerReadyTimeout     time.Duration
	featureControllerReconcileTimeout time.Duration
}

type ControllerOptions struct {
	// FeatureControllerReadyCh is a channel that is closed when the feature controller is ready.
	FeatureControllerReadyCh <-chan struct{}
	// NotifyNetworkFeature is a function that notifies the network feature to reconcile.
	NotifyNetworkFeature func()
	// NotifyGatewayFeature is a function that notifies the gateway feature to reconcile.
	NotifyGatewayFeature func()
	// NotifyIngressFeature is a function that notifies the ingress feature to reconcile.
	NotifyIngressFeature func()
	// NotifyLoadBalancerFeature is a function that notifies the load balancer feature to reconcile.
	NotifyLoadBalancerFeature func()
	// NotifyLocalStorageFeature is a function that notifies the local storage feature to reconcile.
	NotifyLocalStorageFeature func()
	// NotifyMetricsServerFeature is a function that notifies the metrics server feature to reconcile.
	NotifyMetricsServerFeature func()
	// NotifyDNSFeature is a function that notifies the DNS feature to reconcile.
	NotifyDNSFeature func()
	// FeatureToReconciledCh is a map of feature names to channels that are full
	// when the feature controller has reconciled the feature.
	FeatureToReconciledCh map[types.FeatureName]<-chan struct{}
	// FeatureControllerReadyTimeout is the timeout for the feature controller to be ready.
	FeatureControllerReadyTimeout time.Duration
	// FeatureControllerReconcileTimeout is the timeout for each feature to get reconciled by the feature controller.
	FeatureControllerReconcileTimeout time.Duration
}

func NewController(
	logger logr.Logger,
	client client.Client,
	opts ControllerOptions,
) *Controller {
	return &Controller{
		logger:                            logger,
		client:                            client,
		featureControllerReadyCh:          opts.FeatureControllerReadyCh,
		notifyNetworkFeature:              opts.NotifyNetworkFeature,
		notifyGatewayFeature:              opts.NotifyGatewayFeature,
		notifyIngressFeature:              opts.NotifyIngressFeature,
		notifyLoadBalancerFeature:         opts.NotifyLoadBalancerFeature,
		notifyLocalStorageFeature:         opts.NotifyLocalStorageFeature,
		notifyMetricsServerFeature:        opts.NotifyMetricsServerFeature,
		notifyDNSFeature:                  opts.NotifyDNSFeature,
		featureToReconciledCh:             opts.FeatureToReconciledCh,
		featureControllerReadyTimeout:     opts.FeatureControllerReadyTimeout,
		featureControllerReconcileTimeout: opts.FeatureControllerReconcileTimeout,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&upgradesv1alpha1.Upgrade{}).
		For(&corev1.Node{}).
		WithOptions(controller.Options{
			// NOTE(Hue): We use a custom rate limiter to reduce the load on the API server,
			// as the default rate limiter is too aggressive for our use case (baseDelay is 5 Milliseconds).
			RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
				time.Second,
				5*time.Minute,
			),
		}).
		Complete(c)
}
