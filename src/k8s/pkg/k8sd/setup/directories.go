package setup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
)

// EnsureAllDirectories ensures all required configuration and state directories are created.
func EnsureAllDirectories(snap snap.Snap) error {
	if err := ensureCniBinDir(snap.CNIBinDir()); err != nil {
		return err
	}

	for _, dir := range []string{
		snap.CNIConfDir(),
		snap.ContainerdConfigDir(),
		snap.ContainerdExtraConfigDir(),
		snap.ContainerdRegistryConfigDir(),
		snap.K8sDqliteStateDir(),
		snap.KubernetesConfigDir(),
		snap.KubernetesPKIDir(),
		snap.EtcdPKIDir(),
		snap.LockFilesDir(),
		snap.ServiceArgumentsDir(),
		snap.ServiceExtraConfigDir(),
	} {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("failed to create required directory: %w", err)
		}
	}
	return nil
}

// Ensures that the provided path is a directory with the appropriate
// ownership/permissions for it to be used as the CNI binary directory.
// https://github.com/canonical/k8s-snap/issues/567
// https://github.com/cilium/cilium/issues/23838
func ensureCniBinDir(cniBinDir string) error {
	l := log.L().WithValues("cniBinDir", cniBinDir)
	if cniBinDir == "" {
		l.V(1).Info("Skipping creation of cni bin directory since it was not set")
		return nil
	}

	var stat syscall.Stat_t
	if err := syscall.Stat(cniBinDir, &stat); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to syscall.Stat(%q): %w", cniBinDir, err)
		}

		l.Info("Creating cni bin directory")
		if err := os.MkdirAll(cniBinDir, 0o0700); err != nil {
			return fmt.Errorf("failed to os.MkdirAll(%s): %w", cniBinDir, err)
		}

		if err := syscall.Stat(cniBinDir, &stat); err != nil {
			return fmt.Errorf("failed to syscall.Stat(%q) newly-created cni bin dir: %w", cniBinDir, err)
		}
	}

	if stat.Uid != 0 || stat.Gid != 0 {
		l.Info("Ensuring ownership of cni bin directory")
		if err := os.Chown(cniBinDir, 0, 0); err != nil {
			return fmt.Errorf("failed to os.Chown(%q, 0, 0): %w", cniBinDir, err)
		}
	}

	if (stat.Mode & 0o700) != 0o700 {
		l.Info("Ensuring permissions of cni bin directory")
		mode := os.FileMode(stat.Mode | 0o700)
		if err := os.Chmod(cniBinDir, mode); err != nil {
			return fmt.Errorf("failed to os.Chmod(%q, %o): %w", cniBinDir, mode, err)
		}
	}

	f, err := os.CreateTemp(cniBinDir, "test*.txt")
	if err != nil {
		return fmt.Errorf("failed create file in %q: %w", cniBinDir, err)
	}
	defer os.Remove(f.Name())

	return nil
}
