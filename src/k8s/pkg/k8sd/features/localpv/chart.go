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
		ManifestPath: filepath.Join("charts", "rawfile-csi-0.9.0.tgz"),
	}

	// imageRepo is the repository to use for Rawfile LocalPV CSI.
	imageRepo = "ghcr.io/canonical/rawfile-localpv"
	// ImageTag is the image tag to use for Rawfile LocalPV CSI.
	ImageTag = "0.8.1"

	// csiNodeDriverImage is the image to use for the CSI node driver.
	csiNodeDriverImage = "ghcr.io/canonical/k8s-snap/sig-storage/csi-node-driver-registrar:v2.10.1"
	// csiProvisionerImage is the image to use for the CSI provisioner.
	csiProvisionerImage = "ghcr.io/canonical/k8s-snap/sig-storage/csi-provisioner:v5.0.1"
	// csiResizerImage is the image to use for the CSI resizer.
	csiResizerImage = "ghcr.io/canonical/k8s-snap/sig-storage/csi-resizer:v1.11.1"
	// csiSnapshotterImage is the image to use for the CSI snapshotter.
	csiSnapshotterImage = "ghcr.io/canonical/k8s-snap/sig-storage/csi-snapshotter:v8.0.1"
)
