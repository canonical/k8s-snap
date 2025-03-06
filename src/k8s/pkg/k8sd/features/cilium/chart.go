package cilium

import (
	"embed"

	"github.com/canonical/k8s/pkg/client/helm"
)

//go:embed all:charts
var ChartFS embed.FS

var (
	// ChartCilium represents manifests to deploy Cilium.
	ChartCilium = helm.InstallableChart{
		Name:             "cilium",
		Version:          "1.17.1",
		InstallName:      "ck-network",
		InstallNamespace: "kube-system",
	}

	// ChartCiliumLoadBalancer represents manifests to deploy Cilium LoadBalancer resources.
	ChartCiliumLoadBalancer = helm.InstallableChart{
		Name:             "ck-loadbalancer",
		Version:          "0.1.1",
		InstallName:      "ck-loadbalancer",
		InstallNamespace: "kube-system",
	}

	// chartGateway represents manifests to deploy Gateway API CRDs.
	chartGateway = helm.InstallableChart{
		Name:             "gateway-api",
		Version:          "1.2.0",
		InstallName:      "ck-gateway",
		InstallNamespace: "kube-system",
	}

	// chartGatewayClass represents a manifest to deploy a GatewayClass called ck-gateway.
	chartGatewayClass = helm.InstallableChart{
		Name:             "ck-gateway-cilium",
		Version:          "0.1.0",
		InstallName:      "ck-gateway-class",
		InstallNamespace: "default",
	}

	// ciliumAgentImageRepo represents the image to use for cilium-agent.
	ciliumAgentImageRepo = "ghcr.io/canonical/cilium"

	// CiliumAgentImageTag is the tag to use for the cilium-agent image.
	CiliumAgentImageTag = "1.17.1-ck0"

	// ciliumOperatorImageRepo is the image to use for cilium-operator.
	ciliumOperatorImageRepo = "ghcr.io/canonical/cilium-operator"

	// ciliumOperatorImageTag is the tag to use for the cilium-operator image.
	ciliumOperatorImageTag = "1.17.1-ck0"
)
