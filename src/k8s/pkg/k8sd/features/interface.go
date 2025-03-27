package features

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

// ReconcileInterface abstracts the management of built-in Canonical Kubernetes features.
type ReconcileInterface interface {
	// ApplyDNS is used to configure the DNS feature on Canonical Kubernetes.
	ApplyDNS(context.Context, state.State, snap.Snap, types.DNS, types.Kubelet, types.Annotations) (types.FeatureStatus, string, error)
	// ApplyNetwork is used to configure the network feature on Canonical Kubernetes.
	ApplyNetwork(context.Context, state.State, snap.Snap, types.APIServer, types.Network, types.Annotations) (types.FeatureStatus, error)
	// ApplyLoadBalancer is used to configure the load-balancer feature on Canonical Kubernetes.
	ApplyLoadBalancer(context.Context, state.State, snap.Snap, types.LoadBalancer, types.Network, types.Annotations) (types.FeatureStatus, error)
	// ApplyIngress is used to configure the ingress controller feature on Canonical Kubernetes.
	ApplyIngress(context.Context, state.State, snap.Snap, types.Ingress, types.Network, types.Annotations) (types.FeatureStatus, error)
	// ApplyGateway is used to configure the gateway feature on Canonical Kubernetes.
	ApplyGateway(context.Context, state.State, snap.Snap, types.Gateway, types.Network, types.Annotations) (types.FeatureStatus, error)
	// ApplyMetricsServer is used to configure the metrics-server feature on Canonical Kubernetes.
	ApplyMetricsServer(context.Context, state.State, snap.Snap, types.MetricsServer, types.Annotations) (types.FeatureStatus, error)
	// ApplyLocalStorage is used to configure the Local Storage feature on Canonical Kubernetes.
	ApplyLocalStorage(context.Context, state.State, snap.Snap, types.LocalStorage, types.Annotations) (types.FeatureStatus, error)
}
