package dns

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/features/manifests"
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	coreDNSImage := FeatureDNS.GetImage(CoreDNSImageName)

	images.Register(
		fmt.Sprintf("%s:%s", coreDNSImage.GetURI(), coreDNSImage.Tag),
	)

	manifests.Register(&manifest)
}
