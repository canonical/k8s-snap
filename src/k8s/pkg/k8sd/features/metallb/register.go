package metallb

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		fmt.Sprintf("%s:%s", controllerImageRepo, controllerImageTag),
		fmt.Sprintf("%s:%s", speakerImageRepo, speakerImageTag),
		fmt.Sprintf("%s:%s", frrImageRepo, frrImageTag),
	)
}
