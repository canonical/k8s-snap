package local_storage

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	rawFileImage := Manifest.GetImage(RawFileImageName)
	csiNodeDriverImage := Manifest.GetImage(CSINodeDriverImageName)
	csiProvisionerImage := Manifest.GetImage(CSIProvisionerImageName)
	csiResizerImage := Manifest.GetImage(CSIResizerImageName)
	csiSnapshotterImage := Manifest.GetImage(CSISnapshotterImageName)

	images.Register(
		// Rawfile LocalPV CSI driver images
		fmt.Sprintf("%s:%s", rawFileImage.GetURI(), rawFileImage.Tag),
		// CSI images
		fmt.Sprintf("%s:%s", csiNodeDriverImage.GetURI(), csiNodeDriverImage.Tag),
		fmt.Sprintf("%s:%s", csiProvisionerImage.GetURI(), csiProvisionerImage.Tag),
		fmt.Sprintf("%s:%s", csiResizerImage.GetURI(), csiResizerImage.Tag),
		fmt.Sprintf("%s:%s", csiSnapshotterImage.GetURI(), csiSnapshotterImage.Tag),
	)

	features.Register(&Manifest)
}
