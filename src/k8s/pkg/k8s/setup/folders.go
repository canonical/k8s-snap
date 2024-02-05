package setup

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/utils/cert"
)

// InitFolders creates the necessary folders for service arguments and certificates.
func InitFolders(argsDir string) error {
	err := os.MkdirAll(argsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create arguments directory: %w", err)
	}

	err = os.MkdirAll(cert.KubePkiPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create pki directory: %w", err)
	}

	if err := os.MkdirAll("/opt/cni/bin", 0700); err != nil {
		return fmt.Errorf("failed to create cni bin dir: %w", err)
	}

	if err := os.MkdirAll("/etc/cni/net.d", 0700); err != nil {
		return fmt.Errorf("failed to create cni conf dir: %w", err)
	}

	// TODO(neoaggelos): don't use a hardcoded path here
	if err := os.MkdirAll("/var/snap/k8s/common/etc/containerd", 0700); err != nil {
		return fmt.Errorf("failed to create cni bin dir: %w", err)
	}
	return nil
}
