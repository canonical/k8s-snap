package localpv

import (
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// Chart represents manifests to deploy Rawfile LocalPV CSI.
	Chart = helm.InstallableChart{
		Name:         "ck-storage",
		Namespace:    "kube-system",
		ManifestPath: filepath.Join("charts", "rawfile-csi-0.9.1.tgz"),
	}

	// imageRepo is the repository to use for Rawfile LocalPV CSI.
	imageRepo = "ghcr.io/canonical/rawfile-localpv"
	// ImageTag is the image tag to use for Rawfile LocalPV CSI.
	ImageTag = "0.8.2-ck3"

	// csiNodeDriverImage is the image to use for the CSI node driver.
	csiNodeDriverImage = "ghcr.io/canonical/csi-node-driver-registrar:3347350ff6d84b5544f75b67d25200997c5fd6c67affe48feb374fbda1dffae1-amd64"
	// csiProvisionerImage is the image to use for the CSI provisioner.
	csiProvisionerImage = "ghcr.io/canonical/csi-provisioner:cf6224e54904e45fb18cbcde368789e86ea9cddaada559b843da7b8c63397d99-amd64"
	// csiResizerImage is the image to use for the CSI resizer.
	csiResizerImage = "ghcr.io/canonical/csi-resizer:2b3ce9bde190654dae7639f57a1528c79e09913b8bdf6a5dc9e711811c7c076d-amd64"
	// csiSnapshotterImage is the image to use for the CSI snapshotter.
	csiSnapshotterImage = "ghcr.io/canonical/csi-snapshotter:1e41fc1fabf81b20e29ab4f1cff1452dc95f9c70758f3d673fb1dea443f463a4-amd64"
)
