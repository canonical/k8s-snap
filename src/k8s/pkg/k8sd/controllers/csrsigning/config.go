package csrsigning

import "github.com/canonical/k8s/pkg/k8sd/types"

type internalConfig struct {
	autoApprove bool
}

func internalConfigFromAnnotations(annotations types.Annotations) internalConfig {
	var cfg internalConfig
	if v, ok := annotations.Get("k8sd/v1alpha1/csrsigning/auto-approve"); ok && v == "true" {
		cfg.autoApprove = true
	}
	return cfg
}
