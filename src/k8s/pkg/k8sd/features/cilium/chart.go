package cilium

import (
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// ChartCilium represents manifests to deploy Cilium.
	ChartCilium = helm.InstallableChart{
		Name:         "ck-network",
		Namespace:    "kube-system",
		ManifestPath: filepath.Join("charts", "cilium-1.15.2.tgz"),
	}

	// chartCiliumLoadBalancer represents manifests to deploy Cilium LoadBalancer resources.
	chartCiliumLoadBalancer = helm.InstallableChart{
		Name:         "ck-loadbalancer",
		Namespace:    "kube-system",
		ManifestPath: filepath.Join("charts", "ck-loadbalancer"),
	}

	// chartGateway represents manifests to deploy Gateway API CRDs.
	chartGateway = helm.InstallableChart{
		Name:         "ck-gateway",
		Namespace:    "kube-system",
		ManifestPath: filepath.Join("charts", "gateway-api-1.0.0.tgz"),
	}

	//chartGatewayClass represents a manifest to deploy a GatewayClass called ck-gateway.
	chartGatewayClass = helm.InstallableChart{
		Name:         "ck-gateway-class",
		Namespace:    "default",
		ManifestPath: filepath.Join("charts", "ck-gateway-cilium"),
	}

	// ciliumAgentImageRepo represents the image to use for cilium-agent.
	ciliumAgentImageRepo = "ghcr.io/canonical/cilium"

	// CiliumAgentImageTag is the tag to use for the cilium-agent image.
	CiliumAgentImageTag = "1.15.2-ck2"

	// ciliumOperatorImageRepo is the image to use for cilium-operator.
	ciliumOperatorImageRepo = "ghcr.io/canonical/cilium-operator"

	// ciliumOperatorImageTag is the tag to use for the cilium-operator image.
	ciliumOperatorImageTag = "1.15.2-ck2"
)
