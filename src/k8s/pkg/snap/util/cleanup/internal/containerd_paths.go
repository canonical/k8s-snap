package internal

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
)

// TryCleanupContainerdPaths attempts to clean up all containerd directories which were
// created by the k8s-snap based on the existence of their respective lockfiles
// located in the directory returned by `s.LockFilesDir()`.
func TryCleanupContainerdPaths(ctx context.Context, s snap.Snap) {
	log := log.FromContext(ctx)

	snaputil.ForEachContainerdPath(ctx, s, func(lockpath string, dirpath string) error {
		log.Info("Cleaning up containerd data directory", "directory", dirpath)

		// Check directory exists before attempting to remove:
		stat, err := os.Stat(dirpath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("containerd data directory %q does not exist", dirpath)
			}
			return fmt.Errorf("failed to stat containerd data directory %q: %w", dirpath, err)
		}

		realPath := dirpath
		if stat.Mode()&fs.ModeSymlink != 0 {
			// NOTE(aznashwan): because of the convoluted interfaces-based way the snap
			// composes and creates the original lockfiles (see k8sd/setup/containerd.go)
			// this check is meant to defend against accidental code/configuration errors which
			// might lead to the root FS being deleted:
			realPath, err = os.Readlink(dirpath)
			if err != nil {
				return fmt.Errorf("failed to read link for directory %q: %w", dirpath, err)
			}
		}

		if realPath == "/" {
			return fmt.Errorf("there is some configuration/logic error in the current versions of the k8s-snap related to lockfile %q (meant to lock %q, which points to %q) which could lead to accidental wiping of the root file system", lockpath, dirpath, realPath)
		}

		if err := os.RemoveAll(dirpath); err != nil {
			// Avoid removing the lockfile path.
			return fmt.Errorf("failed to remove containerd data directory %q: %w", dirpath, err)
		}

		if err := os.Remove(lockpath); err != nil {
			return fmt.Errorf("failed to remove containerd lockfile %q: %w", lockpath, err)
		}

		return nil
	})
}
