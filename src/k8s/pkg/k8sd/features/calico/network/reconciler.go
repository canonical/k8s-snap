package network

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

type NetworkReconciler struct {
	features.Reconciler
}

// NewNetworkReconciler returns a new instance of NetworkReconciler.
func NewNetworkReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.NetworkReconciler {
	return NetworkReconciler{
		Reconciler: features.NewReconciler(snap, helmClient, state),
	}
}
