package features

import (
	"context"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

type Reconciler struct {
	snap       snap.Snap
	helmClient helm.Client
	state      state.State
}

func NewReconciler(snap snap.Snap, helmClient helm.Client, state state.State) Reconciler {
	return Reconciler{
		snap:       snap,
		helmClient: helmClient,
		state:      state,
	}
}

func (r *Reconciler) Snap() snap.Snap {
	return r.snap
}

func (r *Reconciler) HelmClient() helm.Client {
	return r.helmClient
}

func (r *Reconciler) State() state.State {
	return r.state
}

type DNSReconciler interface {
	// ApplyDNS is used to configure the DNS feature on Canonical Kubernetes.
	ApplyDNS(context.Context, types.DNS, types.Kubelet, types.Annotations) (types.FeatureStatus, string, error)
}

type NetworkReconciler interface {
	// ApplyNetwork is used to configure the network feature on Canonical Kubernetes.
	ApplyNetwork(context.Context, types.APIServer, types.Network, types.Annotations) (types.FeatureStatus, error)
}

type LoadBalancerReconciler interface {
	// ApplyLoadBalancer is used to configure the load-balancer feature on Canonical Kubernetes.
	ApplyLoadBalancer(context.Context, types.LoadBalancer, types.Network, types.Annotations) (types.FeatureStatus, error)
}

type IngressReconciler interface {
	// ApplyIngress is used to configure the ingress controller feature on Canonical Kubernetes.
	ApplyIngress(context.Context, types.Ingress, types.Network, types.Annotations) (types.FeatureStatus, error)
}

type GatewayReconciler interface {
	// ApplyGateway is used to configure the gateway feature on Canonical Kubernetes.
	ApplyGateway(context.Context, types.Gateway, types.Network, types.Annotations) (types.FeatureStatus, error)
}

type MetricsServerReconciler interface {
	// ApplyMetricsServer is used to configure the metrics-server feature on Canonical Kubernetes.
	ApplyMetricsServer(context.Context, types.MetricsServer, types.Annotations) (types.FeatureStatus, error)
}

type LocalStorageReconciler interface {
	// ApplyLocalStorage is used to configure the Local Storage feature on Canonical Kubernetes.
	ApplyLocalStorage(context.Context, types.LocalStorage, types.Annotations) (types.FeatureStatus, error)
}

type Interface interface {
	NewDNSReconciler(snap snap.Snap, helmClient helm.Client, state state.State) DNSReconciler
	NewNetworkReconciler(snap snap.Snap, helmClient helm.Client, state state.State) NetworkReconciler
	NewLoadBalancerReconciler(snap snap.Snap, helmClient helm.Client, state state.State) LoadBalancerReconciler
	NewIngressReconciler(snap snap.Snap, helmClient helm.Client, state state.State) IngressReconciler
	NewGatewayReconciler(snap snap.Snap, helmClient helm.Client, state state.State) GatewayReconciler
	NewMetricsServerReconciler(snap snap.Snap, helmClient helm.Client, state state.State) MetricsServerReconciler
	NewLocalStorageReconciler(snap snap.Snap, helmClient helm.Client, state state.State) LocalStorageReconciler
}
