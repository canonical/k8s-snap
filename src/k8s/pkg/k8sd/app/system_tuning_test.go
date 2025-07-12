package app_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/app"
	. "github.com/onsi/gomega"
)

func TestGetHighestConfigFileOrder(t *testing.T) {
	g := NewWithT(t)

	tempDir := t.TempDir()
	file1 := "/000-custom.conf"
	file2 := "/010-user-file.conf"

	file1Path := filepath.Join(tempDir, file1)
	_, err := os.Create(file1Path)
	g.Expect(err).To(Not(HaveOccurred()))

	file2Path := filepath.Join(tempDir, file2)
	_, err = os.Create(file2Path)
	g.Expect(err).To(Not(HaveOccurred()))

	reConfFiles := regexp.MustCompile(`^(\d+)-.*\.conf$`)
	maxOrder, err := app.GetHighestConfigFileOrder(tempDir, reConfFiles)
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(maxOrder).To(Equal(10))
}
