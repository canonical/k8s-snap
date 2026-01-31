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
		ManifestPath: filepath.Join("charts", "cilium-1.17.1.tgz"),
	}

	// ChartCiliumLoadBalancer represents manifests to deploy Cilium LoadBalancer resources.
	ChartCiliumLoadBalancer = helm.InstallableChart{
		Name:         "ck-loadbalancer",
		Namespace:    "kube-system",
		ManifestPath: filepath.Join("charts", "ck-loadbalancer"),
	}

	// chartGateway represents manifests to deploy Gateway API CRDs.
	chartGateway = helm.InstallableChart{
		Name:         "ck-gateway",
		Namespace:    "kube-system",
		ManifestPath: filepath.Join("charts", "gateway-api-1.2.0.tgz"),
	}

	// chartGatewayClass represents a manifest to deploy a GatewayClass called ck-gateway.
	chartGatewayClass = helm.InstallableChart{
		Name:         "ck-gateway-class",
		Namespace:    "default",
		ManifestPath: filepath.Join("charts", "ck-gateway-cilium"),
	}

	// ciliumAgentImageRepo represents the image to use for cilium-agent.
	ciliumAgentImageRepo = "ghcr.io/canonical/cilium"

	// CiliumAgentImageTag is the tag to use for the cilium-agent image.
	CiliumAgentImageTag = "492af12a9b254e59c4d9156fffe54371426af02e463bda2c704be10d38eb9ba3-amd64"

	// ciliumOperatorImageRepo is the image to use for cilium-operator.
	ciliumOperatorImageRepo = "ghcr.io/canonical/cilium-operator"

	// ciliumOperatorImageTag is the tag to use for the cilium-operator image.
	ciliumOperatorImageTag = "49d02f8f383b9ff6cf2b83f850448e47f71349c1b8bfef5cb6fc16d7b5d24c29-amd64"

	ciliumDefaultVXLANPort = 8472

	ciliumVXLANDeviceName = "cilium_vxlan"
)
