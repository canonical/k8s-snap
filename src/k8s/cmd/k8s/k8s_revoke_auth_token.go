package k8s

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newRevokeAuthTokenCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		token string
	}
	cmd := &cobra.Command{
		Use:    "revoke-auth-token --token <token>",
		Hidden: true,
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {

			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if err := client.RevokeAuthToken(cmd.Context(), opts.token); err != nil {
				cmd.PrintErrf("Error: Failed to revoke the auth token.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}
		},
	}

	cmd.Flags().StringVar(&opts.token, "token", "", "Token")
	return cmd
}
