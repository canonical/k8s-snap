package internal_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/snap/util/cleanup/internal"
	. "github.com/onsi/gomega"
)

func TestTryCleanupContainerdPaths_RemovesDirAndLockfile(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	lockFilesDir := filepath.Join(t.TempDir(), "lockfiles")
	err := os.MkdirAll(lockFilesDir, 0o755)
	g.Expect(err).To(Not(HaveOccurred()))

	t.Run("Removes all data directories and lock files", func(t *testing.T) {
		// Use t.TempDir() for all containerd-related directories to ensure test isolation
		containerdRootDir := filepath.Join(t.TempDir(), "containerd")
		containerdSocketDir := filepath.Join(t.TempDir(), "containerd_socket")
		containerdConfigDir := filepath.Join(t.TempDir(), "containerd_config")
		cniBinDir := filepath.Join(t.TempDir(), "cni_bin")

		// Mock snap with a lock files directory
		s := &mock.Snap{
			Mock: mock.Mock{
				LockFilesDir:        lockFilesDir,
				ContainerdRootDir:   containerdRootDir,
				ContainerdSocketDir: containerdSocketDir,
				ContainerdConfigDir: containerdConfigDir,
				CNIBinDir:           cniBinDir,
			},
		}

		// Create lock files for each path in the map
		m := map[string]string{
			"containerd-socket-path": s.ContainerdSocketDir(),
			"containerd-config-dir":  s.ContainerdConfigDir(),
			"containerd-root-dir":    s.ContainerdRootDir(),
			"containerd-cni-bin-dir": s.CNIBinDir(),
		}
		for name, dir := range m {
			lockFile := filepath.Join(lockFilesDir, name)
			f, err := os.Create(lockFile)
			g.Expect(err).To(Not(HaveOccurred()))
			_, err = f.WriteString(dir)
			g.Expect(err).To(Not(HaveOccurred()))
			f.Close()
			// Create the directory to simulate the data dir
			dataDir := dir
			err = os.MkdirAll(dataDir, 0o755)
			g.Expect(err).To(Not(HaveOccurred()))
		}

		internal.TryCleanupContainerdPaths(ctx, s)

		// All data dirs should be removed
		for _, dir := range m {
			_, err := os.Stat(dir)
			g.Expect(os.IsNotExist(err)).To(BeTrue())
		}
		// All lock files should be removed
		for name := range m {
			lockFile := filepath.Join(lockFilesDir, name)
			_, err := os.Stat(lockFile)
			g.Expect(os.IsNotExist(err)).To(BeTrue())
		}
	})

	// This is a dangerous test that ensures the cleanup function does not remove the root directory
	t.Run("Does not remove root directory", func(t *testing.T) {
		// Simulate a lockfile that (incorrectly) points to "/"
		lockFile := filepath.Join(lockFilesDir, "containerd-root-dir")
		err = os.WriteFile(lockFile, []byte("/"), 0o644)
		g.Expect(err).To(Not(HaveOccurred()))

		// Create a mock snap that returns "/" as the containerd root dir
		s := &mock.Snap{
			Mock: mock.Mock{
				LockFilesDir:      lockFilesDir,
				ContainerdRootDir: "/",
			},
		}

		// Call the cleanup function
		internal.TryCleanupContainerdPaths(ctx, s)

		// The lockfile should still exist, since the root dir must not be deleted
		_, err = os.Stat(lockFile)
		g.Expect(err).ToNot(HaveOccurred())

		// The root directory should still exist
		_, err = os.Stat("/")
		g.Expect(err).ToNot(HaveOccurred())
	})
}
