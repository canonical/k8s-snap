//go:build ignore

package features

import (
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	"github.com/canonical/k8s/pkg/k8sd/features/localpv"
	metrics_server "github.com/canonical/k8s/pkg/k8sd/features/metrics-server"
)

// Implementation contains the moonray features for Canonical Kubernetes.
// TODO: Replace default by moonray.
var Implementation Interface = &implementation{
	applyDNS:           coredns.ApplyDNS,
	applyNetwork:       cilium.ApplyNetwork,
	applyLoadBalancer:  cilium.ApplyLoadBalancer,
	applyIngress:       cilium.ApplyIngress,
	applyGateway:       cilium.ApplyGateway,
	applyMetricsServer: metrics_server.ApplyMetricsServer,
	applyLocalStorage:  localpv.ApplyLocalStorage,
}

// StatusChecks implements the Canonical Kubernetes moonray feature status checks.
// TODO: Replace default by moonray.
var StatusChecks StatusInterface = &statusChecks{
	checkNetwork: cilium.CheckNetwork,
	checkDNS:     coredns.CheckDNS,
}
