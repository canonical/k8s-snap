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

type autodetection struct {
	FirstFound         bool     `json:"firstFound,omitempty"`
	Kubernetes         string   `json:"kubernetes,omitempty"`
	InterfaceRegex     string   `json:"interface,omitempty"`
	SkipInterfaceRegex string   `json:"skipInterface,omitempty"`
	CanReach           string   `json:"canReach,omitempty"`
	CIDRs              []string `json:"cidrs,omitempty"`
}

type config struct {
	encapsulationV4  string
	encapsulationV6  string
	apiServerEnabled bool
	autodetectionV4  *autodetection
	autodetectionV6  *autodetection
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
		encapsulationV4:  encapsulation,
		encapsulationV6:  encapsulation,
		apiServerEnabled: apiServerEnabled,
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

	if v, ok := annotations.Get(annotationAutodetectionV4Firstfound); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4Firstfound)
		}
		config.autodetectionV4 = &autodetection{
			FirstFound: v == "true",
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV4Kubernetes); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4Kubernetes)
		}
		config.autodetectionV4 = &autodetection{
			Kubernetes: v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV4Interface); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4Interface)
		}
		config.autodetectionV4 = &autodetection{
			InterfaceRegex: v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV4SkipInterface); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4SkipInterface)
		}
		config.autodetectionV4 = &autodetection{
			SkipInterfaceRegex: v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV4CanReach); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4CanReach)
		}
		config.autodetectionV4 = &autodetection{
			CanReach: v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV4Cidrs); ok {
		if config.autodetectionV4 != nil {
			return config, fmt.Errorf("multiple autodetection-v4 annotations found: %s", annotationAutodetectionV4Cidrs)
		}
		config.autodetectionV4 = &autodetection{
			CIDRs: strings.Split(v, ","),
		}
	}

	if v, ok := annotations.Get(annotationAutodetectionV6Firstfound); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6Firstfound)
		}
		config.autodetectionV6 = &autodetection{
			FirstFound: v == "true",
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV6Kubernetes); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6Kubernetes)
		}
		config.autodetectionV6 = &autodetection{
			Kubernetes: v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV6Interface); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6Interface)
		}
		config.autodetectionV6 = &autodetection{
			InterfaceRegex: v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV6SkipInterface); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6SkipInterface)
		}
		config.autodetectionV6 = &autodetection{
			SkipInterfaceRegex: v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV6CanReach); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6CanReach)
		}
		config.autodetectionV6 = &autodetection{
			CanReach: v,
		}
	}
	if v, ok := annotations.Get(annotationAutodetectionV6Cidrs); ok {
		if config.autodetectionV6 != nil {
			return config, fmt.Errorf("multiple autodetection-v6 annotations found: %s", annotationAutodetectionV6Cidrs)
		}
		config.autodetectionV6 = &autodetection{
			CIDRs: strings.Split(v, ","),
		}
	}

	return config, nil
}
