package snaputil

import (
	"path/filepath"

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
