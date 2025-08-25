package cilium

import (
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/helm"
	k8sdConfig "github.com/canonical/k8s/pkg/config"

	"github.com/canonical/k8s/pkg/k8sd/types"
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

	ciliumDefaultVXLANPort = 8472

	ciliumVXLANDeviceName = "cilium_vxlan"
)

func CiliumAgentImage() types.Image {
	agentRepo := "ghcr.io/canonical/cilium"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: agentRepo,
			Tag:        "1.17.1-fips-ck0",
		}
	}

	return types.Image{
		Repository: agentRepo,
		Tag:        "1.17.1-ck2",
	}
}

func CiliumOperatorImage() types.Image {
	operatorRepo := "ghcr.io/canonical/cilium-operator"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: operatorRepo,
			Tag:        "1.17.1-fips-ck0",
		}
	}

	return types.Image{
		Repository: operatorRepo,
		Tag:        "1.17.1-ck2",
	}
}
