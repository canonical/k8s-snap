//go:build ignore

package features

import (
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
var Implementation Interface = &implementation{
	applyDNS:           coredns_dns.ApplyDNS,
	applyNetwork:       calico_network.ApplyNetwork,
	applyLoadBalancer:  metallb_loadbalancer.ApplyLoadBalancer,
	applyIngress:       contour_ingress.ApplyIngress,
	applyGateway:       contour_gateway.ApplyGateway,
	applyMetricsServer: metrics_server.ApplyMetricsServer,
	applyLocalStorage:  localpv_local_storage.ApplyLocalStorage,
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
