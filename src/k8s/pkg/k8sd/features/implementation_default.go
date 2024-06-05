package features

import (
	"github.com/canonical/k8s/pkg/k8sd/features/contour"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	"github.com/canonical/k8s/pkg/k8sd/features/fake"
	"github.com/canonical/k8s/pkg/k8sd/features/localpv"
	metrics_server "github.com/canonical/k8s/pkg/k8sd/features/metrics-server"
)

// Default implements the Canonical Kubernetes built-in features.
// Cilium is used for networking (network + load-balancer + ingress + gateway).
// CoreDNS is used for DNS.
// MetricsServer is used for metrics-server.
// LocalPV Rawfile CSI is used for local-storage.
var Implementation Interface = &implementation{
	applyDNS:           coredns.ApplyDNS,
	applyNetwork:       fake.ApplyNetwork, //TODO: remove default overwrite for testing
	applyLoadBalancer:  fake.ApplyLoadBalancer,
	applyIngress:       contour.ApplyIngress,
	applyGateway:       fake.ApplyGateway,
	applyMetricsServer: metrics_server.ApplyMetricsServer,
	applyLocalStorage:  localpv.ApplyLocalStorage,
}

// StatusChecks implements the Canonical Kubernetes built-in feature status checks.
var StatusChecks StatusInterface = &statusChecks{
	checkNetwork: fake.CheckNetwork,
	checkDNS:     coredns.CheckDNS,
}
