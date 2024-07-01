package calico

import (
	"fmt"
	"strings"

	"github.com/canonical/k8s/pkg/k8sd/types"
)

const (
	annotationAPIServerEnabled             = "k8sd/v1alpha1/calico/apiserver-enabled"
	annotationEncapsulationV4              = "k8sd/v1alpha1/calico/encapsulation-v4"
	annotationEncapsulationV6              = "k8sd/v1alpha1/calico/encapsulation-v6"
	annotationAutodetectionV4FirstFound    = "k8sd/v1alpha1/calico/autodetection-v4/firstFound"
	annotationAutodetectionV4Kubernetes    = "k8sd/v1alpha1/calico/autodetection-v4/kubernetes"
	annotationAutodetectionV4Interface     = "k8sd/v1alpha1/calico/autodetection-v4/interface"
	annotationAutodetectionV4SkipInterface = "k8sd/v1alpha1/calico/autodetection-v4/skipInterface"
	annotationAutodetectionV4CanReach      = "k8sd/v1alpha1/calico/autodetection-v4/canReach"
	annotationAutodetectionV4CIDRs         = "k8sd/v1alpha1/calico/autodetection-v4/cidrs"
	annotationAutodetectionV6FirstFound    = "k8sd/v1alpha1/calico/autodetection-v6/firstFound"
	annotationAutodetectionV6Kubernetes    = "k8sd/v1alpha1/calico/autodetection-v6/kubernetes"
	annotationAutodetectionV6Interface     = "k8sd/v1alpha1/calico/autodetection-v6/interface"
	annotationAutodetectionV6SkipInterface = "k8sd/v1alpha1/calico/autodetection-v6/skipInterface"
	annotationAutodetectionV6CanReach      = "k8sd/v1alpha1/calico/autodetection-v6/canReach"
	annotationAutodetectionV6CIDRs         = "k8sd/v1alpha1/calico/autodetection-v6/cidrs"
)

type config struct {
	encapsulationV4  string
	encapsulationV6  string
	apiServerEnabled bool
	autodetectionV4  map[string]any
	autodetectionV6  map[string]any
}

func checkEncapsulation(v string) error {
	switch v {
	case "VXLAN", "IPIP", "IPIPCrossSubnet", "VXLANCrossSubnet", "None":
		return nil
	}
	return fmt.Errorf("unsupported encapsulation type: %s", v)
}

func internalConfig(annotations types.Annotations) (config, error) {
	config := config{
		encapsulationV4:  defaultEncapsulation,
		encapsulationV6:  defaultEncapsulation,
		apiServerEnabled: defaultAPIServerEnabled,
	}

	if v, ok := annotations.Get(annotationAPIServerEnabled); ok {
		config.apiServerEnabled = v == "true"
	}

	if v, ok := annotations.Get(annotationEncapsulationV4); ok {
		if err := checkEncapsulation(v); err != nil {
			return config, fmt.Errorf("invalid encapsulation-v4 annotation: %w", err)
		}
		config.encapsulationV4 = v
	}

	if v, ok := annotations.Get(annotationEncapsulationV6); ok {
		if err := checkEncapsulation(v); err != nil {
			return config, fmt.Errorf("invalid encapsulation-v6 annotation: %w", err)
		}
		config.encapsulationV6 = v
	}

	var autodetectionV4Key string
	var autodetectionV4Value any

	if v, ok := annotations.Get(annotationAutodetectionV4FirstFound); ok {
		if autodetectionV4Key != "" {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4FirstFound)
		}
		autodetectionV4Key = "firstFound"
		autodetectionV4Value = v == "true"
	}
	if v, ok := annotations.Get(annotationAutodetectionV4Kubernetes); ok {
		if autodetectionV4Key != "" {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4Kubernetes)
		}
		autodetectionV4Key = "kubernetes"
		autodetectionV4Value = v
	}
	if v, ok := annotations.Get(annotationAutodetectionV4Interface); ok {
		if autodetectionV4Key != "" {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4Interface)
		}
		autodetectionV4Key = "interface"
		autodetectionV4Value = v
	}
	if v, ok := annotations.Get(annotationAutodetectionV4SkipInterface); ok {
		if autodetectionV4Key != "" {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4SkipInterface)
		}
		autodetectionV4Key = "skipInterface"
		autodetectionV4Value = v
	}
	if v, ok := annotations.Get(annotationAutodetectionV4CanReach); ok {
		if autodetectionV4Key != "" {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4CanReach)
		}
		autodetectionV4Key = "canReach"
		autodetectionV4Value = v
	}
	if v, ok := annotations.Get(annotationAutodetectionV4CIDRs); ok {
		if autodetectionV4Key != "" {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4CIDRs)
		}
		autodetectionV4Key = "cidrs"
		autodetectionV4Value = strings.Split(v, ",")
	}

	// If any annotation is set, pass the map to the config otherwise it's left nil
	if autodetectionV4Key != "" {
		config.autodetectionV4 = map[string]any{
			autodetectionV4Key: autodetectionV4Value,
		}
	}

	var autodetectionV6Key string
	var autodetectionV6Value any

	if v, ok := annotations.Get(annotationAutodetectionV6FirstFound); ok {
		if autodetectionV6Key != "" {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6FirstFound)
		}
		autodetectionV6Key = "firstFound"
		autodetectionV6Value = v == "true"
	}
	if v, ok := annotations.Get(annotationAutodetectionV6Kubernetes); ok {
		if autodetectionV6Key != "" {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6Kubernetes)
		}
		autodetectionV6Key = "kubernetes"
		autodetectionV6Value = v
	}
	if v, ok := annotations.Get(annotationAutodetectionV6Interface); ok {
		if autodetectionV6Key != "" {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6Interface)
		}
		autodetectionV6Key = "interface"
		autodetectionV6Value = v
	}
	if v, ok := annotations.Get(annotationAutodetectionV6SkipInterface); ok {
		if autodetectionV6Key != "" {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6SkipInterface)
		}
		autodetectionV6Key = "skipInterface"
		autodetectionV6Value = v
	}
	if v, ok := annotations.Get(annotationAutodetectionV6CanReach); ok {
		if autodetectionV6Key != "" {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6CanReach)
		}
		autodetectionV6Key = "canReach"
		autodetectionV6Value = v
	}
	if v, ok := annotations.Get(annotationAutodetectionV6CIDRs); ok {
		if autodetectionV6Key != "" {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6CIDRs)
		}
		autodetectionV6Key = "cidrs"
		autodetectionV6Value = strings.Split(v, ",")
	}

	// If any annotation is set, pass the map to the config otherwise it's left nil
	if autodetectionV6Key != "" {
		config.autodetectionV6 = map[string]any{
			autodetectionV6Key: autodetectionV6Value,
		}
	}

	return config, nil
}
