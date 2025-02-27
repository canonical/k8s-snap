package dns

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	CoreDNSChartName = "coredns"
	CoreDNSImageName = "coredns"
)

var manifest = types.FeatureManifest{
	Name:    "dns",
	Version: "1.0.0",
	Charts: map[string]helm.InstallableChart{
		CoreDNSChartName: {
			Name:             "coredns",
			Version:          "1.36.2",
			InstallName:      "ck-dns",
			InstallNamespace: "kube-system",
		},
	},

	Images: map[string]types.Image{
		CoreDNSImageName: {
			Registry:   "ghcr.io/canonical",
			Repository: "coredns",
			Tag:        "1.11.4-ck1",
		},
	},
}

var FeatureDNS types.Feature = manifest
