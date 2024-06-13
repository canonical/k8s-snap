package cilium

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/canonical/k8s/pkg/snap"
)

func CleanupNetwork(ctx context.Context, snap snap.Snap) error {
	os.Remove("/var/run/cilium/cilium.pid")

	if _, err := os.Stat("/opt/cni/bin/cilium-dbg"); err == nil {
		if err := exec.CommandContext(ctx, "/opt/cni/bin/cilium-dbg", "cleanup", "--all-state", "--force").Run(); err != nil {
			return fmt.Errorf("cilium-dbg cleanup failed: %w", err)
		}
	}

	return nil
}
