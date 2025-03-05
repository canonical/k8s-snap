package ingress

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

type IngressReconciler struct {
	features.Reconciler
}

// NewIngressReconciler returns a new instance of IngressReconciler.
func NewIngressReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.IngressReconciler {
	return IngressReconciler{
		Reconciler: features.NewReconciler(snap, helmClient, state),
	}
}
