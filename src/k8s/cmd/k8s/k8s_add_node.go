package k8s

import (
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

var (
	addNodeCmdOpts struct {
		worker bool
	}
	addNodeCmdErrorMsgs = map[error]string{
		apiv1.ErrTokenAlreadyCreated: "A token for this node was already created and the node did not join.",
	}
)

func newAddNodeCmd() *cobra.Command {
	addNodeCmd := &cobra.Command{
		Use:     "add-node <name>",
		Short:   "Create a connection token for a node to join the cluster",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) > 1 {
				return fmt.Errorf("too many arguments: provide only the node name to add")
			}
			if len(args) < 1 {
				return fmt.Errorf("missing argument: provide the node name to add")
			}

			defer errors.Transform(&err, addNodeCmdErrorMsgs)
			name := args[0]

			// Create a token that will be used by the joining node to join the cluster.
			token, err := k8sdClient.CreateJoinToken(cmd.Context(), name, addNodeCmdOpts.worker)
			if err != nil {
				return fmt.Errorf("failed to retrieve token: %w", err)
			}

			// TODO: Print guidance on what to do with the token.
			//       This requires a --format flag first as we still need some machine readable output for the integration tests.
			fmt.Println(token)
			return nil
		},
	}

	addNodeCmd.Flags().BoolVar(&addNodeCmdOpts.worker, "worker", false, "generate a token for a worker node")
	return addNodeCmd
}
