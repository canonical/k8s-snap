package snaputil

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

// NodeLabelToDqliteFailureDomain hashes (sha256) node labels to produce
// uint64 failure domain identifiers.
func NodeLabelToDqliteFailureDomain(label string) uint64 {
	sha256Sum := sha256.Sum256([]byte(label))
	// Select the first 8 bytes of the sha256 hash.
	return binary.LittleEndian.Uint64(sha256Sum[:])
}

func UpdateDqliteFailureDomain(snap snap.Snap, failureDomain uint64) (bool, error) {
	// We need to update both k8s-snap Dqlite databases (k8sd and k8s-dqlite).
	k8sDqliteStateDir := snap.K8sDqliteStateDir()
	k8sdDbStateDir := filepath.Join(snap.K8sdStateDir(), "database")

	k8sDqliteModified, err := updateDbFailureDomain(failureDomain, k8sDqliteStateDir)
	if err != nil {
		return false, err
	}

	k8sdModified, err := updateDbFailureDomain(failureDomain, k8sdDbStateDir)
	if err != nil {
		return false, err
	}

	return k8sDqliteModified || k8sdModified, nil
}

func updateDbFailureDomain(failureDomain uint64, dbStateDir string) (bool, error) {
	failureDomainStr := fmt.Sprintf("%v", failureDomain)
	failureDomainFile := filepath.Join(dbStateDir, "failure-domain")
	fileExists, err := utils.FileExists(failureDomainFile)
	if err != nil {
		return false, fmt.Errorf("unable to check if file exists %s: %w", failureDomainFile, err)
	}
	if !fileExists {
		// Failure domain not set, create the file.
		err := os.WriteFile(failureDomainFile, []byte(failureDomainStr), 0644)
		if err != nil {
			return false, fmt.Errorf("failed to update failure-domain file %s: %w", failureDomainFile, err)
		}
		return true, nil
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
