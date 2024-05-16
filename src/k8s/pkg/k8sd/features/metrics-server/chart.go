package metrics_server

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chart is manifests for the built-in metrics-server feature, powered by the upstream metrics-server.
	chart = helm.InstallableChart{
		Name:         "metrics-server",
		Namespace:    "kube-system",
		ManifestPath: path.Join("charts", "metrics-server-3.12.0.tgz"),
	}

	// imageRepo is the image to use for metrics-server.
	imageRepo = "ghcr.io/canonical/metrics-server"

	// imageTag is the image tag to use for metrics-server.
	imageTag = "0.8.0-ck5"
)
