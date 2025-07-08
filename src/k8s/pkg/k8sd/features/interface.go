package features

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

// Interface abstracts the management of built-in Canonical Kubernetes features.
type Interface interface {
	// ApplyCilium is used to configure the Cilium components on Canonical Kubernetes.
	ApplyCilium(context.Context, snap.Snap, state.State, types.APIServer, types.Network, types.Gateway, types.Ingress, types.Annotations) (map[types.FeatureName]types.FeatureStatus, error)
	// ApplyDNS is used to configure the DNS feature on Canonical Kubernetes.
	ApplyDNS(context.Context, snap.Snap, types.DNS, types.Kubelet, types.Annotations) (types.FeatureStatus, string, error)
	// ApplyLoadBalancer is used to configure the load-balancer feature on Canonical Kubernetes.
	ApplyLoadBalancer(context.Context, snap.Snap, types.LoadBalancer, types.Network, types.Annotations) (types.FeatureStatus, error)
	// ApplyMetricsServer is used to configure the metrics-server feature on Canonical Kubernetes.
	ApplyMetricsServer(context.Context, snap.Snap, types.MetricsServer, types.Annotations) (types.FeatureStatus, error)
	// ApplyLocalStorage is used to configure the Local Storage feature on Canonical Kubernetes.
	ApplyLocalStorage(context.Context, snap.Snap, types.LocalStorage, types.Annotations) (types.FeatureStatus, error)
}

// implementation implements Interface.
type implementation struct {
	applyCilium        func(context.Context, snap.Snap, state.State, types.APIServer, types.Network, types.Gateway, types.Ingress, types.Annotations) (map[types.FeatureName]types.FeatureStatus, error)
	applyDNS           func(context.Context, snap.Snap, types.DNS, types.Kubelet, types.Annotations) (types.FeatureStatus, string, error)
	applyLoadBalancer  func(context.Context, snap.Snap, types.LoadBalancer, types.Network, types.Annotations) (types.FeatureStatus, error)
	applyMetricsServer func(context.Context, snap.Snap, types.MetricsServer, types.Annotations) (types.FeatureStatus, error)
	applyLocalStorage  func(context.Context, snap.Snap, types.LocalStorage, types.Annotations) (types.FeatureStatus, error)
}

func (i *implementation) ApplyCilium(ctx context.Context, snap snap.Snap, s state.State, apiServer types.APIServer, network types.Network, gateway types.Gateway, ingress types.Ingress, annotations types.Annotations) (map[types.FeatureName]types.FeatureStatus, error) {
	return i.applyCilium(ctx, snap, s, apiServer, network, gateway, ingress, annotations)
}

func (i *implementation) ApplyDNS(ctx context.Context, snap snap.Snap, dns types.DNS, kubelet types.Kubelet, annotations types.Annotations) (types.FeatureStatus, string, error) {
	return i.applyDNS(ctx, snap, dns, kubelet, annotations)
}

func (i *implementation) ApplyLoadBalancer(ctx context.Context, snap snap.Snap, loadbalancer types.LoadBalancer, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyLoadBalancer(ctx, snap, loadbalancer, network, annotations)
}

func (i *implementation) ApplyMetricsServer(ctx context.Context, snap snap.Snap, cfg types.MetricsServer, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyMetricsServer(ctx, snap, cfg, annotations)
}

func (i *implementation) ApplyLocalStorage(ctx context.Context, snap snap.Snap, cfg types.LocalStorage, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyLocalStorage(ctx, snap, cfg, annotations)
}
