package features

import "path"

var (
	// featureDNS is manifests for the built-in featureDNS feature, powered by CoreDNS.
	featureDNS = feature{
		name:         "ck-dns",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "coredns-1.29.0.tgz"),
	}

	// featureLocalStorage is manifests for the built-in local storage feature, powered by Rawfile LocalPV CSI.
	featureLocalStorage = feature{
		name:         "ck-storage",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "rawfile-csi-0.8.0.tgz"),
	}

	// featureMetricsServer is manifests for the built-in metrics-server feature, powered by the upstream metrics-server.
	featureMetricsServer = feature{
		name:         "metrics-server",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "metrics-server-3.12.0.tgz"),
	}

	// featureNetwork is manifests for the built-in featureNetwork feature, powered by Cilium.
	featureNetwork = feature{
		name:         "ck-network",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "cilium-1.15.2.tgz"),
	}

	// featureLoadBalancer is manifests for the built-in featureLoadBalancer feature, powered by Cilium.
	featureLoadBalancer = feature{
		name:         "ck-loadbalancer",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "ck-loadbalancer"),
	}

	// featureGateway is manifests for the built-in featureGateway API feature, powered by Cilium.
	featureGateway = feature{
		name:         "ck-gateway",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "gateway-api-1.0.0.tgz"),
	}
)
