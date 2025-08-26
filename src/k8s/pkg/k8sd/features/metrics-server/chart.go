package metrics_server

import (
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chart represents manifests to deploy metrics-server.
	chart = helm.InstallableChart{
		Name:         "metrics-server",
		Namespace:    "kube-system",
		ManifestPath: filepath.Join("charts", "metrics-server-3.12.2.tgz"),
	}

	// imageRepo is the image to use for metrics-server.
	imageRepo = "ghcr.io/canonical/metrics-server"

	// imageTag is the image tag to use for metrics-server.
	imageTag = "70fb4d79a8adf04f1de176784e0b40afce13a997bbf7c5dcc3b3717e953bfe1a-amd64"
)
