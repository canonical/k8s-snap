package snaputil

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

func argumentsFileForService(s snap.Snap, serviceName string) string {
	return filepath.Join(s.ServiceArgumentsDir(), serviceName)
}

// GetServiceArgument retrieves the value of a specific argument from the $SNAP_DATA/args/$service file.
// The argument name should include preceding dashes (e.g. "--secure-port").
// If any errors occur, or the argument is not present, an empty string is returned.
func GetServiceArgument(s snap.Snap, serviceName string, argument string) (string, error) {
	arguments, err := os.ReadFile(argumentsFileForService(s, serviceName))
	if err != nil {
		return "", fmt.Errorf("failed to read arguments file for service %s: %w", serviceName, err)
	}

	for _, line := range strings.Split(string(arguments), "\n") {
		line = strings.TrimSpace(line)
		// ignore empty lines
		if line == "" {
			continue
		}
		if key, value := utils.ParseArgumentLine(line); key == argument {
			return value, nil
		}
	}
	return "", nil
}

// UpdateServiceArguments updates the arguments file for a service.
// UpdateServiceArguments is a no-op if updateList and delete are empty.
// updateList is a map of key-value pairs. It will replace the argument with the new value (or just append).
// delete is a list of arguments to remove completely. The argument is removed if present.
// Returns a boolean whether any of the arguments were changed, as well as any errors that may have occurred.
func UpdateServiceArguments(snap snap.Snap, serviceName string, updateMap map[string]string, deleteList []string) (bool, error) {
	if updateMap == nil {
		updateMap = map[string]string{}
	}
	if deleteList == nil {
		deleteList = []string{}
	}

	// If no updates are requested, exit early
	if len(updateMap) == 0 && len(deleteList) == 0 {
		return false, nil
	}

	deleteMap := make(map[string]struct{}, len(deleteList))
	for _, k := range deleteList {
		deleteMap[k] = struct{}{}
	}

	argumentsFile := argumentsFileForService(snap, serviceName)
	arguments, err := os.ReadFile(argumentsFile)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to read arguments file for service %s: %w", serviceName, err)
	}

	changed := false
	existingArguments := map[string]struct{}{}
	newArguments := []string{}
	for _, line := range strings.Split(string(arguments), "\n") {
		line = strings.TrimSpace(line)
		// ignore empty lines
		if line == "" {
			continue
		}
		key, oldValue := utils.ParseArgumentLine(line)
		existingArguments[key] = struct{}{}
		if newValue, ok := updateMap[key]; ok {
			// update argument with new value
			newArguments = append(newArguments, fmt.Sprintf("%s=%s", key, newValue))
			if oldValue != newValue {
				changed = true
			}
		} else if _, ok := deleteMap[key]; ok {
			// remove argument
			changed = true
			continue
		} else {
			// no change
			newArguments = append(newArguments, line)
		}
	}

	for key, value := range updateMap {
		if _, argExists := existingArguments[key]; !argExists {
			changed = true
			newArguments = append(newArguments, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// sort arguments so that output is consistent
	sort.Strings(newArguments)

	if err := os.WriteFile(argumentsFile, []byte(strings.Join(newArguments, "\n")+"\n"), 0600); err != nil {
		return false, fmt.Errorf("failed to write arguments for service %s: %w", serviceName, err)
	}
	return changed, nil
}
