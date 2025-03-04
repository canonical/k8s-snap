package local_storage

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/features/manifests"
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	rawFileImage := FeatureLocalStorage.GetImage(RawFileImageName)
	csiNodeDriverImage := FeatureLocalStorage.GetImage(CSINodeDriverImageName)
	csiProvisionerImage := FeatureLocalStorage.GetImage(CSIProvisionerImageName)
	csiResizerImage := FeatureLocalStorage.GetImage(CSIResizerImageName)
	csiSnapshotterImage := FeatureLocalStorage.GetImage(CSISnapshotterImageName)

	images.Register(
		// Rawfile LocalPV CSI driver images
		fmt.Sprintf("%s:%s", rawFileImage.GetURI(), rawFileImage.Tag),
		// CSI images
		fmt.Sprintf("%s:%s", csiNodeDriverImage.GetURI(), csiNodeDriverImage.Tag),
		fmt.Sprintf("%s:%s", csiProvisionerImage.GetURI(), csiProvisionerImage.Tag),
		fmt.Sprintf("%s:%s", csiResizerImage.GetURI(), csiResizerImage.Tag),
		fmt.Sprintf("%s:%s", csiSnapshotterImage.GetURI(), csiSnapshotterImage.Tag),
	)

	manifests.Register(&manifest)
}
