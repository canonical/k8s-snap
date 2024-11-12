package cilium

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
)

func CleanupNetwork(ctx context.Context, snap snap.Snap) error {
	os.Remove("/var/run/cilium/cilium.pid")

	if _, err := os.Stat("/opt/cni/bin/cilium-dbg"); err == nil {
		if err := exec.CommandContext(ctx, "/opt/cni/bin/cilium-dbg", "cleanup", "--all-state", "--force").Run(); err != nil {
			return fmt.Errorf("cilium-dbg cleanup failed: %w", err)
		}
	}

	for _, cmd := range []string{"iptables", "ip6tables", "iptables-legacy", "ip6tables-legacy"} {
		out, err := exec.Command(fmt.Sprintf("%s-save", cmd)).Output()
		if err != nil {
			return fmt.Errorf("failed to read iptables rules: %w", err)
		}

		lines := strings.Split(string(out), "\n")
		for i, line := range lines {
			for _, word := range []string{"cilium", "kube", "CILIUM", "KUBE"} {
				if strings.Contains(line, word) {
					lines[i] = ""
					break
				}
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
