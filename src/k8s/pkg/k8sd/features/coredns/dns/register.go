package dns

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	coreDNSImage := Manifest.GetImage(CoreDNSImageName)

	images.Register(
		fmt.Sprintf("%s:%s", coreDNSImage.GetURI(), coreDNSImage.Tag),
	)

	features.Register(&Manifest)
}
