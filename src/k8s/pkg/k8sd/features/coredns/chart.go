package coredns

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chartCoreDNS represents manifests to deploy CoreDNS.
	chart = helm.InstallableChart{
		Name:         "ck-dns",
		Namespace:    "kube-system",
		ManifestPath: path.Join("charts", "coredns-1.29.0.tgz"),
	}

	// imageRepo is the image to use for CoreDNS.
	imageRepo = "ghcr.io/canonical/coredns"

	// imageTag is the tag to use for the CoreDNS image.
	imageTag = "1.11.1-ck4"
)
