package k8s

import (
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

type GetJoinTokenResult struct {
	JoinToken string `json:"join-token" yaml:"join-token"`
}

func (g GetJoinTokenResult) String() string {
	return fmt.Sprintf("On the node you want to join call:\n\n  sudo k8s join-cluster %s\n", g.JoinToken)
}

func newGetJoinTokenCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		worker bool
	}
	cmd := &cobra.Command{
		Use:     "get-join-token <node-name>",
		GroupID: "cluster",
		Short:   "Create a token for a node to join the cluster",
		PreRun:  chainPreRunHooks(hookRequireRoot(env)),
		Args:    cmdutil.ExactArgs(env, 1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			token, err := client.GetJoinToken(cmd.Context(), apiv1.GetJoinTokenRequest{Name: name, Worker: opts.worker})
			if err != nil {
				cmd.PrintErrf("Error: Could not generate a join token for %q.\n\nThe error was: %v\n", name, err)
				env.Exit(1)
				return
			}

			if err := cmdutil.FormatterFromContext(cmd.Context()).Print(GetJoinTokenResult{JoinToken: token}); err != nil {
				cmd.PrintErrf("Error: Failed to print the join token.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}
		},
	}

	cmd.PersistentFlags().BoolVar(&opts.worker, "worker", false, "generate a join token for a worker node")
	return cmd
}
