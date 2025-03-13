package ingress

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	ChartContourName           = "contour"
	ChartDefaultTLSName        = "ck-ingress-tls"
	ChartCommonContourCRDSName = "ck-contour-common"

	ContourIngressEnvoyImageName   = "envoy-ingress"
	ContourIngressContourImageName = "contour-ingress"
)

var manifest = types.FeatureManifest{
	Name:    "ingress",
	Version: "1.0.0",
	Charts: map[string]helm.InstallableChart{
		ChartContourName: {
			Name:             "contour",
			Version:          "17.0.4",
			InstallName:      "ck-ingress",
			InstallNamespace: "projectcontour",
		},
		ChartCommonContourCRDSName: {
			Name:             "ck-contour-common",
			Version:          "1.28.2",
			InstallName:      "ck-contour-common",
			InstallNamespace: "projectcontour",
		},
		ChartDefaultTLSName: {
			Name:             "ck-ingress-tls",
			Version:          "0.1.0",
			InstallName:      "ck-ingress-tls",
			InstallNamespace: "projectcontour-root",
		},
	},

	Images: map[string]types.Image{
		ContourIngressEnvoyImageName: {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "bitnami/envoy",
			Tag:        "1.28.2-debian-12-r0",
		},
		ContourIngressContourImageName: {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "bitnami/contour",
			Tag:        "1.28.2-debian-12-r4",
		},
	},
}

var FeatureIngress types.Feature = manifest
