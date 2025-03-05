package gateway

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	ChartGatewayName           = "ck-gateway-contour"
	ChartCommonContourCRDSName = "ck-contour-common"

	ContourGatewayProvisionerEnvoyImageName   = "envoy-gateway"
	ContourGatewayProvisionerContourImageName = "contour-gateway"
)

var Manifest = types.FeatureManifest{
	Name:    "gateway",
	Version: "1.0.0",
	Charts: map[string]helm.InstallableChart{
		ChartGatewayName: {
			Name:             "ck-gateway-contour",
			Version:          "1.28.2",
			InstallName:      "ck-gateway",
			InstallNamespace: "projectcontour",
		},
		ChartCommonContourCRDSName: {
			Name:             "ck-contour-common",
			Version:          "1.28.2",
			InstallName:      "ck-contour-common",
			InstallNamespace: "projectcontour",
		},
	},

	Images: map[string]types.Image{
		ContourGatewayProvisionerContourImageName: {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "projectcontour/contour",
			Tag:        "v1.28.2",
		},
		ContourGatewayProvisionerEnvoyImageName: {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "envoyproxy/envoy",
			Tag:        "v1.29.2",
		},
	},
}
