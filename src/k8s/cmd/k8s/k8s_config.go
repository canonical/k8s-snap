package k8s

import (
	"context"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newKubeConfigCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		server  string
		timeout time.Duration
	}
	cmd := &cobra.Command{
		Use:    "config",
		Hidden: true,
		Short:  "Generate an admin kubeconfig that can be used to access the Kubernetes cluster",
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			if opts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", opts.timeout, minTimeout, minTimeout)
				opts.timeout = minTimeout
			}

			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			_ = GetNodeStatus(client, cmd, env)

			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			cobra.OnFinalize(cancel)

			config, err := client.KubeConfig(ctx, apiv1.GetKubeConfigRequest{Server: opts.server})
			if err != nil {
				cmd.PrintErrf("Error: Failed to generate an admin kubeconfig for %q.\n\nThe error was: %v\n", opts.server, err)
				env.Exit(1)
				return
			}

			cmd.Println(config)
		},
	}
	cmd.Flags().StringVar(&opts.server, "server", "", "custom cluster server address")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")
	return cmd
}
