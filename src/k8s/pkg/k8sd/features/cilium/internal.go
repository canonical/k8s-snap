package cilium

import (
	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations/cilium"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type config struct {
	devices             string
	directRoutingDevice string
}

func internalConfig(annotations types.Annotations) (config, error) {
	c := config{}

	if v, ok := annotations.Get(apiv1_annotations.AnnotationDevices); ok {
		c.devices = v
	}

	if v, ok := annotations.Get(apiv1_annotations.AnnotationDirectRoutingDevice); ok {
		c.directRoutingDevice = v
	}

	return c, nil
}
