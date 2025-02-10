package features

import (
	"context"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

// Interface abstracts the management of built-in Canonical Kubernetes features.
type Interface interface {
	// ApplyDNS is used to configure the DNS feature on Canonical Kubernetes.
	ApplyDNS(context.Context, snap.Snap, helm.Client, types.DNS, types.Kubelet, types.Annotations) (types.FeatureStatus, string, error)
	// ApplyNetwork is used to configure the network feature on Canonical Kubernetes.
	ApplyNetwork(context.Context, snap.Snap, helm.Client, state.State, types.APIServer, types.Network, types.Annotations) (types.FeatureStatus, error)
	// ApplyLoadBalancer is used to configure the load-balancer feature on Canonical Kubernetes.
	ApplyLoadBalancer(context.Context, snap.Snap, helm.Client, types.LoadBalancer, types.Network, types.Annotations) (types.FeatureStatus, error)
	// ApplyIngress is used to configure the ingress controller feature on Canonical Kubernetes.
	ApplyIngress(context.Context, snap.Snap, helm.Client, types.Ingress, types.Network, types.Annotations) (types.FeatureStatus, error)
	// ApplyGateway is used to configure the gateway feature on Canonical Kubernetes.
	ApplyGateway(context.Context, snap.Snap, helm.Client, types.Gateway, types.Network, types.Annotations) (types.FeatureStatus, error)
	// ApplyMetricsServer is used to configure the metrics-server feature on Canonical Kubernetes.
	ApplyMetricsServer(context.Context, snap.Snap, helm.Client, types.MetricsServer, types.Annotations) (types.FeatureStatus, error)
	// ApplyLocalStorage is used to configure the Local Storage feature on Canonical Kubernetes.
	ApplyLocalStorage(context.Context, snap.Snap, helm.Client, types.LocalStorage, types.Annotations) (types.FeatureStatus, error)
}

// implementation implements Interface.
type implementation struct {
	applyDNS           func(context.Context, snap.Snap, helm.Client, types.DNS, types.Kubelet, types.Annotations) (types.FeatureStatus, string, error)
	applyNetwork       func(context.Context, snap.Snap, helm.Client, state.State, types.APIServer, types.Network, types.Annotations) (types.FeatureStatus, error)
	applyLoadBalancer  func(context.Context, snap.Snap, helm.Client, types.LoadBalancer, types.Network, types.Annotations) (types.FeatureStatus, error)
	applyIngress       func(context.Context, snap.Snap, helm.Client, types.Ingress, types.Network, types.Annotations) (types.FeatureStatus, error)
	applyGateway       func(context.Context, snap.Snap, helm.Client, types.Gateway, types.Network, types.Annotations) (types.FeatureStatus, error)
	applyMetricsServer func(context.Context, snap.Snap, helm.Client, types.MetricsServer, types.Annotations) (types.FeatureStatus, error)
	applyLocalStorage  func(context.Context, snap.Snap, helm.Client, types.LocalStorage, types.Annotations) (types.FeatureStatus, error)
}

func (i *implementation) ApplyDNS(ctx context.Context, snap snap.Snap, client helm.Client, dns types.DNS, kubelet types.Kubelet, annotations types.Annotations) (types.FeatureStatus, string, error) {
	return i.applyDNS(ctx, snap, client, dns, kubelet, annotations)
}

func (i *implementation) ApplyNetwork(ctx context.Context, snap snap.Snap, client helm.Client, s state.State, apiserver types.APIServer, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyNetwork(ctx, snap, client, s, apiserver, network, annotations)
}

func (i *implementation) ApplyLoadBalancer(ctx context.Context, snap snap.Snap, client helm.Client, loadbalancer types.LoadBalancer, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyLoadBalancer(ctx, snap, client, loadbalancer, network, annotations)
}

func (i *implementation) ApplyIngress(ctx context.Context, snap snap.Snap, client helm.Client, ingress types.Ingress, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyIngress(ctx, snap, client, ingress, network, annotations)
}

func (i *implementation) ApplyGateway(ctx context.Context, snap snap.Snap, client helm.Client, gateway types.Gateway, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyGateway(ctx, snap, client, gateway, network, annotations)
}

func (i *implementation) ApplyMetricsServer(ctx context.Context, snap snap.Snap, client helm.Client, cfg types.MetricsServer, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyMetricsServer(ctx, snap, client, cfg, annotations)
}

func (i *implementation) ApplyLocalStorage(ctx context.Context, snap snap.Snap, client helm.Client, cfg types.LocalStorage, annotations types.Annotations) (types.FeatureStatus, error) {
	return i.applyLocalStorage(ctx, snap, client, cfg, annotations)
}
