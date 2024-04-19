package features

import "path"

var (
	// featureCoreDNS is manifests for the built-in DNS feature, powered by CoreDNS.
	featureCoreDNS = Feature{
		name:         "ck-dns",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "coredns-1.29.0.tgz"),
	}

	// featureCiliumCNI is manifests for the built-in CNI feature, powered by Cilium.
	featureCiliumCNI = Feature{
		name:         "ck-network",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "cilium-1.15.2.tgz"),
	}

	// featureCiliumLoadBalancer is manifests for the built-in load-balancer feature, powered by Cilium.
	featureCiliumLoadBalancer = Feature{
		name:         "ck-loadbalancer",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "ck-loadbalancer"),
	}

	// featureCiliumGateway is manifests for the built-in gateway feature, powered by Cilium.
	featureCiliumGateway = Feature{
		name:         "ck-gateway",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "gateway-api-1.0.0.tgz"),
	}

	// featureLocalStorage is manifests for the built-in local storage feature, powered by Rawfile LocalPV CSI.
	featureLocalStorage = Feature{
		name:         "ck-storage",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "rawfile-csi-0.8.0.tgz"),
	}

	// featureMetricsServer is manifests for the built-in metrics-server feature, powered by the upstream metrics-server.
	featureMetricsServer = Feature{
		name:         "metrics-server",
		namespace:    "kube-system",
		manifestPath: path.Join("charts", "metrics-server-3.12.0.tgz"),
	}
)
