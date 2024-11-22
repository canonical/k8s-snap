package calico

import (
	"fmt"
	"strings"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations/calico"
	"github.com/canonical/k8s/pkg/k8sd/types"
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
			if strValue, ok := autodetectionValue.(string); ok {
				autodetectionValue = strings.Split(strValue, ",")
			} else {
				return nil, fmt.Errorf("invalid type for cidrs annotation: %T", autodetectionValue)
			}
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

	if v, ok := annotations.Get(apiv1_annotations.AnnotationAPIServerEnabled); ok {
		c.apiServerEnabled = v == "true"
	}

	if v, ok := annotations.Get(apiv1_annotations.AnnotationEncapsulationV4); ok {
		if err := checkEncapsulation(v); err != nil {
			return config{}, fmt.Errorf("invalid encapsulation-v4 annotation: %w", err)
		}
		c.encapsulationV4 = v
	}

	if v, ok := annotations.Get(apiv1_annotations.AnnotationEncapsulationV6); ok {
		if err := checkEncapsulation(v); err != nil {
			return config{}, fmt.Errorf("invalid encapsulation-v6 annotation: %w", err)
		}
		c.encapsulationV6 = v
	}

	v4Map := map[string]string{
		apiv1_annotations.AnnotationAutodetectionV4FirstFound:    "firstFound",
		apiv1_annotations.AnnotationAutodetectionV4Kubernetes:    "kubernetes",
		apiv1_annotations.AnnotationAutodetectionV4Interface:     "interface",
		apiv1_annotations.AnnotationAutodetectionV4SkipInterface: "skipInterface",
		apiv1_annotations.AnnotationAutodetectionV4CanReach:      "canReach",
		apiv1_annotations.AnnotationAutodetectionV4CIDRs:         "cidrs",
	}

	autodetectionV4, err := parseAutodetectionAnnotations(annotations, v4Map)
	if err != nil {
		return config{}, fmt.Errorf("error parsing autodetection-v4 annotations: %w", err)
	}

	if autodetectionV4 != nil {
		c.autodetectionV4 = autodetectionV4
	}

	v6Map := map[string]string{
		apiv1_annotations.AnnotationAutodetectionV6FirstFound:    "firstFound",
		apiv1_annotations.AnnotationAutodetectionV6Kubernetes:    "kubernetes",
		apiv1_annotations.AnnotationAutodetectionV6Interface:     "interface",
		apiv1_annotations.AnnotationAutodetectionV6SkipInterface: "skipInterface",
		apiv1_annotations.AnnotationAutodetectionV6CanReach:      "canReach",
		apiv1_annotations.AnnotationAutodetectionV6CIDRs:         "cidrs",
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
