package localpv

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chart represents manifests to deploy Rawfile LocalPV CSI.
	chart = helm.InstallableChart{
		Name:         "ck-storage",
		Namespace:    "kube-system",
		ManifestPath: path.Join("charts", "rawfile-csi-0.9.0.tgz"),
	}

	// imageRepo is the repository to use for Rawfile LocalPV CSI.
	imageRepo = "ghcr.io/canonical/rawfile-localpv"
	// csiImageRepo is the repository to use for the CSI images.
	csiImageRepo = "ghcr.io/canonical/k8s-snap/sig-storage"
	// csiNodeDriverImage is the image to use for the CSI node driver.
	csiNodeDriverImage = "csi-node-driver-registrar:v2.10.1"
	// csiProvisionerImage is the image to use for the CSI provisioner.
	csiProvisionerImage = "csi-provisioner:v5.0.1"
	// csiResizerImage is the image to use for the CSI resizer.
	csiResizerImage = "csi-resizer:v1.11.1"
	// csiSnapshotterImage is the image to use for the CSI snapshotter.
	csiSnapshotterImage = "csi-snapshotter:v8.0.1"

	// imageTag is the image tag to use.
	imageTag = "0.8.0-ck5"
)
