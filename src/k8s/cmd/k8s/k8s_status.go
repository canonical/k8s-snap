package k8s

import (
	"context"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newStatusCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		waitReady    bool
		outputFormat string
		timeout      time.Duration
	}
	cmd := &cobra.Command{
		Use:    "status",
		Short:  "Retrieve the current status of the cluster",
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Run: func(cmd *cobra.Command, args []string) {
			if opts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", opts.timeout, minTimeout, minTimeout)
				opts.timeout = minTimeout
			}

			client, err := env.Snap.K8sdClient()
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			cobra.OnFinalize(cancel)

			if _, err := client.NodeStatus(cmd.Context()); err != nil {
				cmd.PrintErrln("Error: The node is not part of a Kubernetes cluster. You can bootstrap a new cluster with:\n\n  sudo k8s bootstrap")
				env.Exit(1)
				return
			}

			status, err := client.ClusterStatus(ctx, opts.waitReady)
			if err != nil {
				cmd.PrintErrf("Error: Failed to retrieve the cluster status.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			status.Config.MetricsServer = apiv1.MetricsServerConfig{}
			status.Config.CloudProvider = nil

			outputFormatter.Print(status)
		},
	}

	cmd.Flags().BoolVar(&opts.waitReady, "wait-ready", false, "wait until at least one cluster node is ready")
	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")
	return cmd
}
