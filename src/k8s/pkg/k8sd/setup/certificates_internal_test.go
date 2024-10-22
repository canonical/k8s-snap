package setup

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func TestEnsureFile(t *testing.T) {
	t.Run("CreateFile", func(t *testing.T) {
		g := NewWithT(t)

		tempDir := t.TempDir()
		fname := filepath.Join(tempDir, "test")
		updated, err := ensureFile(fname, "test", os.Getuid(), os.Getgid(), 0o777)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(updated).To(BeTrue())

		createdFile, _ := os.ReadFile(fname)
		g.Expect(string(createdFile)).To(Equal("test"))
	})

	t.Run("DeleteFile", func(t *testing.T) {
		g := NewWithT(t)
		tempDir := t.TempDir()
		fname := filepath.Join(tempDir, "test")

		// Create a file with some content.
		updated, err := ensureFile(fname, "test", os.Getuid(), os.Getgid(), 0o777)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(updated).To(BeTrue())

		// Delete the file.
		updated, err = ensureFile(fname, "", os.Getuid(), os.Getgid(), 0o777)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(updated).To(BeTrue())

		_, err = os.Stat(fname)
		g.Expect(os.IsNotExist(err)).To(BeTrue())
	})

	t.Run("ChangeContent", func(t *testing.T) {
		g := NewWithT(t)
		tempDir := t.TempDir()
		fname := filepath.Join(tempDir, "test")

		// Create a file with some content.
		updated, err := ensureFile(fname, "test", os.Getuid(), os.Getgid(), 0o777)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(updated).To(BeTrue())

		// ensureFile with same content should return that the file was not updated.
		updated, err = ensureFile(fname, "test", os.Getuid(), os.Getgid(), 0o777)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(updated).To(BeFalse())

		// Change the content and ensureFile should return that the file was updated.
		updated, err = ensureFile(fname, "new content", os.Getuid(), os.Getgid(), 0o777)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(updated).To(BeTrue())

		createdFile, _ := os.ReadFile(fname)
		g.Expect(string(createdFile)).To(Equal("new content"))

		// Change permissions and ensureFile should return that the file was not updated.
		updated, err = ensureFile(fname, "new content", os.Getuid(), os.Getgid(), 0o666)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(updated).To(BeFalse())
	})

	t.Run("NotExist", func(t *testing.T) {
		g := NewWithT(t)
		tempDir := t.TempDir()
		fname := filepath.Join(tempDir, "test")

		// ensureFile on inexistent file with empty content should return that the file was not updated.
		updated, err := ensureFile(fname, "", os.Getuid(), os.Getgid(), 0o777)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(updated).To(BeFalse())
	})
}
