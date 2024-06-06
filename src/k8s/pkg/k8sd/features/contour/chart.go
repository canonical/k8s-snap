package contour

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chartContour represents manifests to deploy Contour.
	chartContour = helm.InstallableChart{
		Name:         "ck-ingress",
		Namespace:    "projectcontour",
		ManifestPath: path.Join("charts", "contour-18.1.2.tgz"),
	}
	chartGateway = helm.InstallableChart{
		Name:         "ck-gateway",
		Namespace:    "projectcontour",
		ManifestPath: path.Join("charts", "ck-gateway"),
	}
)
