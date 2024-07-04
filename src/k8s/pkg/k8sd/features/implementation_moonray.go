package features

import (
	"github.com/canonical/k8s/pkg/k8sd/features/calico"
	"github.com/canonical/k8s/pkg/k8sd/features/contour"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	"github.com/canonical/k8s/pkg/k8sd/features/localpv"
	"github.com/canonical/k8s/pkg/k8sd/features/metallb"
	metrics_server "github.com/canonical/k8s/pkg/k8sd/features/metrics-server"
)

// Implementation contains the moonray features for Canonical Kubernetes.
// TODO: Replace default by moonray.
var Implementation Interface = &implementation{
	applyDNS:           coredns.ApplyDNS,
	applyNetwork:       calico.ApplyNetwork,
	applyLoadBalancer:  metallb.ApplyLoadBalancer,
	applyIngress:       contour.ApplyIngress,
	applyGateway:       contour.ApplyGateway,
	applyMetricsServer: metrics_server.ApplyMetricsServer,
	applyLocalStorage:  localpv.ApplyLocalStorage,
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
