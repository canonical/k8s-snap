package cilium

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations/cilium"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

const (
	// maxVLANTags is the maximum number of VLAN tags that can be configured
	maxVLANTags = 5
	// minVLANIDValue is the minimum valid 802.1Q VLAN ID value
	minVLANIDValue = 1
	// maxVLANIDValue is the maximum valid 802.1Q VLAN ID value
	maxVLANIDValue = 4094
)

type config struct {
	devices             string
	directRoutingDevice string
	vlanBPFBypass       []int
}

func validateVLANBPFBypass(vlanList string) ([]int, error) {
	vlanList = strings.TrimSpace(vlanList)
	// Maintain compatibility with the Cilium chart definition
	vlanList = strings.Trim(vlanList, "{}")
	vlans := strings.Split(vlanList, ",")

	// Special case: wildcard "0" allows all VLANs
	if len(vlans) == 1 && strings.TrimSpace(vlans[0]) == "0" {
		return []int{0}, nil
	}

	if len(vlans) > maxVLANTags {
		return []int{}, fmt.Errorf("the VLAN tag list cannot contain more than %d entries unless '0' is used to allow all VLANs", maxVLANTags)
	}

	vlanTags := make([]int, 0, len(vlans))
	seenTags := make(map[int]struct{})

	for _, vlan := range vlans {
		vlanID, err := strconv.Atoi(strings.TrimSpace(vlan))
		if err != nil {
			return []int{}, fmt.Errorf("failed to parse VLAN tag: %w", err)
		}
		if vlanID < minVLANIDValue || vlanID > maxVLANIDValue {
			return []int{}, fmt.Errorf("VLAN tag must be between 0 and %d", maxVLANIDValue)
		}

		if _, ok := seenTags[vlanID]; ok {
			return []int{}, fmt.Errorf("VLAN tag %d is duplicated", vlanID)
		}
		seenTags[vlanID] = struct{}{}
		vlanTags = append(vlanTags, vlanID)
	}

	slices.Sort(vlanTags)
	return vlanTags, nil
}

func internalConfig(annotations types.Annotations) (config, error) {
	c := config{}

	if v, ok := annotations[apiv1_annotations.AnnotationVLANBPFBypass]; ok {
		vlanTags, err := validateVLANBPFBypass(v)
		if err != nil {
			return config{}, fmt.Errorf("failed to parse VLAN BPF bypass list: %w", err)
		}
		c.vlanBPFBypass = vlanTags
	}
	if v, ok := annotations.Get(apiv1_annotations.AnnotationDevices); ok {
		c.devices = v
	}

	if v, ok := annotations.Get(apiv1_annotations.AnnotationDirectRoutingDevice); ok {
		c.directRoutingDevice = v
	}

	return c, nil
}
