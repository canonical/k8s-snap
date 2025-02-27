package loadbalancer

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	MetalLBChartName      = "metallb"
	LoadBalancerChartName = "ck-loadbalancer"

	MetalLBControllerImageName = "metallb-controller"
	MetalLBSpeakerImageName    = "metallb-speaker"
	FRRImageName               = "frr"
)

var manifest = types.FeatureManifest{
	Name:    "loadbalancer",
	Version: "1.0.0",
	Charts: map[string]helm.InstallableChart{
		MetalLBChartName: {
			Name:             "metallb",
			Version:          "0.14.9",
			InstallName:      "metallb",
			InstallNamespace: "metallb-system",
		},
		LoadBalancerChartName: {
			Name:             "ck-loadbalancer",
			Version:          "0.1.1",
			InstallName:      "metallb-loadbalancer",
			InstallNamespace: "metallb-system",
		},
	},

	Images: map[string]types.Image{
		MetalLBControllerImageName: {
			Registry:   "ghcr.io/canonical",
			Repository: "metallb-controller",
			Tag:        "v0.14.9-ck0",
		},
		MetalLBSpeakerImageName: {
			Registry:   "ghcr.io/canonical",
			Repository: "metallb-speaker",
			Tag:        "v0.14.9-ck0",
		},
		FRRImageName: {
			Registry:   "ghcr.io/canonical",
			Repository: "frr",
			Tag:        "9.1.3",
		},
	},
}

var FeatureLoadBalancer types.Feature = manifest
