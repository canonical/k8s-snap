package dns

import (
	"github.com/canonical/k8s/pkg/k8sd/features"
)

type reconciler struct {
	features.BaseReconciler
}

func NewReconciler(base features.BaseReconciler) features.Reconciler {
	return reconciler{
		base,
	}
}
