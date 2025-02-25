package utils

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetAllIPsFromInterface(t *testing.T) {
	ipv4 := "192.123.123.123"
	ipv4Other := "192.223.223.223"
	globalIPv6 := "2001:db8:abcd:1234:5678:90ab:cdef:1234"
	localIPv6 := "fe80::1a2b:3c4d:5e6f:7890"
	localIPv6Other := "fe80::1a2b:3c4d:5e6f:7899"

	commands := []string{
		"sudo ip link add foo type dummy",
		fmt.Sprintf("sudo ip addr add %s/24 dev foo", ipv4),
		fmt.Sprintf("sudo ip -6 addr add %s/64 dev foo", globalIPv6),
		fmt.Sprintf("sudo ip -6 addr add %s/64 dev foo", localIPv6),
		"sudo ip link add lish type dummy",
		fmt.Sprintf("sudo ip addr add %s/24 dev lish", ipv4Other),
		fmt.Sprintf("sudo ip -6 addr add %s/64 dev lish", localIPv6Other),
	}

	cleanupCommands := []string{
		"sudo ip link delete foo type dummy",
		"sudo ip link delete lish type dummy",
	}

	defer func() {
		for _, command := range cleanupCommands {
			arr := strings.Fields(command)
			cmd := exec.Command(arr[0], arr[1:]...)
			cmd.Run()
		}
	}()

	for _, command := range commands {
		arr := strings.Fields(command)
		cmd := exec.Command(arr[0], arr[1:]...)
		err := cmd.Run()
		if err != nil {
			t.Fatalf("Failed to run command '%s': %v", command, err)
		}
	}

	tests := []struct {
		name        string
		ipAddr      string
		expectError bool
		expectedIPs []string
	}{
		{
			name:        "lo ipv4",
			ipAddr:      "127.0.0.1",
			expectedIPs: []string{"127.0.0.1"},
		},
		{
			name:        "lo ipv6",
			ipAddr:      "::1",
			expectedIPs: []string{"::1"},
		},
		{
			name:        "ipv4/ipv6 pair from ipv4",
			ipAddr:      ipv4,
			expectedIPs: []string{ipv4, globalIPv6},
		},
		{
			name:        "ipv4/ipv6 pair from ipv6",
			ipAddr:      globalIPv6,
			expectedIPs: []string{globalIPv6, ipv4},
		},
		{
			name:        "ignore local scope ipv6",
			ipAddr:      ipv4Other,
			expectedIPs: []string{ipv4Other},
		},
		{
			name:        "address not found",
			ipAddr:      "8.8.8.8",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			ip := net.ParseIP(tc.ipAddr)

			ips, err := GetIPv46Addresses(ip)
			if tc.expectError {
				g.Expect(err).To(HaveOccurred())
				g.Expect(ips).To(BeEmpty())
				return
			}

			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(ips).To(HaveLen(len(tc.expectedIPs)))
			for i, expectedIP := range tc.expectedIPs {
				g.Expect(ips[i].String()).To(Equal(expectedIP))
			}
		})
	}
}
