package calico

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		fmt.Sprintf("%s/%s:%s", imageRepo, tigeraOperatorImage, tigeraOperatorVersion),
	)

	// TODO: configurable Calico images, include in this list
	//
	// Hardcoded list based on "k8s kubectl get node -o template='{{ range .items }}{{ .metadata.name }}{{":"}}{{ range .status.images }}{{ "\n- " }}{{ index .names 1 }}{{ end }}{{"\n"}}{{ end }}' | grep calico":
	//
	// - docker.io/calico/node:v3.28.0
	// - docker.io/calico/cni:v3.28.0
	// - docker.io/calico/apiserver:v3.28.0
	// - docker.io/calico/kube-controllers:v3.28.0
	// - docker.io/calico/typha:v3.28.0
	// - docker.io/calico/node-driver-registrar:v3.28.0
	// - docker.io/calico/csi:v3.28.0
	// - docker.io/calico/pod2daemon-flexvol:v3.28.0

	images.Register(
		"docker.io/calico/node:v3.28.0",
		"docker.io/calico/cni:v3.28.0",
		"docker.io/calico/apiserver:v3.28.0",
		"docker.io/calico/kube-controllers:v3.28.0",
		"docker.io/calico/typha:v3.28.0",
		"docker.io/calico/node-driver-registrar:v3.28.0",
		"docker.io/calico/csi:v3.28.0",
		"docker.io/calico/pod2daemon-flexvol:v3.28.0",
	)
}
