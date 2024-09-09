package cmdutil

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

const initialProcesEnvironmentVariables = "/proc/1/environ"

// paths to validate if root in the owner
var pathsOwnershipCheck = []string{"/sys", "/proc", "/dev/kmsg"}

// HookVerifyResources checks ownership of dirs required for k8s to run.
// HookVerifyResources validates AppArmor configurations.
// If potential issue found pops up warning.
func HookVerifyResources() func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		var warnList []string

		// check ownership of required dirs
		for _, path := range pathsOwnershipCheck {
			if msg, err := validateRootOwnership(path); err != nil {
				cmd.PrintErrf(err.Error())
			} else if len(msg) > 0 {
				warnList = append(warnList, msg)
			}
		}

		// check App Armor
		if armor, err := checkAppArmor(); err != nil {
			cmd.PrintErr(err.Error())
		} else if len(armor) > 0 {
			warnList = append(warnList, armor)
		}

		// check LXD
		if lxd, err := validateLXD(); err != nil {
			cmd.PrintErr(err.Error())
		} else if len(warnList) > 0 && len(lxd) > 0 {
			warnList = append(warnList, lxd)

		}

		// generate report
		if len(warnList) > 0 {
			cmd.PrintErrf("Warning: k8s may not run correctly due to reasons:\n%s",
				strings.Join(warnList, ""))
		}
	}
}

// validateLXD checks if k8s runs in lxd container if so returns link to documentation
func validateLXD() (string, error) {
	dat, err := os.ReadFile(initialProcesEnvironmentVariables)
	if err != nil {
		return "", err
	}
	env := string(dat)
	if strings.Contains(env, "container=lxc") {
		return "For running k8s inside LXD container refer to " +
			"https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/howto/install/lxd/.\n", nil
	}
	return "", nil
}

// validateRootOwnership checks if given path owner root and root group.
func validateRootOwnership(path string) (string, error) {

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("%s do not exist.\n", path), nil
		} else {
			return "", err
		}
	}
	var UID int
	var GID int
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		UID = int(stat.Uid)
		GID = int(stat.Gid)
	} else {
		return "", errors.New(fmt.Sprintf("cannot access path %s", path))
	}
	var warnList string
	if UID != 0 {
		warnList += fmt.Sprintf("owner of %s is user with UID %d expected 0.\n", path, UID)
	}
	if GID != 0 {
		warnList += fmt.Sprintf("owner of %s is group with GID %d expected 0.\n", path, GID)
	}
	return warnList, nil
}

// checkAppArmor checks AppArmor status.
func checkAppArmor() (string, error) {
	cmd := exec.Command("journalctl", "-u", "apparmor")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	output := string(out)
	// AppArmor configured for container or service not present
	if strings.Contains(output, "Not starting AppArmor in container") || strings.Contains(output, "-- No entries --") {
		return "", nil
		// cannot read status of AppArmor
	} else if strings.Contains(output, "Users in groups 'adm', 'systemd-journal' can see all messages.") {
		return "could not validate AppArmor status.\n", nil
	}

	return "AppArmor may block hosting of nested containers.\n", nil
}
