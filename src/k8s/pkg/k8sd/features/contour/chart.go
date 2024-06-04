package contour

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chartContour represents manifests to deploy Contour.
	chartContour = helm.InstallableChart{
		Name:         "ck-ingress",
		Namespace:    "project-contour",
		ManifestPath: path.Join("charts", "contour-18.1.1.tgz"),
	}

	// contourAgentImageRepo represents the image to use for contour-agent.
	contourChartRepo = "https://charts.bitnami.com/bitnami"

	// contourAgentImageTag is the tag to use for the contour-agent image.
	contourChartVersionTag = "18.1.1"
)
