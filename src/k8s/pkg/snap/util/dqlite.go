package snaputil

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/canonical/k8s/pkg/utils"
)

// NodeLabelToDqliteFailureDomain hashes (sha256) node labels to produce
// uint64 failure domain identifiers.
func NodeLabelToDqliteFailureDomain(label string) uint64 {
	sha256Sum := sha256.Sum256([]byte(label))
	// Select the first 8 bytes of the sha256 hash.
	return binary.LittleEndian.Uint64(sha256Sum[:])
}

// UpdateDqliteFailureDomain updates the failure domain of the dqlite database
// with the given state directory and returns a (boolean, error) tuple,
// specifying whether any changes were made. If the failure domain was modified,
// a service restart is required.
func UpdateDqliteFailureDomain(failureDomain uint64, dbStateDir string) (bool, error) {
	failureDomainStr := fmt.Sprintf("%v", failureDomain)
	failureDomainFile := filepath.Join(dbStateDir, "failure-domain")
	fileExists, err := utils.FileExists(failureDomainFile)
	if err != nil {
		return false, fmt.Errorf("unable to check if file exists %s: %w", failureDomainFile, err)
	}
	if !fileExists {
		var modified bool = failureDomain != 0
		err := os.WriteFile(failureDomainFile, []byte(failureDomainStr), 0644)
		if err != nil {
			return false, fmt.Errorf("failed to update failure-domain file %s: %w", failureDomainFile, err)
		}
		return modified, nil
	}

	contents, err := os.ReadFile(failureDomainFile)
	if err != nil {
		return false, fmt.Errorf("failed to read failure-domain file %s: %w", failureDomainFile, err)
	}
	existingFailureDomainStr := strings.Split(string(contents), "\n")[0]
	if existingFailureDomainStr == failureDomainStr {
		// Failure domain already set.
		return false, nil
	} else {
		// Updating failure domain.
		err := os.WriteFile(failureDomainFile, []byte(failureDomainStr), 0644)
		if err != nil {
			return false, fmt.Errorf("failed to update failure-domain file %s: %w", failureDomainFile, err)
		}
		return true, nil
	}
}
