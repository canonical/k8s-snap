package upgrade

import (
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// upgradeReconciler is a controller that reconciles the upgrade custom resource.
type upgradeReconciler struct {
	// getState is a function that returns the Microcluster state.
	getState func() state.State
	// snap is the snap instance.
	snap snap.Snap
	// featureControllerReadyTimeout is the timeout for the feature controller to be ready.
	featureControllerReadyTimeout time.Duration
	// featureControllerReconcileTimeout is the timeout for the feature controller to reconcile.
	featureControllerReconcileTimeout time.Duration
	// featureControllerReadyCh is a channel that is closed when the feature controller is ready.
	featureControllerReadyCh <-chan struct{}
	// notifyFeatureController is a function that notifies the feature controller to reconcile.
	notifyFeatureController func()
	// featureToReconciledCh is a map of feature names to channels that are full
	// when the feature controller has reconciled the feature.
	featureToReconciledCh map[string]<-chan struct{}

	Manager manager.Manager
	Logger  logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *upgradeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubernetes.Upgrade{}).
		Complete(r)
}
