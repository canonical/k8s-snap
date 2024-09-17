package k8s

import (
	cmdutil "github.com/canonical/k8s/cmd/util"

	"github.com/spf13/cobra"
)

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

// hookCheckLXD verifies the ownership of directories needed for Kubernetes to function.
// If a potential issue is detected, it displays a warning to the user.
func hookCheckLXD() func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		// pathsOwnershipCheck paths to validate root is the owner
		var pathsOwnershipCheck = []string{"/sys", "/proc", "/dev/kmsg"}
		inLXD, err := cmdutil.InLXDContainer()
		if err != nil {
			cmd.PrintErrf("Failed to check if running inside LXD container: %w", err)
			return
		}
		if inLXD {
			var errMsgs []string
			for _, pathToCheck := range pathsOwnershipCheck {
				if err = cmdutil.ValidateRootOwnership(pathToCheck); err != nil {
					errMsgs = append(errMsgs, err.Error())
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
		}
		return
	}
}
