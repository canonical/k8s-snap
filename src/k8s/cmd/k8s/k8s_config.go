package k8s

import (
	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newKubeConfigCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		server string
	}
	cmd := &cobra.Command{
		Use:    "config",
		Short:  "Generate a kubeconfig that can be used to access the Kubernetes cluster",
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

			config, err := client.KubeConfig(cmd.Context(), apiv1.GetKubeConfigRequest{Server: opts.server})
			if err != nil {
				cmd.PrintErrf("Error: Failed to generate an admin kubeconfig for %q.\n\nThe error was: %v\n", opts.server, err)
				env.Exit(1)
				return
			}

			cmd.Println(config)
		},
	}
	cmd.PersistentFlags().StringVar(&opts.server, "server", "", "custom cluster server address")
	return cmd
}
