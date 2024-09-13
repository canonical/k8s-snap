package k8s

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	cmdutil "github.com/canonical/k8s/cmd/util"

	"github.com/spf13/cobra"
)

// initialProcessEnvironmentVariables environment variables of initial process
const initialProcessEnvironmentVariables = "/proc/1/environ"

// pathsOwnershipCheck paths to validate root is the ownership
var pathsOwnershipCheck = []string{"/sys", "/proc", "/dev/kmsg"}

func chainPreRunHooks(hooks ...func(*cobra.Command, []string)) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		for _, hook := range hooks {
			hook(cmd, args)
		}
	}
}

func hookRequireRoot(env cmdutil.ExecutionEnvironment) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if env.Getuid() != 0 {
			cmd.PrintErrln("You do not have enough permissions. Please re-run the command with sudo.")
			env.Exit(1)
		}
	}
}

func hookInitializeFormatter(env cmdutil.ExecutionEnvironment, format *string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		// initialize formatter
		var err error
		outputFormatter, err = cmdutil.NewFormatter(*format, cmd.OutOrStdout())
		if err != nil {
			cmd.PrintErrf("Error: Unknown --output-format %q. It must be one of %q (default), %q or %q.", *format, "plain", "json", "yaml")
			env.Exit(1)
			return
		}
	}
}

// hookCheckLXD checks ownership of directories required for k8s to run.
// If potential issue found pops up warning for user.
func hookCheckLXD() func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if lxd, err := inLXDContainer(); lxd {
			var errMsgs []string
			for _, pathToCheck := range pathsOwnershipCheck {
				if err2 := validateRootOwnership(pathToCheck); err2 != nil {
					errMsgs = append(errMsgs, err2.Error())
				}
			}
			if len(errMsgs) > 0 {
				if debug, _ := cmd.Flags().GetBool("debug"); debug {
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

// getOwnership given path of file returns UID, GID and error.
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
	content, err := os.ReadFile(initialProcessEnvironmentVariables)
	if err != nil {
		// if permission to file is missing we still want to display info about lxd
		if os.IsPermission(err) {
			return true, err
		}
		return false, err
	}
	env := string(content)
	if strings.Contains(env, "container=lxc") {
		return true, nil
	}
	return false, nil
}
