package coredns

import (
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/helm"
	k8sdConfig "github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	// chartCoreDNS represents manifests to deploy CoreDNS.
	Chart = helm.InstallableChart{
		Name:         "ck-dns",
		Namespace:    "kube-system",
		ManifestPath: filepath.Join("charts", "coredns-1.39.2.tgz"),
	}
)

// CoreDNSImage returns the image to use for CoreDNS.
func CoreDNSImage() types.Image {
	agentRepo := "ghcr.io/canonical/coredns"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: agentRepo,
			Tag:        "1.12.0-fips-ck0",
		}
	}

	return types.Image{
		Repository: agentRepo,
		Tag:        "1.12.0-ck1",
	}
}
