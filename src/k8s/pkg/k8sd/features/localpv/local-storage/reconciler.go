package local_storage

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

type LocalStorageReconciler struct {
	features.Reconciler
}

// NewLocalStorageReconciler returns a new instance of LocalStorageReconciler.
func NewLocalStorageReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.LocalStorageReconciler {
	return LocalStorageReconciler{
		Reconciler: features.NewReconciler(snap, helmClient, state),
	}
}
