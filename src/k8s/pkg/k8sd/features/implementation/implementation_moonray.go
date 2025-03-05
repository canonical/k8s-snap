//go:build ignore

package implementation

import (
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/features/calico"
	calico_network "github.com/canonical/k8s/pkg/k8sd/features/calico/network"
	contour_gateway "github.com/canonical/k8s/pkg/k8sd/features/contour/gateway"
	contour_ingress "github.com/canonical/k8s/pkg/k8sd/features/contour/ingress"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	coredns_dns "github.com/canonical/k8s/pkg/k8sd/features/coredns/dns"
	localpv_local_storage "github.com/canonical/k8s/pkg/k8sd/features/localpv/local-storage"
	metallb_loadbalancer "github.com/canonical/k8s/pkg/k8sd/features/metallb/loadbalancer"
	metrics_server "github.com/canonical/k8s/pkg/k8sd/features/metrics-server"
)

// Implementation contains the moonray features for Canonical Kubernetes.
// TODO: Replace default by moonray.
var Implementation features.Interface = &implementation{
	newDNSReconciler:           coredns_dns.NewDNSReconciler,
	newNetworkReconciler:       calico_network.NewNetworkReconciler,
	newLoadBalancerReconciler:  metallb_loadbalancer.NewLoadBalancerReconciler,
	newIngressReconciler:       contour_ingress.NewIngressReconciler,
	newGatewayReconciler:       contour_gateway.NewGatewayReconciler,
	newMetricsServerReconciler: metrics_server.NewMetricsServerReconciler,
	newLocalStorageReconciler:  localpv_local_storage.NewLocalStorageReconciler,
}

// StatusChecks implements the Canonical Kubernetes moonray feature status checks.
// TODO: Replace default by moonray.
var StatusChecks StatusInterface = &statusChecks{
	checkNetwork: calico.CheckNetwork,
	checkDNS:     coredns.CheckDNS,
}

var Cleanup CleanupInterface = &cleanup{
	cleanupNetwork: calico.CleanupNetwork,
}
