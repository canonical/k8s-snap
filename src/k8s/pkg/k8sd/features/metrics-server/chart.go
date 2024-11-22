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
	imageTag = "0.7.2-ck0"
)
