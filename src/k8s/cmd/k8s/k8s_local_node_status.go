package k8s

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newLocalNodeStatusCommand(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		outputFormat string
	}
	cmd := &cobra.Command{
		Use:    "local-node-status",
		Short:  "Retrieve the current status of the local node",
		Hidden: true,
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			response, initialized, err := client.NodeStatus(cmd.Context())
			if err != nil {
				cmd.PrintErrf("Error: Failed to check the current node status.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			} else if !initialized {
				cmd.PrintErrln("Error: The node is not part of a Kubernetes cluster. You can bootstrap a new cluster with:\n\n  sudo k8s bootstrap")
				env.Exit(1)
				return
			}

			outputFormatter.Print(response.NodeStatus)
		},
	}
	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")

	return cmd
}
