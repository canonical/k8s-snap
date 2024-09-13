package cmdutil

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

const initialProcesEnvironmentVariables = "/proc/1/environ"

// paths to validate if root in the owner
var pathsOwnershipCheck = []string{"/sys", "/proc", "/dev/kmsg"}

// HookCheckLXD checks ownership of dirs required for k8s to run.
// HookCheckLXD if verbose true prints out list of potenital issues.
// If potential issue found pops up warning for user.
func HookCheckLXD(verbose bool) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if lxd, err := inLXDContainer(); lxd {
			errMsgs := VerifyPaths()
			if len(errMsgs) > 0 {
				if verbose {
					cmd.PrintErrln("Warning: When validating required resources potential issues found:")
					for _, errMsg := range errMsgs {
						cmd.PrintErrln("\t", errMsg)
					}
				}
				cmd.PrintErrln("The lxc profile for Canonical Kubernetes might be missing.")
				cmd.PrintErrln("For running k8s inside LXD container refer to " +
					"https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/howto/install/lxd/")
			}
		} else if err != nil {
			cmd.PrintErrf(err.Error())
		}
	}
}

func VerifyPaths() []string {
	var errMsg []string
	// check ownership of required dirs
	for _, pathToCheck := range pathsOwnershipCheck {
		if err := validateRootOwnership(pathToCheck); err != nil {
			errMsg = append(errMsg, err.Error())
		}
	}
	return errMsg
}

func getOwnership(path string) (int, int, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, fmt.Errorf("%s do not exist", path)
		} else {
			return 0, 0, err
		}
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return int(stat.Uid), int(stat.Gid), nil
	} else {
		return 0, 0, fmt.Errorf("cannot access %s", path)
	}
}

// validateRootOwnership checks if given path owner root and root group.
func validateRootOwnership(path string) error {
	UID, GID, err := getOwnership(path)
	if err != nil {
		return err
	}
	if UID != 0 {
		return fmt.Errorf("owner of %s is user with UID %d expected 0", path, UID)
	}
	if GID != 0 {
		return fmt.Errorf("owner of %s is group with GID %d expected 0", path, GID)
	}
	return nil
}

// inLXDContainer checks if k8s runs in lxd container if so returns link to documentation
func inLXDContainer() (bool, error) {
	dat, err := os.ReadFile(initialProcesEnvironmentVariables)
	if err != nil {
		// if permission to file is missing we still want to display info about lxd
		if os.IsPermission(err) {
			return true, err
		}
		return false, err
	}
	env := string(dat)
	if strings.Contains(env, "container=lxc") {

		return true, nil
	}
	return false, nil
}
