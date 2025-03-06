package metrics_server

import (
	"embed"

	"github.com/canonical/k8s/pkg/client/helm"
)

//go:embed all:charts
var ChartFS embed.FS

var (
	// chart represents manifests to deploy metrics-server.
	chart = helm.InstallableChart{
		Name:             "metrics-server",
		Version:          "3.12.2",
		InstallName:      "metrics-server",
		InstallNamespace: "kube-system",
	}

	// imageRepo is the image to use for metrics-server.
	imageRepo = "ghcr.io/canonical/metrics-server"

	// imageTag is the image tag to use for metrics-server.
	imageTag = "0.7.2-ck0"
)
