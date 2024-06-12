package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/onsi/gomega"
)

func TestExtraNodeConfigFiles(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		expectErr  bool
		errMessage string
	}{
		{
			name: "ValidFiles",
			files: map[string]string{
				"config1": "content1",
				"config2": "content2",
			},
			expectErr: false,
		},
		{
			name: "InvalidFilename",
			files: map[string]string{
				"invalid/config": "content",
			},
			expectErr:  true,
			errMessage: "file name \"invalid/config\" must not contain any slashes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)

			tmpDir := t.TempDir()
			snap := &mock.Snap{
				Mock: mock.Mock{
					ServiceExtraConfigDir: tmpDir,
					UID:                   os.Getuid(),
					GID:                   os.Getgid(),
				},
			}

			err := ExtraNodeConfigFiles(snap, tt.files)
			if tt.expectErr {
				g.Expect(err).To(gomega.HaveOccurred())
				g.Expect(err.Error()).To(gomega.ContainSubstring(tt.errMessage))
			} else {
				g.Expect(err).ToNot(gomega.HaveOccurred())

				for filename, content := range tt.files {
					filePath := filepath.Join(tmpDir, filename)

					// Verify the file exists
					info, err := os.Stat(filePath)
					g.Expect(err).ToNot(gomega.HaveOccurred())
					g.Expect(info.Mode().Perm()).To(gomega.Equal(os.FileMode(0400)))

					// Verify the file content
					actualContent, err := os.ReadFile(filePath)
					g.Expect(err).ToNot(gomega.HaveOccurred())
					g.Expect(string(actualContent)).To(gomega.Equal(content))
				}
			}
		})
	}
}
