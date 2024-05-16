package features

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// Interface abstracts the management of built-in Canonical Kubernetes features.
type Interface interface {
	// ApplyDNS is used to configure the DNS feature on Canonical Kubernetes.
	ApplyDNS(context.Context, snap.Snap, types.DNS, types.Kubelet) (string, error)
	// ApplyDNS is used to configure the network feature on Canonical Kubernetes.
	ApplyNetwork(context.Context, snap.Snap, types.Network) error
	// ApplyLoadBalancer is used to configure the load-balancer feature on Canonical Kubernetes.
	ApplyLoadBalancer(context.Context, snap.Snap, types.LoadBalancer, types.Network) error
	// ApplyIngress is used to configure the ingress controller feature on Canonical Kubernetes.
	ApplyIngress(context.Context, snap.Snap, types.Ingress, types.Network) error
	// ApplyGateway is used to configure the gateway feature on Canonical Kubernetes.
	ApplyGateway(context.Context, snap.Snap, types.Gateway, types.Network) error
	// ApplyMetricsServer is used to configure the metrics-server feature on Canonical Kubernetes.
	ApplyMetricsServer(context.Context, snap.Snap, types.MetricsServer) error
	// ApplyLocalStorage is used to configure the Local Storage feature on Canonical Kubernetes.
	ApplyLocalStorage(context.Context, snap.Snap, types.LocalStorage) error
}

// implementation implements Interface.
type implementation struct {
	applyDNS           func(context.Context, snap.Snap, types.DNS, types.Kubelet) (string, error)
	applyNetwork       func(context.Context, snap.Snap, types.Network) error
	applyLoadBalancer  func(context.Context, snap.Snap, types.LoadBalancer, types.Network) error
	applyIngress       func(context.Context, snap.Snap, types.Ingress, types.Network) error
	applyGateway       func(context.Context, snap.Snap, types.Gateway, types.Network) error
	applyMetricsServer func(context.Context, snap.Snap, types.MetricsServer) error
	applyLocalStorage  func(context.Context, snap.Snap, types.LocalStorage) error
}

func (i *implementation) ApplyDNS(ctx context.Context, snap snap.Snap, dns types.DNS, kubelet types.Kubelet) (string, error) {
	return i.applyDNS(ctx, snap, dns, kubelet)
}

func (i *implementation) ApplyNetwork(ctx context.Context, snap snap.Snap, cfg types.Network) error {
	return i.applyNetwork(ctx, snap, cfg)
}

func (i *implementation) ApplyLoadBalancer(ctx context.Context, snap snap.Snap, loadbalancer types.LoadBalancer, network types.Network) error {
	return i.applyLoadBalancer(ctx, snap, loadbalancer, network)
}

func (i *implementation) ApplyIngress(ctx context.Context, snap snap.Snap, ingress types.Ingress, network types.Network) error {
	return i.applyIngress(ctx, snap, ingress, network)
}

func (i *implementation) ApplyGateway(ctx context.Context, snap snap.Snap, gateway types.Gateway, network types.Network) error {
	return i.applyGateway(ctx, snap, gateway, network)
}

func (i *implementation) ApplyMetricsServer(ctx context.Context, snap snap.Snap, cfg types.MetricsServer) error {
	return i.applyMetricsServer(ctx, snap, cfg)
}

func (i *implementation) ApplyLocalStorage(ctx context.Context, snap snap.Snap, cfg types.LocalStorage) error {
	return i.applyLocalStorage(ctx, snap, cfg)
}
