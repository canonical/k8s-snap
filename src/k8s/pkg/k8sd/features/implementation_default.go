//go:build ignore

package features

import (
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	"github.com/canonical/k8s/pkg/k8sd/features/localpv"
	"github.com/canonical/k8s/pkg/k8sd/features/metallb"
	metrics_server "github.com/canonical/k8s/pkg/k8sd/features/metrics-server"
)

// Default implements the Canonical Kubernetes built-in features.
// Cilium is used for networking (network + ingress + gateway).
// MetalLB is used for LoadBalancer.
// CoreDNS is used for DNS.
// MetricsServer is used for metrics-server.
// LocalPV Rawfile CSI is used for local-storage.
var Implementation Interface = &implementation{
	applyDNS:           coredns.ApplyDNS,
	applyNetwork:       cilium.ApplyNetwork,
	applyLoadBalancer:  metallb.ApplyLoadBalancer,
	applyIngress:       cilium.ApplyIngress,
	applyGateway:       cilium.ApplyGateway,
	applyMetricsServer: metrics_server.ApplyMetricsServer,
	applyLocalStorage:  localpv.ApplyLocalStorage,
}

// StatusChecks implements the Canonical Kubernetes built-in feature status checks.
var StatusChecks StatusInterface = &statusChecks{
	checkNetwork: cilium.CheckNetwork,
	checkDNS:     coredns.CheckDNS,
}

var Cleanup CleanupInterface = &cleanup{
	cleanupNetwork: cilium.CleanupNetwork,
}
