package local_storage

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	RawFileChartName = "rawfile-csi"

	RawFileImageName        = "rawfile-localpv"
	CSINodeDriverImageName  = "csi-node-driver-registrar"
	CSIProvisionerImageName = "csi-provisioner"
	CSIResizerImageName     = "csi-resizer"
	CSISnapshotterImageName = "csi-snapshotter"
)

var manifest = types.FeatureManifest{
	Name:    "local-storage",
	Version: "1.0.0",
	Charts: map[string]helm.InstallableChart{
		RawFileChartName: {
			Name:             "rawfile-csi",
			Version:          "0.9.0",
			InstallName:      "ck-storage",
			InstallNamespace: "kube-system",
		},
	},

	Images: map[string]types.Image{
		RawFileImageName: {
			Registry:   "ghcr.io/canonical",
			Repository: "rawfile-localpv",
			Tag:        "0.8.1",
		},
		CSINodeDriverImageName: {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "sig-storage/csi-node-driver-registrar",
			Tag:        "v2.10.1",
		},
		CSIProvisionerImageName: {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "sig-storage/csi-provisioner",
			Tag:        "v5.0.2",
		},
		CSIResizerImageName: {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "sig-storage/csi-resizer",
			Tag:        "v1.11.2",
		},
		CSISnapshotterImageName: {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "sig-storage/csi-snapshotter",
			Tag:        "v8.0.2",
		},
	},
}

var FeatureLocalStorage types.Feature = manifest
