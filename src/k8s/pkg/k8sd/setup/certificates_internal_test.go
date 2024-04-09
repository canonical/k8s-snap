package setup

import (
	"os"
	"path"
	"testing"

	. "github.com/onsi/gomega"
)

func TestEnsureFile(t *testing.T) {

	t.Run("CreateFile", func(t *testing.T) {
		g := NewWithT(t)
		tempDir := t.TempDir()
		fname := path.Join(tempDir, "test")
		updated, err := ensureFile(fname, "test", 1000, 1000, 0777)
		g.Expect(err).To(BeNil())
		g.Expect(updated).To(BeTrue())

		createdFile, _ := os.ReadFile(fname)
		g.Expect(string(createdFile) == "test").To(BeTrue())
	})

	t.Run("DeleteFile", func(t *testing.T) {
		g := NewWithT(t)
		tempDir := t.TempDir()
		fname := path.Join(tempDir, "test")

		// Create a file with some content.
		updated, err := ensureFile(fname, "test", 1000, 1000, 0777)
		g.Expect(err).To(BeNil())
		g.Expect(updated).To(BeTrue())

		// Delete the file.
		updated, err = ensureFile(fname, "", 1000, 1000, 0777)
		g.Expect(err).To(BeNil())
		g.Expect(updated).To(BeTrue())

		_, err = os.Stat(fname)
		g.Expect(os.IsNotExist(err)).To(BeTrue())
	})

	t.Run("ChangeContent", func(t *testing.T) {
		g := NewWithT(t)
		tempDir := t.TempDir()
		fname := path.Join(tempDir, "test")

		// Create a file with some content.
		updated, err := ensureFile(fname, "test", 1000, 1000, 0777)
		g.Expect(err).To(BeNil())
		g.Expect(updated).To(BeTrue())

		// ensureFile with same content should return that the file was not updated.
		updated, err = ensureFile(fname, "test", 1000, 1000, 0777)
		g.Expect(err).To(BeNil())
		g.Expect(updated).To(BeFalse())

		// Change the content and ensureFile should return that the file was updated.
		updated, err = ensureFile(fname, "new content", 1000, 1000, 0777)
		g.Expect(err).To(BeNil())
		g.Expect(updated).To(BeTrue())

		createdFile, _ := os.ReadFile(fname)
		g.Expect(string(createdFile) == "new content").To(BeTrue())

		// Change permissions and ensureFile should return that the file was not updated.
		updated, err = ensureFile(fname, "new content", 1000, 1000, 0666)
		g.Expect(err).To(BeNil())
		g.Expect(updated).To(BeFalse())

		// TODO: test with a different guid/uid than 1000
	})
}
