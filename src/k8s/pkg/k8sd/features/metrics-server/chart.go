package metrics_server

import (
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/helm"
	k8sdConfig "github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	// chart represents manifests to deploy metrics-server.
	chart = helm.InstallableChart{
		Name:         "metrics-server",
		Namespace:    "kube-system",
		ManifestPath: filepath.Join("charts", "metrics-server-3.12.2.tgz"),
	}
)

func MetricsServerImage() types.Image {
	imageRepo := "ghcr.io/canonical/metrics-server"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: imageRepo,
			Tag:        "0.7.2-fips-ck0",
		}
	}

	return types.Image{
		Repository: imageRepo,
		Tag:        "0.7.2-ck0",
	}
}
