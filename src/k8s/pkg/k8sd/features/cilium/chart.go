package cilium

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
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

	// ciliumAgentImageRepo is the image to use for cilium-agent.
	ciliumAgentImageRepo = "ghcr.io/canonical/cilium"

	// agentImageTag is the tag to use for the cilium-agent image.
	ciliumAgentImageTag = "1.15.2-ck1"

	// ciliumOperatorImageRepo is the image to use for cilium-operator.
	ciliumOperatorImageRepository = "ghcr.io/canonical/cilium-operator"

	// ciliumOperatorImageTag is the tag to use for the cilium-operator image.
	ciliumOperatorImageTag = "1.15.2-ck1"
)
