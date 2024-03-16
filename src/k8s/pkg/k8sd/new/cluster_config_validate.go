package newtypes

import (
	"fmt"
	"net"
	"strings"
)

func (c *ClusterConfig) Validate() error {
	podCIDRs := strings.Split(c.Network.GetPodCIDR(), ",")
	if len(podCIDRs) != 1 && len(podCIDRs) != 2 {
		return fmt.Errorf("invalid number of pod CIDRs: %d", len(podCIDRs))
	}
	serviceCIDRs := strings.Split(c.Network.GetServiceCIDR(), ",")
	if len(serviceCIDRs) != 1 && len(serviceCIDRs) != 2 {
		return fmt.Errorf("invalid number of service CIDRs: %d", len(serviceCIDRs))
	}
	for _, cidr := range podCIDRs {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return fmt.Errorf("invalid pod CIDR %q: %w", cidr, err)
		}
	}
	for _, cidr := range serviceCIDRs {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return fmt.Errorf("invalid service CIDR %q: %w", cidr, err)
		}
	}

	return nil
}
