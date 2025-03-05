package loadbalancer

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

type LoadBalancerReconciler struct {
	features.Reconciler
}

// NewLoadBalancerReconciler returns a new instance of LoadBalancerReconciler.
func NewLoadBalancerReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.LoadBalancerReconciler {
	return LoadBalancerReconciler{
		Reconciler: features.NewReconciler(snap, helmClient, state),
	}
}
