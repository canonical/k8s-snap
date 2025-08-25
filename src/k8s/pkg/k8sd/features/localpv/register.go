package localpv

import (
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		// Rawfile LocalPV CSI driver images
		LocalPVImage().String(),
		// CSI images
		CSINodeDriverImage().String(),
		CSIProvisionerImage().String(),
		CSIResizerImage().String(),
		CSISnapshotterImage().String(),
	)
}
