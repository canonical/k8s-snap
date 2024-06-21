package localpv

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		// Rawfile LocalPV CSI driver images
		fmt.Sprintf("%s:%s", imageRepo, imageTag),
		// CSI images
		fmt.Sprintf("%s/%s", csiImageRepo, csiProvisionerImage),
		fmt.Sprintf("%s/%s", csiImageRepo, csiResizerImage),
		fmt.Sprintf("%s/%s", csiImageRepo, csiSnapshotterImage),
		fmt.Sprintf("%s/%s", csiImageRepo, csiNodeDriverImage),
	)
}
