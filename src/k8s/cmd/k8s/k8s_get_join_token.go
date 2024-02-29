package k8s

import (
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

var (
	getJoinTokenCmdOpts struct {
		worker bool
	}
	getJoinTokenCmdErrorMsgs = map[error]string{
		apiv1.ErrTokenAlreadyCreated: "A token for this node was already created and the node did not join.",
	}
)

func newGetJoinTokenCmd() *cobra.Command {
	getJoinTokenCmd := &cobra.Command{
		Use:     "get-join-token <name>",
		Short:   "Create a join token for a node to join the cluster",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) > 1 {
				return fmt.Errorf("too many arguments: only provide the node name for 'get-join-token'")
			}
			if len(args) < 1 {
				return fmt.Errorf("missing argument: please provide the node name for 'get-join-token'")
			}

			defer errors.Transform(&err, getJoinTokenCmdErrorMsgs)
			name := args[0]

			// Create a joinToken that will be used by the joining node to join the cluster.
			joinToken, err := k8sdClient.CreateJoinToken(cmd.Context(), name, getJoinTokenCmdOpts.worker)
			if err != nil {
				return fmt.Errorf("failed to retrieve join token: %w", err)
			}

			// TODO: Print guidance on what to do with the token.
			//       This requires a --format flag first as we still need some machine readable output for the integration tests.
			fmt.Println(joinToken)
			return nil
		},
	}

	getJoinTokenCmd.Flags().BoolVar(&getJoinTokenCmdOpts.worker, "worker", false, "generate a join token for a worker node")
	return getJoinTokenCmd
}
