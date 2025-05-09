package internal_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/snap/util/cleanup/internal"
	mountutils "github.com/canonical/k8s/pkg/utils/mount"
	. "github.com/onsi/gomega"
)

func TestRemoveVolumeMountsGracefully(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	lockFilesDir := filepath.Join(t.TempDir(), "lockfiles")
	err := os.MkdirAll(lockFilesDir, 0o755)
	g.Expect(err).To(Not(HaveOccurred()))
	// Mock snap with a lock files directory

	s := &mock.Snap{
		Mock: mock.Mock{
			LockFilesDir:        lockFilesDir,
			ContainerdRootDir:   "/var/lib/containerd",
			ContainerdSocketDir: "/run/containerd",
			ContainerdConfigDir: "/etc/containerd",
			CNIBinDir:           "/opt/cni/bin",
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
	}

	// Prepare a temp file to simulate /proc/mounts
	procMountsFile := filepath.Join(t.TempDir(), "mounts")
	mountsContent := `
/dev/sda1 /var/lib/kubelet/pods/abc ext4 rw,relatime 0 0
/dev/sda2 /var/lib/kubelet/pods/def nfs rw,relatime 0 0
/dev/sda3 /run/containerd/io.containerd.xzy ext4 rw,relatime 0 0
/dev/sda4 /var/lib/containerd/vol1 ext4 rw,relatime 0 0
/dev/sda5 /not/affected/path ext4 rw,relatime 0 0
`
	err = os.WriteFile(procMountsFile, []byte(mountsContent), 0o644)
	g.Expect(err).To(Not(HaveOccurred()))

	// Create a mock MountHelper that records unmounts
	mockHelper := mountutils.NewMockMountHelper(procMountsFile)

	internal.RemoveVolumeMountsGracefully(ctx, s, mockHelper)
	// Read the file again to check that the correct mount points are removed
	updatedContent, err := os.ReadFile(procMountsFile)
	g.Expect(err).To(Not(HaveOccurred()))
	contentStr := string(updatedContent)

	// Should remove all except the NFS and unrelated mount
	g.Expect(contentStr).ToNot(ContainSubstring("/var/lib/kubelet/pods/abc"))
	g.Expect(contentStr).ToNot(ContainSubstring("/run/containerd/io.containerd.xzy"))
	g.Expect(contentStr).ToNot(ContainSubstring("/var/lib/containerd/vol1"))
	// NFS mount should remain
	g.Expect(contentStr).To(ContainSubstring("/var/lib/kubelet/pods/def"))
	// Unrelated mount should remain
	g.Expect(contentStr).To(ContainSubstring("/not/affected/path"))
}

func TestRemoveVolumeMountsForce(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	lockFilesDir := filepath.Join(t.TempDir(), "lockfiles")
	err := os.MkdirAll(lockFilesDir, 0o755)
	g.Expect(err).To(Not(HaveOccurred()))
	// Mock snap with a lock files directory

	s := &mock.Snap{
		Mock: mock.Mock{
			LockFilesDir:        lockFilesDir,
			ContainerdRootDir:   "/var/lib/containerd",
			ContainerdSocketDir: "/run/containerd",
			ContainerdConfigDir: "/etc/containerd",
			CNIBinDir:           "/opt/cni/bin",
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
	}

	// Prepare a temp file to simulate /proc/mounts
	procMountsFile := filepath.Join(t.TempDir(), "mounts")
	mountsContent := `
/dev/sda1 /var/lib/kubelet/pods/abc ext4 rw,relatime 0 0
/dev/sda2 /run/containerd/io.containerd.xzy ext4 rw,relatime 0 0
/dev/sda3 /var/lib/kubelet/plugins/vol1 ext4 rw,relatime 0 0
/dev/sda4 /var/lib/containerd/vol2 ext4 rw,relatime 0 0
/dev/sda5 /not/affected/path ext4 rw,relatime 0 0
`
	err = os.WriteFile(procMountsFile, []byte(mountsContent), 0o644)
	g.Expect(err).To(Not(HaveOccurred()))

	// Create a mock MountHelper that records unmounts
	mockHelper := mountutils.NewMockMountHelper(procMountsFile)

	internal.RemoveVolumeMountsForce(ctx, s, mockHelper)

	// Read the file again to check that the correct mount points are removed
	updatedContent, err := os.ReadFile(procMountsFile)
	g.Expect(err).To(Not(HaveOccurred()))
	contentStr := string(updatedContent)

	// Should remove all matching the prefixes
	g.Expect(contentStr).ToNot(ContainSubstring("/var/lib/kubelet/pods/abc"))
	g.Expect(contentStr).ToNot(ContainSubstring("/run/containerd/io.containerd.xzy"))
	g.Expect(contentStr).ToNot(ContainSubstring("/var/lib/kubelet/plugins/vol1"))
	g.Expect(contentStr).ToNot(ContainSubstring("/var/lib/containerd/vol2"))
	// Unrelated mount should remain
	g.Expect(contentStr).To(ContainSubstring("/not/affected/path"))
}
