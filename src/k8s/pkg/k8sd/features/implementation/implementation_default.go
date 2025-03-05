package implementation

import (
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	cilium_gateway "github.com/canonical/k8s/pkg/k8sd/features/cilium/gateway"
	cilium_ingress "github.com/canonical/k8s/pkg/k8sd/features/cilium/ingress"
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	coredns_dns "github.com/canonical/k8s/pkg/k8sd/features/coredns/dns"
	localpv_local_storage "github.com/canonical/k8s/pkg/k8sd/features/localpv/local-storage"
	metallb_loadbalancer "github.com/canonical/k8s/pkg/k8sd/features/metallb/loadbalancer"
	metrics_server "github.com/canonical/k8s/pkg/k8sd/features/metrics-server"
)

// Default implements the Canonical Kubernetes built-in features.
// Cilium is used for networking (network + ingress + gateway).
// MetalLB is used for LoadBalancer.
// CoreDNS is used for DNS.
// MetricsServer is used for metrics-server.
// LocalPV Rawfile CSI is used for local-storage.
var Implementation features.Interface = &implementation{
	newDNSReconciler:           coredns_dns.NewReconciler,
	newNetworkReconciler:       cilium_network.NewReconciler,
	newLoadBalancerReconciler:  metallb_loadbalancer.NewReconciler,
	newIngressReconciler:       cilium_ingress.NewReconciler,
	newGatewayReconciler:       cilium_gateway.NewReconciler,
	newMetricsServerReconciler: metrics_server.NewReconciler,
	newLocalStorageReconciler:  localpv_local_storage.NewReconciler,
}

// StatusChecks implements the Canonical Kubernetes built-in feature status checks.
var StatusChecks features.StatusInterface = &statusChecks{
	checkNetwork: cilium.CheckNetwork,
	checkDNS:     coredns.CheckDNS,
}

var Cleanup features.CleanupInterface = &cleanup{
	cleanupNetwork: cilium.CleanupNetwork,
}
