package k8s

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cmdutil "github.com/canonical/k8s/cmd/util"

	. "github.com/onsi/gomega"
)

func TestReadAndParseConfigFile(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		files          map[string]string
		expectedOutput string
		expectErr      bool
		errMessage     string
	}{
		{
			name: "ValidConfigWithoutExtraFiles",
			configContent: `
key1: value1
key2: value2
`,
			expectedOutput: "key1: value1\nkey2: value2\n",
			expectErr:      false,
		},
		{
			name: "ValidConfigWithExtraFiles",
			configContent: `
key1: value1
extra-node-config-files:
  - $TMPDIR/file1
  - $TMPDIR/file2
`,
			files: map[string]string{
				"file1": "content of file1",
				"file2": "content of file2",
			},
			expectedOutput: "extra-node-config-files:\n- content of file1\n- content of file2\nkey1: value1\n",
			expectErr:      false,
		},
		{
			name: "FileReadError",
			configContent: `
key1: value1
extra-node-config-files:
  - missing_file
`,
			expectErr:  true,
			errMessage: "failed to read extra node config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			tmpDir := t.TempDir()
			configFilePath := filepath.Join(tmpDir, "config.yaml")

			// Write the config content to the config file
			configContent := strings.ReplaceAll(tt.configContent, "$TMPDIR", tmpDir)
			err := os.WriteFile(configFilePath, []byte(configContent), 0600)
			g.Expect(err).ToNot(HaveOccurred())

			// Write the extra files
			for filename, content := range tt.files {
				err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0600)
				g.Expect(err).ToNot(HaveOccurred())
			}

			env := cmdutil.ExecutionEnvironment{
				Stdin: os.Stdin,
			}

			result, err := readAndParseConfigFile(env, configFilePath)
			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tt.errMessage))
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(result).To(Equal(tt.expectedOutput))
			}
		})
	}
}
