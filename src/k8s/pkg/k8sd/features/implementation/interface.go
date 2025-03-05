package implementation

import (
	"github.com/canonical/k8s/pkg/k8sd/features"
)

type implementation struct {
	newDNSReconciler           func(base features.BaseReconciler) features.Reconciler
	newNetworkReconciler       func(base features.BaseReconciler) features.Reconciler
	newLoadBalancerReconciler  func(base features.BaseReconciler) features.Reconciler
	newIngressReconciler       func(base features.BaseReconciler) features.Reconciler
	newGatewayReconciler       func(base features.BaseReconciler) features.Reconciler
	newMetricsServerReconciler func(base features.BaseReconciler) features.Reconciler
	newLocalStorageReconciler  func(base features.BaseReconciler) features.Reconciler
}

func (i *implementation) NewDNSReconciler(base features.BaseReconciler) features.Reconciler {
	return i.newDNSReconciler(base)
}

func (i *implementation) NewNetworkReconciler(base features.BaseReconciler) features.Reconciler {
	return i.newNetworkReconciler(base)
}

func (i *implementation) NewLoadBalancerReconciler(base features.BaseReconciler) features.Reconciler {
	return i.newLoadBalancerReconciler(base)
}

func (i *implementation) NewIngressReconciler(base features.BaseReconciler) features.Reconciler {
	return i.newIngressReconciler(base)
}

func (i *implementation) NewGatewayReconciler(base features.BaseReconciler) features.Reconciler {
	return i.newGatewayReconciler(base)
}

func (i *implementation) NewMetricsServerReconciler(base features.BaseReconciler) features.Reconciler {
	return i.newMetricsServerReconciler(base)
}

func (i *implementation) NewLocalStorageReconciler(base features.BaseReconciler) features.Reconciler {
	return i.newLocalStorageReconciler(base)
}
