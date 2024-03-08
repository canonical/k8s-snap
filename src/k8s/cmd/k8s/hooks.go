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
