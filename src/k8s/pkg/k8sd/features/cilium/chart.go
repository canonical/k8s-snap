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
	CiliumAgentImageTag = "3bb952b7bb6e77b04cea05a9c220dcfe6a70da56cdb1879d035d3382782c149a-amd64"

	// ciliumOperatorImageRepo is the image to use for cilium-operator.
	ciliumOperatorImageRepo = "ghcr.io/canonical/cilium-operator"

	// ciliumOperatorImageTag is the tag to use for the cilium-operator image.
	ciliumOperatorImageTag = "593f0252cdcaf6d0d13528fadaaa2a106fec41b1dc979f0eaeb355a4e85c362c-amd64"

	ciliumDefaultVXLANPort = 8472

	ciliumVXLANDeviceName = "cilium_vxlan"
)
