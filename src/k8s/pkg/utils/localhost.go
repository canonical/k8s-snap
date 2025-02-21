package utils

import (
	"fmt"
	"net"

	mctypes "github.com/canonical/microcluster/v2/rest/types"
)

func DetermineLocalhostAddress(clusterMembers []mctypes.ClusterMember) (string, error) {
	// Check if any of the cluster members have an IPv6 address, if so return "::1"
	// if one member has an IPv6 address, other members should also have IPv6 interfaces
	for _, clusterMember := range clusterMembers {
		memberAddress := clusterMember.Address.Addr().String()
		nodeIP := net.ParseIP(memberAddress)
		if nodeIP == nil {
			return "", fmt.Errorf("failed to parse node IP address %q", memberAddress)
		}

		if nodeIP.To4() == nil {
			return "[::1]", nil
		}
	}

	// If no IPv6 addresses are found this means the cluster is IPv4 only
	return "127.0.0.1", nil
}
