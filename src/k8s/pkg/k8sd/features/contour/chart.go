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
	chartDefaultTLS = helm.InstallableChart{
		Name:         "ck-ingress-tls",
		Namespace:    "projectcontour-root",
		ManifestPath: path.Join("charts", "ck-ingress-tls"),
	}
	chartCommonContourCRDS = helm.InstallableChart{
		Name:         "ck-contour-common",
		Namespace:    "projectcontour",
		ManifestPath: path.Join("charts", "ck-contour-common"),
	}
)
