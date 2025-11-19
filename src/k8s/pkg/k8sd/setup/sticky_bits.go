package setup

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/canonical/k8s/pkg/log"
)

// ApplyStickyBitsToWorldWritableDirectories finds all world-writable directories
// without the sticky bit set and adds it. This prevents users from deleting files
// owned by other users in shared directories.
func ApplyStickyBitsToWorldWritableDirectories(ctx context.Context) error {
	log := log.FromContext(ctx).WithValues("func", "ApplyStickyBitsToWorldWritableDirectories")

	// The command finds all local filesystems, then searches each one for directories
	// that are world-writable (-perm -0002) but don't have the sticky bit (-perm -1000).
	// It then adds the sticky bit (+t) to those directories.
	cmd := exec.CommandContext(ctx, "bash", "-c",
		`df --local -P | awk '{if (NR!=1) print $6}' | xargs -I '$6' find '$6' -xdev -type d \( -perm -0002 -a ! -perm -1000 \) 2>/dev/null -exec chmod a+t {} +`)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err, "Failed to apply sticky bits to world-writable directories", "output", string(output))
		return fmt.Errorf("failed to apply sticky bits: %w", err)
	}

	log.Info("Successfully applied sticky bits to world-writable directories")
	return nil
}
