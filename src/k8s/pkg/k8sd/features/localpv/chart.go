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
	ImageTag = "df92e4b86a2978f1caa323c075597dc03ab03b9fd8140ea29cbb6489c49cef19-amd64"

	// csiNodeDriverImage is the image to use for the CSI node driver.
	csiNodeDriverImage = "ghcr.io/canonical/csi-node-driver-registrar:c6df6b7b0750dbf2bd62f8a334ad230005918007f7e55d2a0d53036e7da7d42b-amd64"
	// csiProvisionerImage is the image to use for the CSI provisioner.
	csiProvisionerImage = "ghcr.io/canonical/csi-provisioner:32d209965cc4815a421427d1e78d64180e7892844e4de6533a7a7201bbdfbd73-amd64"
	// csiResizerImage is the image to use for the CSI resizer.
	csiResizerImage = "ghcr.io/canonical/csi-resizer:cc127d222e216c6c819b8709c4149d47f179f17b7714553f46bfee1ce9b90613-amd64"
	// csiSnapshotterImage is the image to use for the CSI snapshotter.
	csiSnapshotterImage = "ghcr.io/canonical/csi-snapshotter:c26d33f793cc33fe24f59bc41616d09017a258e609dc53f2141f97370edf035b-amd64"
)
