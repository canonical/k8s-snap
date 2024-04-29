package features

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chartCoreDNS is manifests for the built-in DNS feature, powered by CoreDNS.
	chartCoreDNS = helm.InstallableChart{
		Name:         "ck-dns",
		Namespace:    "kube-system",
		ManifestPath: path.Join("charts", "coredns-1.29.0.tgz"),
	}

	// chartCilium is manifests for the built-in CNI feature, powered by Cilium.
	chartCilium = helm.InstallableChart{
		Name:         "ck-network",
		Namespace:    "kube-system",
		ManifestPath: path.Join("charts", "cilium-1.15.2.tgz"),
	}

	// chartCiliumLoadBalancer is manifests for the built-in load-balancer feature, powered by Cilium.
	chartCiliumLoadBalancer = helm.InstallableChart{
		Name:         "ck-loadbalancer",
		Namespace:    "kube-system",
		ManifestPath: path.Join("charts", "ck-loadbalancer"),
	}

	// chartCiliumGateway is manifests for the built-in gateway feature, powered by Cilium.
	chartCiliumGateway = helm.InstallableChart{
		Name:         "ck-gateway",
		Namespace:    "kube-system",
		ManifestPath: path.Join("charts", "gateway-api-1.0.0.tgz"),
	}

	// chartLocalStorage is manifests for the built-in local storage feature, powered by Rawfile LocalPV CSI.
	chartLocalStorage = helm.InstallableChart{
		Name:         "ck-storage",
		Namespace:    "kube-system",
		ManifestPath: path.Join("charts", "rawfile-csi-0.8.0.tgz"),
	}

	// chartMetricsServer is manifests for the built-in metrics-server feature, powered by the upstream metrics-server.
	chartMetricsServer = helm.InstallableChart{
		Name:         "metrics-server",
		Namespace:    "kube-system",
		ManifestPath: path.Join("charts", "metrics-server-3.12.0.tgz"),
	}
)
