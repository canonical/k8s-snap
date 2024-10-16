package metrics_server

import (
	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations/metrics-server"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type config struct {
	imageRepo string
	imageTag  string
}

func internalConfig(annotations types.Annotations) config {
	config := config{
		imageRepo: imageRepo,
		imageTag:  imageTag,
	}

	if v, ok := annotations.Get(apiv1_annotations.AnnotationImageRepo); ok {
		config.imageRepo = v
	}
	if v, ok := annotations.Get(apiv1_annotations.AnnotationImageTag); ok {
		config.imageTag = v
	}

	return config
}
