package k8s

import (
	cmdutil "github.com/canonical/k8s/cmd/util"

	"github.com/spf13/cobra"
)

func newGenerateAuthTokenCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		username string
		groups   []string
	}

	cmd := &cobra.Command{
		Use:    "generate-auth-token --username <user> [--groups <group1>,<group2>]",
		Hidden: true,
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}
			token, err := client.GenerateAuthToken(cmd.Context(), opts.username, opts.groups)
			if err != nil {
				cmd.PrintErrf("Error: Failed to generate the requested Kubernetes auth token.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}
			cmd.Println(token)
		},
	}
	cmd.Flags().StringVar(&opts.username, "username", "", "Username")
	cmd.Flags().StringSliceVar(&opts.groups, "groups", nil, "Groups")
	return cmd
}
