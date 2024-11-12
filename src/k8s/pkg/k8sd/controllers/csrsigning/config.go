package csrsigning

import (
	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations/csrsigning"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type internalConfig struct {
	autoApprove bool
}

func internalConfigFromAnnotations(annotations types.Annotations) internalConfig {
	var cfg internalConfig
	if v, ok := annotations.Get(apiv1_annotations.AnnotationAutoApprove); ok && v == "true" {
		cfg.autoApprove = true
	}
	return cfg
}
