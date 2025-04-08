package cilium

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations/cilium"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

const (
	// minVLANIDValue is the minimum valid 802.1Q VLAN ID value.
	minVLANIDValue = 0
	// maxVLANIDValue is the maximum valid 802.1Q VLAN ID value.
	maxVLANIDValue = 4094
)

type config struct {
	devices             string
	directRoutingDevice string
	vlanBPFBypass       []int
	cniExclusive        bool
	sctpEnabled         bool
	tunnelPort          int
}

func validatePort(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, errors.New("invalid port: not a number")
	}
	if port < 1 || port > 65535 {
		return 0, errors.New("invalid port: out of range")
	}
	return port, nil
}

func validateVLANBPFBypass(vlanList string) ([]int, error) {
	vlanList = strings.TrimSpace(vlanList)
	// Maintain compatibility with the Cilium chart definition
	vlanList = strings.Trim(vlanList, "{}")
	vlans := strings.Split(vlanList, ",")

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
			continue
		}
		seenTags[vlanID] = struct{}{}
		vlanTags = append(vlanTags, vlanID)
	}

	slices.Sort(vlanTags)
	return vlanTags, nil
}

func internalConfig(annotations types.Annotations) (config, error) {
	c := config{}

	if v, ok := annotations.Get(apiv1_annotations.AnnotationDevices); ok {
		c.devices = v
	}

	if v, ok := annotations.Get(apiv1_annotations.AnnotationDirectRoutingDevice); ok {
		c.directRoutingDevice = v
	}

	if v, ok := annotations[apiv1_annotations.AnnotationVLANBPFBypass]; ok {
		vlanTags, err := validateVLANBPFBypass(v)
		if err != nil {
			return config{}, fmt.Errorf("failed to parse VLAN BPF bypass list: %w", err)
		}
		c.vlanBPFBypass = vlanTags
	}

	if _, ok := annotations.Get(apiv1_annotations.AnnotationCNIExclusive); ok {
		c.cniExclusive = true
	}

	if _, ok := annotations.Get(apiv1_annotations.AnnotationSCTPEnabled); ok {
		c.sctpEnabled = true
	}

	if v, ok := annotations.Get(apiv1_annotations.AnnotationTunnelPort); ok {
		tunnelPort, err := validatePort(v)
		if err != nil {
			return config{}, fmt.Errorf("failed to parse Tunnel encapsulation port: %w", err)
		}

		if tunnelPort == 0 {
			tunnelPort = ciliumDefaultVXLANPort
		}

		c.tunnelPort = tunnelPort
	}

	return c, nil
}
