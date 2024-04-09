package k8s

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newStatusCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		waitReady bool
	}
	cmd := &cobra.Command{
		Use:    "status",
		Short:  "Retrieve the current status of the cluster",
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if !client.IsBootstrapped(cmd.Context()) {
				cmd.PrintErrln("Error: The node is not part of a Kubernetes cluster. You can bootstrap a new cluster with:\n\n  sudo k8s bootstrap")
				env.Exit(1)
				return
			}

			status, err := client.ClusterStatus(cmd.Context(), opts.waitReady)
			if err != nil {
				cmd.PrintErrf("Error: Failed to retrieve the cluster status.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if err := cmdutil.FormatterFromContext(cmd.Context()).Print(status); err != nil {
				cmd.PrintErrf("Error: Failed to print the cluster status.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}
		},
	}

	cmd.PersistentFlags().BoolVar(&opts.waitReady, "wait-ready", false, "wait until at least one cluster node is ready")
	return cmd
}
