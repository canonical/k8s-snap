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
		ManifestPath: filepath.Join("charts", "cilium-1.16.3.tgz"),
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
		ManifestPath: filepath.Join("charts", "gateway-api-1.1.0.tgz"),
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
	CiliumAgentImageTag = "cd9b04e84bb68cae7c672b66c839b011321b6b44cd91d7521d6f5a63e232ae71-amd64"

	// ciliumOperatorImageRepo is the image to use for cilium-operator.
	ciliumOperatorImageRepo = "ghcr.io/canonical/cilium-operator"

	// ciliumOperatorImageTag is the tag to use for the cilium-operator image.
	ciliumOperatorImageTag = "8d1f1ef6ee8e0036760d131a02f5b598e54779f568bb2b272f3f639c96cfa121-amd64"
)
