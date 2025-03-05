package dns

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

type DNSReconciler struct {
	features.Reconciler
}

func NewDNSReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.DNSReconciler {
	return DNSReconciler{
		Reconciler: features.NewReconciler(snap, helmClient, state),
	}
}
