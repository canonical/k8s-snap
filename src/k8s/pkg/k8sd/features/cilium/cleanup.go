package cilium

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/canonical/k8s/pkg/snap"
)

func CleanupNetwork(ctx context.Context, snap snap.Snap) error {
	if err := os.Remove("/var/run/cilium/cilium.pid"); err != nil {
		return fmt.Errorf("failed to remove cilium pid file: %w", err)
	}

	if _, err := os.Stat("/opt/cni/bin/cilium-dbg"); err == nil {
		if err := exec.CommandContext(ctx, "/opt/cni/bin/cilium-dbg", "cleanup", "--all-state", "--force").Run(); err != nil {
			return fmt.Errorf("cilium-dbg cleanup failed: %w", err)
		}
	} else {
		return fmt.Errorf("cilium-dbg can't be accessed: %w", err)
	}

	return nil
}
