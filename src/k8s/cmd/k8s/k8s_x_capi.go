package k8s

import (
	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/spf13/cobra"
)

func newXCAPICmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	setAuthTokenCmd := &cobra.Command{
		Use:   "set-auth-token <token>",
		Short: "Set the auth token for the CAPI provider",
		Args:  cmdutil.ExactArgs(env, 1),
		Run: func(cmd *cobra.Command, args []string) {
			token := args[0]
			if token == "" {
				cmd.PrintErrf("Error: The token must be provided.\n")
				env.Exit(1)
				return
			}

			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			err = client.SetClusterAPIAuthToken(cmd.Context(), apiv1.ClusterAPISetAuthTokenRequest{Token: token})
			if err != nil {
				cmd.PrintErrf("Error: Failed to set the CAPI auth token.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}
		},
	}
	setNodeToken := &cobra.Command{
		Use:   "set-node-token <token>",
		Short: "Set the node token to authenticate with per-node k8sd endpoints",
		Args:  cmdutil.ExactArgs(env, 1),
		Run: func(cmd *cobra.Command, args []string) {
			token := args[0]
			if token == "" {
				cmd.PrintErrf("Error: The token must be provided.\n")
				env.Exit(1)
				return
			}

			if err := utils.WriteFile(env.Snap.NodeTokenFile(), []byte(token), 0o600); err != nil {
				cmd.PrintErrf("Error: Failed to write the node token to file.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}
		},
	}

	cmd := &cobra.Command{
		Use:    "x-capi",
		Short:  "Manage the CAPI integration",
		Hidden: true,
	}

	cmd.AddCommand(setAuthTokenCmd)
	cmd.AddCommand(setNodeToken)

	return cmd
}
