package coredns

import (
	"embed"

	"github.com/canonical/k8s/pkg/client/helm"
)

//go:embed all:charts
var ChartFS embed.FS

var (
	// chartCoreDNS represents manifests to deploy CoreDNS.
	Chart = helm.InstallableChart{
		Name:             "coredns",
		Version:          "1.36.0",
		InstallName:      "ck-dns",
		InstallNamespace: "kube-system",
	}

	// imageRepo is the image to use for CoreDNS.
	imageRepo = "ghcr.io/canonical/coredns"

	// ImageTag is the tag to use for the CoreDNS image.
	ImageTag = "1.11.3-ck0"
)
