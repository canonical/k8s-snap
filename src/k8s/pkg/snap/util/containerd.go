package snaputil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
)

// ContainerdLockPathsForSnap returns a mapping between the absolute paths of
// the lockfiles within the k8s snap and the absolute paths of the containerd
// directory they lock.
//
// WARN: these lockfiles are meant to be used in later cleanup stages.
// DO NOT include any system paths which are not managed by the k8s-snap!
//
// It intentionally does NOT include the containerd base dir lockfile
// (which most of the rest of the paths are based on), as it is meant
// to indicate the root of the containerd install ('/' or '/var/snap/k8s/*').
func ContainerdLockPathsForSnap(s snap.Snap) map[string]string {
	m := map[string]string{
		"containerd-socket-path": s.ContainerdSocketDir(),
		"containerd-config-dir":  s.ContainerdConfigDir(),
		"containerd-root-dir":    s.ContainerdRootDir(),
		"containerd-cni-bin-dir": s.CNIBinDir(),
	}

	prefixed := map[string]string{}
	for k, v := range m {
		prefixed[filepath.Join(s.LockFilesDir(), k)] = v
	}

	return prefixed
}

func ForEachContainerdPath(ctx context.Context, s snap.Snap, callback func(lockPath string, dirPath string) error) {
	log := log.FromContext(ctx)

	for lockpath, dirpath := range ContainerdLockPathsForSnap(s) {
		// Ensure the directory exists
		if err := func() error {
			// Ensure lockfile exists:
			if _, err := os.Stat(lockpath); os.IsNotExist(err) {
				return fmt.Errorf("failed to find containerd lockfile %q for directory %q", lockpath, dirpath)
			}

			// Ensure lockfile's contents is the one we expect:
			lockfile_contents := ""
			if contents, err := os.ReadFile(lockpath); err != nil {
				return fmt.Errorf("failed to read contents of lockfile %q: %w", lockpath, err)
			} else {
				lockfile_contents = string(contents)
			}

			if lockfile_contents != dirpath {
				return fmt.Errorf("lockfile %q points to different path than expected: %q != %q", lockpath, dirpath, lockfile_contents)
			}

			return nil
		}(); err != nil {
			log.Info("WARN: skipping containerd path due to error", "lockpath", lockpath, "dirpath", dirpath, "error", err)
			continue
		}

		if err := callback(lockpath, dirpath); err != nil {
			log.Error(err, "callback failed for containerd lock path and directory", "lockpath", lockpath, "dirpath", dirpath)
		}
	}
}

func IsContainerdPathManaged(ctx context.Context, s snap.Snap, hee string) bool {
	isManaged := false

	ForEachContainerdPath(ctx, s, func(lockpath string, dirpath string) error {
		// Check if the directory matches the given path
		if dirpath == hee {
			isManaged = true
		}
		return nil
	})

	return isManaged
}
