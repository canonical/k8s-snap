package localpv

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		fmt.Sprintf("%s:%s", imageRepo, imageTag),
	)

	// TODO: configurable CSI images, include in this list
	//
	// Hardcoded list based on "k8s kubectl get node -o template='{{ range .items }}{{ .metadata.name }}{{":"}}{{ range .status.images }}{{ "\n- " }}{{ index .names 1 }}{{ end }}{{"\n"}}{{ end }}' | grep csi"
	//
	// - k8s.gcr.io/sig-storage/csi-provisioner:v3.4.1
	// - k8s.gcr.io/sig-storage/csi-resizer:v1.7.0
	// - k8s.gcr.io/sig-storage/csi-snapshotter:v6.2.1
	// - k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.10.0

	images.Register(
		"k8s.gcr.io/sig-storage/csi-provisioner:v3.4.1",
		"k8s.gcr.io/sig-storage/csi-resizer:v1.7.0",
		"k8s.gcr.io/sig-storage/csi-snapshotter:v6.2.1",
		"k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.10.0",
	)
}
