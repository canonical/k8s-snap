package implementation

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

type implementation struct {
	newDNSReconciler           func(snap snap.Snap, helmClient helm.Client, state state.State) features.DNSReconciler
	newNetworkReconciler       func(snap snap.Snap, helmClient helm.Client, state state.State) features.NetworkReconciler
	newLoadBalancerReconciler  func(snap snap.Snap, helmClient helm.Client, state state.State) features.LoadBalancerReconciler
	newIngressReconciler       func(snap snap.Snap, helmClient helm.Client, state state.State) features.IngressReconciler
	newGatewayReconciler       func(snap snap.Snap, helmClient helm.Client, state state.State) features.GatewayReconciler
	newMetricsServerReconciler func(snap snap.Snap, helmClient helm.Client, state state.State) features.MetricsServerReconciler
	newLocalStorageReconciler  func(snap snap.Snap, helmClient helm.Client, state state.State) features.LocalStorageReconciler
}

func (i *implementation) NewDNSReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.DNSReconciler {
	return i.newDNSReconciler(snap, helmClient, state)
}

func (i *implementation) NewNetworkReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.NetworkReconciler {
	return i.newNetworkReconciler(snap, helmClient, state)
}

func (i *implementation) NewLoadBalancerReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.LoadBalancerReconciler {
	return i.newLoadBalancerReconciler(snap, helmClient, state)
}

func (i *implementation) NewIngressReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.IngressReconciler {
	return i.newIngressReconciler(snap, helmClient, state)
}

func (i *implementation) NewGatewayReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.GatewayReconciler {
	return i.newGatewayReconciler(snap, helmClient, state)
}

func (i *implementation) NewMetricsServerReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.MetricsServerReconciler {
	return i.newMetricsServerReconciler(snap, helmClient, state)
}

func (i *implementation) NewLocalStorageReconciler(snap snap.Snap, helmClient helm.Client, state state.State) features.LocalStorageReconciler {
	return i.newLocalStorageReconciler(snap, helmClient, state)
}
