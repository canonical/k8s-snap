package setup

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
)

// ExtraNodeConfigFiles writes the file contents to the specified filenames in the snap.ExtraFilesDir directory.
// The files are created with 0400 permissions and owned by root.
// The filenames must not contain any slashes to prevent path traversal.
func ExtraNodeConfigFiles(snap snap.Snap, files map[string]string) error {
	for filename, content := range files {
		if strings.Contains(filename, "/") {
			return fmt.Errorf("file name %q must not contain any slashes (possible path-traversal prevented)", filename)
		}

		filePath := path.Join(snap.ServiceExtraConfigDir(), filename)
		// Create or truncate the file
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}
		defer file.Close()

		// Write the content to the file
		_, err = file.WriteString(content)
		if err != nil {
			return fmt.Errorf("failed to write to file %s: %w", filePath, err)
		}

		// Set file owner to root
		if err := os.Chown(filePath, snap.UID(), snap.GID()); err != nil {
			return fmt.Errorf("failed to change owner of file %s: %w", filePath, err)
		}

		if err := os.Chmod(filePath, 0400); err != nil {
			return fmt.Errorf("failed to change mode of file %s: %w", filePath, err)
		}
	}
	return nil
}
