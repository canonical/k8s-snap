package metallb

import (
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/helm"
	k8sdConfig "github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	// ChartMetalLB represents manifests to deploy MetalLB speaker and controller.
	ChartMetalLB = helm.InstallableChart{
		Name:         "metallb",
		Namespace:    "metallb-system",
		ManifestPath: filepath.Join("charts", "metallb-0.14.9.tgz"),
	}

	// ChartMetalLBLoadBalancer represents manifests to deploy MetalLB L2 or BGP resources.
	ChartMetalLBLoadBalancer = helm.InstallableChart{
		Name:         "metallb-loadbalancer",
		Namespace:    "metallb-system",
		ManifestPath: filepath.Join("charts", "ck-loadbalancer"),
	}
)

func MetalLBControllerImage() types.Image {
	imageRepo := "ghcr.io/canonical/metallb-controller"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: imageRepo,
			Tag:        "v0.14.9-fips-ck0",
		}
	}

	return types.Image{
		Repository: imageRepo,
		Tag:        "v0.14.9-ck0",
	}
}

func MetalLBSpeakerImage() types.Image {
	imageRepo := "ghcr.io/canonical/metallb-speaker"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: imageRepo,
			Tag:        "v0.14.9-fips-ck0",
		}
	}

	return types.Image{
		Repository: imageRepo,
		Tag:        "v0.14.9-ck0",
	}
}

func FRRImage() types.Image {
	imageRepo := "ghcr.io/canonical/frr"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: imageRepo,
			Tag:        "9.1.3-fips-ck0",
		}
	}

	return types.Image{
		Repository: imageRepo,
		Tag:        "9.1.3-ck1",
	}
}
