package k8s

import (
	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newStatusCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		waitReady    bool
		outputFormat string
	}
	cmd := &cobra.Command{
		Use:    "status",
		Short:  "Retrieve the current status of the cluster",
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
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

			// hide MetricsServer config from user as it is enabled by default
			status.Config.MetricsServer = apiv1.MetricsServerConfig{}

			outputFormatter.Print(status)
		},
	}

	cmd.Flags().BoolVar(&opts.waitReady, "wait-ready", false, "wait until at least one cluster node is ready")
	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")
	return cmd
}
