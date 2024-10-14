package cmdutil

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

// getFileOwnerAndGroup retrieves the UID and GID of a file.
func getFileOwnerAndGroup(filePath string) (uid, gid uint32, err error) {
	// Get file info using os.Stat
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting file info: %w", err)
	}
	// Convert the fileInfo.Sys() to syscall.Stat_t to access UID and GID
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, fmt.Errorf("failed to cast to syscall.Stat_t")
	}
	// Return the UID and GID
	return stat.Uid, stat.Gid, nil
}

// ValidateRootOwnership checks if the specified path is owned by the root user and root group.
func ValidateRootOwnership(path string) (err error) {
	uid, gid, err := getFileOwnerAndGroup(path)
	if err != nil {
		return err
	}
	if uid != 0 {
		return fmt.Errorf("owner of %s is user with UID %d expected 0", path, uid)
	}
	if gid != 0 {
		return fmt.Errorf("owner of %s is group with GID %d expected 0", path, gid)
	}
	return nil
}

// InLXDContainer checks if k8s runs in a lxd container.
func InLXDContainer() (isLXD bool, err error) {
	initialProcessEnvironmentVariables := "/proc/1/environ"
	content, err := os.ReadFile(initialProcessEnvironmentVariables)
	if err != nil {
		// if the permission to file is missing we still want to display info about lxd
		if os.IsPermission(err) {
			return true, fmt.Errorf("cannnot access %s to check if runing in LXD container: %w", initialProcessEnvironmentVariables, err)
		}
		return false, fmt.Errorf("cannnot read %s to check if runing in LXD container: %w", initialProcessEnvironmentVariables, err)
	}
	if strings.Contains(string(content), "container=lxc") {
		return true, nil
	}
	return false, nil
}
