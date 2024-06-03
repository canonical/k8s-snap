package metrics_server

import "github.com/canonical/k8s/pkg/k8sd/types"

const (
	annotationImageRepo = "k8sd/v1alpha1/metrics-server/image-repo"
	annotationImageTag  = "k8sd/v1alpha1/metrics-server/image-tag"
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

	if v, ok := annotations.Get(annotationImageRepo); ok {
		config.imageRepo = v
	}
	if v, ok := annotations.Get(annotationImageTag); ok {
		config.imageTag = v
	}

	return config
}
