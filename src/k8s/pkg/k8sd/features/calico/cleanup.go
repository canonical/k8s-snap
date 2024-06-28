package calico

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
	"golang.org/x/sys/unix"
)

func CleanupNetwork(ctx context.Context, snap snap.Snap) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("failed to list network interfaces: %w", err)
	}

	// Compile the regular expression outside the loop
	regex, err := regexp.Compile("^vxlan[-v6]*.calico|cali[a-f0-9]*|tunl[0-9]*$")
	if err != nil {
		return fmt.Errorf("failed to compile regex pattern: %w", err)
	}

	// Find the interfaces created by Calico
	for _, iface := range interfaces {
		// Check if the interface name matches the regex pattern
		// Adapted from MicroK8s' link removal hook:
		// https://github.com/canonical/microk8s/blob/dff3627959d4774198000795a0a0afcaa003324b/microk8s-resources/default-hooks/remove.d/10-cni-link#L15
		match := regex.MatchString(iface.Name)
		if match {
			// Perform cleanup for Calico interface
			if err := exec.CommandContext(ctx, "ip", "link", "delete", iface.Name).Run(); err != nil {
				return fmt.Errorf("failed to delete interface %s: %w", iface.Name, err)
			}
		}
	}

	// Delete network namespaces that start with "cali-"
	netnsDir := "/run/netns"
	entries, err := os.ReadDir(netnsDir)
	if err != nil {
		return fmt.Errorf("failed to list files under %s: %w", netnsDir, err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cali-") {
			nsPath := path.Join(netnsDir, entry.Name())

			if err := unix.Unmount(nsPath, unix.MNT_DETACH); err != nil {
				return fmt.Errorf("failed to unmount network namespace %s: %w", entry.Name(), err)
			}

			if err := os.Remove(nsPath); err != nil {
				return fmt.Errorf("failed to remove network namespace %s: %w", entry.Name(), err)
			}
		}
	}

	for _, cmd := range []string{"iptables", "ip6tables", "iptables-legacy", "ip6tables-legacy"} {
		out, err := exec.Command(fmt.Sprintf("%s-save", cmd)).Output()
		if err != nil {
			return fmt.Errorf("failed to read iptables rules: %w", err)
		}

		lines := strings.Split(string(out), "\n")
		for i, line := range lines {
			if strings.Contains(line, "cali") {
				lines[i] = ""
			}
		}

		restore := exec.Command(fmt.Sprintf("%s-restore", cmd))
		restore.Stdin = strings.NewReader(strings.Join(lines, "\n"))
		if err := restore.Run(); err != nil {
			return fmt.Errorf("failed to restore iptables rules: %w", err)
		}
	}

	return nil
}
