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
	annotationAutodetectionV4Firstfound    = "k8sd/v1alpha1/calico/autodetection-v4/firstFound"
	annotationAutodetectionV4Kubernetes    = "k8sd/v1alpha1/calico/autodetection-v4/kubernetes"
	annotationAutodetectionV4Interface     = "k8sd/v1alpha1/calico/autodetection-v4/interface"
	annotationAutodetectionV4SkipInterface = "k8sd/v1alpha1/calico/autodetection-v4/skipInterface"
	annotationAutodetectionV4CanReach      = "k8sd/v1alpha1/calico/autodetection-v4/canReach"
	annotationAutodetectionV4Cidrs         = "k8sd/v1alpha1/calico/autodetection-v4/cidrs"
	annotationAutodetectionV6Firstfound    = "k8sd/v1alpha1/calico/autodetection-v6/firstFound"
	annotationAutodetectionV6Kubernetes    = "k8sd/v1alpha1/calico/autodetection-v6/kubernetes"
	annotationAutodetectionV6Interface     = "k8sd/v1alpha1/calico/autodetection-v6/interface"
	annotationAutodetectionV6SkipInterface = "k8sd/v1alpha1/calico/autodetection-v6/skipInterface"
	annotationAutodetectionV6CanReach      = "k8sd/v1alpha1/calico/autodetection-v6/canReach"
	annotationAutodetectionV6Cidrs         = "k8sd/v1alpha1/calico/autodetection-v6/cidrs"
)

type config struct {
	encapsulationV4  string
	encapsulationV6  string
	apiServerEnabled bool
	autodetectionV4  map[string]any
	autodetectionV6  map[string]any
}

func validateEncapsulation(encapsulation string) bool {
	switch encapsulation {
	case "VXLAN",
		"IPIP",
		"IPIPCrossSubnet",
		"VXLANCrossSubnet",
		"None":
		return true
	}
	return false
}

func internalConfig(annotations types.Annotations) (config, error) {
	config := config{
		encapsulationV4:  encapsulation,
		encapsulationV6:  encapsulation,
		apiServerEnabled: apiServerEnabled,
	}

	if v, ok := annotations.Get(annotationAPIServerEnabled); ok {
		config.apiServerEnabled = v == "true"
	}

	if v, ok := annotations.Get(annotationEncapsulationV4); ok {
		if !validateEncapsulation(v) {
			return config, fmt.Errorf("invalid encapsulation-v4 annotation: %s", v)
		}
		config.encapsulationV4 = v
	}

	if v, ok := annotations.Get(annotationEncapsulationV6); ok {
		if !validateEncapsulation(v) {
			return config, fmt.Errorf("invalid encapsulation-v6 annotation: %s", v)
		}
		config.encapsulationV6 = v
	}

	if v, ok := annotations.Get(annotationAutodetectionV4Firstfound); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found")
		}
		config.autodetectionV4 = map[string]any{
			"firstFound": v == "true",
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV4Kubernetes); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found")
		}
		config.autodetectionV4 = map[string]any{
			"kubernetes": v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV4Interface); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found")
		}
		config.autodetectionV4 = map[string]any{
			"interface": v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV4SkipInterface); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found")
		}
		config.autodetectionV4 = map[string]any{
			"skipInterface": v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV4CanReach); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found")
		}
		config.autodetectionV4 = map[string]any{
			"canReach": v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV4Cidrs); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found")
		}
		config.autodetectionV4 = map[string]any{
			"cidrs": strings.Split(v, ","),
		}
	}

	if v, ok := annotations.Get(annotationAutodetectionV6Firstfound); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found")
		}
		config.autodetectionV6 = map[string]any{
			"firstFound": v == "true",
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV6Kubernetes); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found")
		}
		config.autodetectionV6 = map[string]any{
			"kubernetes": v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV6Interface); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found")
		}
		config.autodetectionV6 = map[string]any{
			"interface": v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV6SkipInterface); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found")
		}
		config.autodetectionV6 = map[string]any{
			"skipInterface": v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV6CanReach); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found")
		}
		config.autodetectionV6 = map[string]any{
			"canReach": v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV6Cidrs); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found")
		}
		config.autodetectionV6 = map[string]any{
			"cidrs": strings.Split(v, ","),
		}
	}

	return config, nil
}
