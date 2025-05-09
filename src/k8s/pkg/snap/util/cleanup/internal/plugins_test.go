package internal_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/snap/util/cleanup/internal"
	. "github.com/onsi/gomega"
)

func TestRemovePluginSockets(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	// Create temporary plugin directories
	tmpDir := t.TempDir()
	pluginDir1 := filepath.Join(tmpDir, "plugins")
	pluginDir2 := filepath.Join(tmpDir, "plugins_registry")
	err := os.MkdirAll(pluginDir1, 0o755)
	g.Expect(err).To(Not(HaveOccurred()))
	err = os.MkdirAll(pluginDir2, 0o755)
	g.Expect(err).To(Not(HaveOccurred()))

	// Create subdirectories
	subDir1 := filepath.Join(pluginDir1, "subdir1")
	subDir2 := filepath.Join(pluginDir2, "subdir2")
	err = os.MkdirAll(subDir1, 0o755)
	g.Expect(err).To(Not(HaveOccurred()))
	err = os.MkdirAll(subDir2, 0o755)
	g.Expect(err).To(Not(HaveOccurred()))

	// Patch the global pluginDirs variable for test isolation
	origPluginDirs := internal.PluginDirs
	internal.PluginDirs = []string{pluginDir1, pluginDir2}
	defer func() { internal.PluginDirs = origPluginDirs }()

	// Create some .sock files and some non-sock files
	sock1 := filepath.Join(pluginDir1, "test1.sock")
	sock2 := filepath.Join(pluginDir2, "test2.sock")
	sock3 := filepath.Join(subDir1, "nested1.sock")
	sock4 := filepath.Join(subDir2, "nested2.sock")
	other1 := filepath.Join(pluginDir1, "notasocket.txt")
	other2 := filepath.Join(pluginDir2, "anotherfile.conf")
	other3 := filepath.Join(subDir1, "notasocket3.txt")
	other4 := filepath.Join(subDir2, "notasocket4.conf")
	err = os.WriteFile(sock1, []byte("socket1"), 0o644)
	g.Expect(err).To(Not(HaveOccurred()))
	err = os.WriteFile(sock2, []byte("socket2"), 0o644)
	g.Expect(err).To(Not(HaveOccurred()))
	err = os.WriteFile(sock3, []byte("socket3"), 0o644)
	g.Expect(err).To(Not(HaveOccurred()))
	err = os.WriteFile(sock4, []byte("socket4"), 0o644)
	g.Expect(err).To(Not(HaveOccurred()))
	err = os.WriteFile(other1, []byte("notasocket"), 0o644)
	g.Expect(err).To(Not(HaveOccurred()))
	err = os.WriteFile(other2, []byte("notasocket2"), 0o644)
	g.Expect(err).To(Not(HaveOccurred()))
	err = os.WriteFile(other3, []byte("notasocket3"), 0o644)
	g.Expect(err).To(Not(HaveOccurred()))
	err = os.WriteFile(other4, []byte("notasocket4"), 0o644)
	g.Expect(err).To(Not(HaveOccurred()))

	// Call the function under test
	internal.RemovePluginSockets(ctx)

	// .sock files should be removed (including in subdirectories)
	_, err = os.Stat(sock1)
	g.Expect(os.IsNotExist(err)).To(BeTrue())
	_, err = os.Stat(sock2)
	g.Expect(os.IsNotExist(err)).To(BeTrue())
	_, err = os.Stat(sock3)
	g.Expect(os.IsNotExist(err)).To(BeTrue())
	_, err = os.Stat(sock4)
	g.Expect(os.IsNotExist(err)).To(BeTrue())

	// Other files should remain
	_, err = os.Stat(other1)
	g.Expect(err).ToNot(HaveOccurred())
	_, err = os.Stat(other2)
	g.Expect(err).ToNot(HaveOccurred())
	_, err = os.Stat(other3)
	g.Expect(err).ToNot(HaveOccurred())
	_, err = os.Stat(other4)
	g.Expect(err).ToNot(HaveOccurred())
}
