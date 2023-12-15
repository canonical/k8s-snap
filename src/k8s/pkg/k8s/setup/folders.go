package setup

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/k8s/certutils"
	"github.com/canonical/k8s/pkg/snap"
)

// InitFolders creates the necessary folders for service arguments and certificates.
func InitFolders() error {
	argsDir := snap.DataPath("args")
	err := os.MkdirAll(argsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create arguments directory: %w", err)
	}

	err = os.MkdirAll(certutils.KubePkiPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create pki directory: %w", err)
	}

	return nil
}
