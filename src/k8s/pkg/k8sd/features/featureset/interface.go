package featureset

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

// implementation implements Interface.
type implementation struct {
	applyDNS           func(context.Context, snap.Snap, types.DNS, types.Kubelet, types.Annotations) (types.FeatureStatus, string, error)
	applyNetwork       func(context.Context, snap.Snap, state.State, types.APIServer, types.Network, types.Annotations) (types.FeatureStatus, error)
	applyLoadBalancer  func(context.Context, snap.Snap, types.LoadBalancer, types.Network, types.Annotations) (types.FeatureStatus, error)
	applyIngress       func(context.Context, snap.Snap, types.Ingress, types.Network, types.Annotations) (types.FeatureStatus, error)
	applyGateway       func(context.Context, snap.Snap, types.Gateway, types.Network, types.Annotations) (types.FeatureStatus, error)
	applyMetricsServer func(context.Context, snap.Snap, types.MetricsServer, types.Annotations) (types.FeatureStatus, error)
	applyLocalStorage  func(context.Context, snap.Snap, types.LocalStorage, types.Annotations) (types.FeatureStatus, error)
}

func (i *implementation) ApplyDNS(ctx context.Context, snap snap.Snap, dns types.DNS, kubelet types.Kubelet, annotations types.Annotations) (types.FeatureStatus, string, error) {
	return i.applyDNS(ctx, snap, dns, kubelet, annotations)
}

func (i *implementation) ApplyNetwork(ctx context.Context, snap snap.Snap, s state.State, apiserver types.APIServer, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyNetwork(ctx, snap, s, apiserver, network, annotations)
}

func (i *implementation) ApplyLoadBalancer(ctx context.Context, snap snap.Snap, loadbalancer types.LoadBalancer, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyLoadBalancer(ctx, snap, loadbalancer, network, annotations)
}

func (i *implementation) ApplyIngress(ctx context.Context, snap snap.Snap, ingress types.Ingress, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyIngress(ctx, snap, ingress, network, annotations)
}

func (i *implementation) ApplyGateway(ctx context.Context, snap snap.Snap, gateway types.Gateway, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyGateway(ctx, snap, gateway, network, annotations)
}

func (i *implementation) ApplyMetricsServer(ctx context.Context, snap snap.Snap, cfg types.MetricsServer, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyMetricsServer(ctx, snap, cfg, annotations)
}

func (i *implementation) ApplyLocalStorage(ctx context.Context, snap snap.Snap, cfg types.LocalStorage, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyLocalStorage(ctx, snap, cfg, annotations)
}
