package k8s

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newLocalNodeStatusCommand(env cmdutil.ExecutionEnvironment) *cobra.Command {
	localNodeStatusCmd := &cobra.Command{
		Use:    "local-node-status",
		Short:  "Retrieve the current status of the local node",
		Hidden: true,
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {

			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("ERROR: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			status, err := client.NodeStatus(cmd.Context())
			if err != nil {
				cmd.PrintErrf("ERROR: Failed to get the status of the local node.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if err := cmdutil.FormatterFromContext(cmd.Context()).Print(status); err != nil {
				cmd.PrintErrf("ERROR: Failed to print the status of the local node.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}
		},
	}
	return localNodeStatusCmd
}
