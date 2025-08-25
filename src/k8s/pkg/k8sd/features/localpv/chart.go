package localpv

import (
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/helm"
	k8sdConfig "github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	// Chart represents manifests to deploy Rawfile LocalPV CSI.
	Chart = helm.InstallableChart{
		Name:         "ck-storage",
		Namespace:    "kube-system",
		ManifestPath: filepath.Join("charts", "rawfile-csi-0.9.1.tgz"),
	}
)

func LocalPVImage() types.Image {
	imageRepo := "ghcr.io/canonical/rawfile-localpv"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: imageRepo,
			Tag:        "0.8.2-fips-ck0",
		}
	}

	return types.Image{
		Repository: imageRepo,
		Tag:        "0.8.2-ck1",
	}
}

func CSINodeDriverImage() types.Image {
	imageRepo := "ghcr.io/canonical/k8s-snap/sig-storage/csi-node-driver-registrar"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: imageRepo,
			Tag:        "v2.10.1-fips-ck0",
		}
	}

	return types.Image{
		Repository: imageRepo,
		Tag:        "v2.10.1",
	}
}

func CSIProvisionerImage() types.Image {
	imageRepo := "ghcr.io/canonical/k8s-snap/sig-storage/csi-provisioner"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: imageRepo,
			Tag:        "v5.0.2-fips-ck0",
		}
	}

	return types.Image{
		Repository: imageRepo,
		Tag:        "v5.0.2",
	}
}

func CSIResizerImage() types.Image {
	imageRepo := "ghcr.io/canonical/k8s-snap/sig-storage/csi-resizer"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: imageRepo,
			Tag:        "v1.11.2-fips-ck0",
		}
	}

	return types.Image{
		Repository: imageRepo,
		Tag:        "v1.11.2",
	}
}

func CSISnapshotterImage() types.Image {
	imageRepo := "ghcr.io/canonical/k8s-snap/sig-storage/csi-snapshotter"

	if k8sdConfig.GetFlavor() == k8sdConfig.FlavorFIPS {
		return types.Image{
			Repository: imageRepo,
			Tag:        "v8.0.2-fips-ck0",
		}
	}

	return types.Image{
		Repository: imageRepo,
		Tag:        "v8.0.2",
	}
}
