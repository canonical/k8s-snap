package internal_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/snap/util/cleanup/internal"
	mountutils "github.com/canonical/k8s/pkg/utils/mount"
	. "github.com/onsi/gomega"
)

func TestRemoveLoopDevices(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	// Create a temp file to simulate /proc/mounts with loop device entries
	procMountsFile := filepath.Join(t.TempDir(), "mounts")
	mountsContent := `
/dev/loop123 /var/lib/kubelet/pods/abc ext4 rw,relatime 0 0
/dev/loop124 /not/kubelet/pods ext4 rw,relatime 0 0
/dev/sda1 /var/lib/kubelet/pods/def ext4 rw,relatime 0 0
`
	err := os.WriteFile(procMountsFile, []byte(mountsContent), 0o644)
	g.Expect(err).To(Not(HaveOccurred()))

	// Create a mock MountHelper
	mockHelper := mountutils.NewMockMountHelper(procMountsFile)

	internal.RemoveLoopDevices(ctx, mockHelper)

	// Read the file again to check that loop device entries are removed
	updatedContent, err := os.ReadFile(procMountsFile)
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(string(updatedContent)).ToNot(ContainSubstring("/dev/loop123"))
	// Ensure that the loop device that was not in kubelet pods is still present
	g.Expect(string(updatedContent)).To(ContainSubstring("/dev/loop124"))
	// Non-loop device entries should remain
	g.Expect(string(updatedContent)).To(ContainSubstring("/dev/sda1"))
}
