package setup

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/cert"
)

// InitFolders creates the necessary folders for service arguments and certificates.
func InitFolders() error {
	argsDir := utils.DataPath("args")
	err := os.MkdirAll(argsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create arguments directory: %w", err)
	}

	err = os.MkdirAll(cert.KubePkiPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create pki directory: %w", err)
	}

	return nil
}
