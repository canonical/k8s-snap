package metrics_server

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

type MetricsServerReconciler struct {
	features.Reconciler
}

// NewMetricsServerReconciler returns a new instance of MetricsServerReconciler.
func NewMetricsServerReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.MetricsServerReconciler {
	return MetricsServerReconciler{
		Reconciler: features.NewReconciler(snap, helmClient, state),
	}
}
