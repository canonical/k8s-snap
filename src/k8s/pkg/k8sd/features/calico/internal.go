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

func parseAutodetectionAnnotations(annotations types.Annotations, autodetectionMap map[string]string) (map[string]any, error) {
	var autodetectionAnnotations []string
	var autodetectionKey string
	var autodetectionValue any

	for annotation, key := range autodetectionMap {
		if v, ok := annotations.Get(annotation); ok {
			autodetectionAnnotations = append(autodetectionAnnotations, annotation)
			autodetectionKey = key
			autodetectionValue = v
		}
	}

	if len(autodetectionAnnotations) > 1 {
		return nil, fmt.Errorf("multiple annotations found: %s", strings.Join(autodetectionAnnotations, ", "))
	}

	// If any annotation is set, return the map otherwise it's left nil
	if autodetectionKey != "" {
		switch autodetectionKey {
		case "firstFound":
			autodetectionValue = autodetectionValue == "true"
		case "cidrs":
			autodetectionValue = strings.Split(autodetectionValue.(string), ",")
		}

		return map[string]any{
			autodetectionKey: autodetectionValue,
		}, nil
	}

	return nil, nil
}

func internalConfig(annotations types.Annotations) (config, error) {
	c := config{
		encapsulationV4:  defaultEncapsulation,
		encapsulationV6:  defaultEncapsulation,
		apiServerEnabled: defaultAPIServerEnabled,
	}

	if v, ok := annotations.Get(annotationAPIServerEnabled); ok {
		c.apiServerEnabled = v == "true"
	}

	if v, ok := annotations.Get(annotationEncapsulationV4); ok {
		if err := checkEncapsulation(v); err != nil {
			return config{}, fmt.Errorf("invalid encapsulation-v4 annotation: %w", err)
		}
		c.encapsulationV4 = v
	}

	if v, ok := annotations.Get(annotationEncapsulationV6); ok {
		if err := checkEncapsulation(v); err != nil {
			return config{}, fmt.Errorf("invalid encapsulation-v6 annotation: %w", err)
		}
		c.encapsulationV6 = v
	}

	v4Map := map[string]string{
		annotationAutodetectionV4FirstFound:    "firstFound",
		annotationAutodetectionV4Kubernetes:    "kubernetes",
		annotationAutodetectionV4Interface:     "interface",
		annotationAutodetectionV4SkipInterface: "skipInterface",
		annotationAutodetectionV4CanReach:      "canReach",
		annotationAutodetectionV4CIDRs:         "cidrs",
	}

	autodetectionV4, err := parseAutodetectionAnnotations(annotations, v4Map)
	if err != nil {
		return config{}, fmt.Errorf("error parsing autodetection-v4 annotations: %w", err)
	}

	if autodetectionV4 != nil {
		c.autodetectionV4 = autodetectionV4
	}

	v6Map := map[string]string{
		annotationAutodetectionV6FirstFound:    "firstFound",
		annotationAutodetectionV6Kubernetes:    "kubernetes",
		annotationAutodetectionV6Interface:     "interface",
		annotationAutodetectionV6SkipInterface: "skipInterface",
		annotationAutodetectionV6CanReach:      "canReach",
		annotationAutodetectionV6CIDRs:         "cidrs",
	}

	autodetectionV6, err := parseAutodetectionAnnotations(annotations, v6Map)
	if err != nil {
		return config{}, fmt.Errorf("error parsing autodetection-v6 annotations: %w", err)
	}

	if autodetectionV6 != nil {
		c.autodetectionV6 = autodetectionV6
	}

	return c, nil
}
