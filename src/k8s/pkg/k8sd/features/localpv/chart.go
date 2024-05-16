package localpv

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chart represents manifests to deploy Rawfile LocalPV CSI.
	chart = helm.InstallableChart{
		Name:         "ck-storage",
		Namespace:    "kube-system",
		ManifestPath: path.Join("charts", "rawfile-csi-0.8.0.tgz"),
	}

	// imageRepo is the image to use for Rawfile LocalPV CSI.
	imageRepo = "ghcr.io/canonical/rawfile-localpv"

	// imageTag is the image tag to use.
	imageTag = "0.8.0-ck5"
)
