package features

import (
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
var Implementation Interface = &implementation{
	applyDNS:           coredns_dns.ApplyDNS,
	applyNetwork:       cilium_network.ApplyNetwork,
	applyLoadBalancer:  metallb_loadbalancer.ApplyLoadBalancer,
	applyIngress:       cilium_ingress.ApplyIngress,
	applyGateway:       cilium_gateway.ApplyGateway,
	applyMetricsServer: metrics_server.ApplyMetricsServer,
	applyLocalStorage:  localpv_local_storage.ApplyLocalStorage,
}

// StatusChecks implements the Canonical Kubernetes built-in feature status checks.
var StatusChecks StatusInterface = &statusChecks{
	checkNetwork: cilium.CheckNetwork,
	checkDNS:     coredns.CheckDNS,
}

var Cleanup CleanupInterface = &cleanup{
	cleanupNetwork: cilium.CleanupNetwork,
}
