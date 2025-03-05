package gateway

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

type GatewayReconciler struct {
	features.Reconciler
}

// NewGatewayReconciler returns a new instance of GatewayReconciler.
func NewGatewayReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.GatewayReconciler {
	return GatewayReconciler{
		Reconciler: features.NewReconciler(snap, helmClient, state),
	}
}
