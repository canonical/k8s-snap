package k8s

import (
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/cmd/k8s/formatter"
	"github.com/spf13/cobra"
)

var (
	getJoinTokenCmdOpts struct {
		worker       bool
		outputFormat string
	}
	getJoinTokenCmdErrorMsgs = map[error]string{
		apiv1.ErrTokenAlreadyCreated: "A token for this node was already created and the node did not join.",
	}
)

type GetJoinTokenResult struct {
	JoinToken string `json:"join-token" yaml:"join-token"`
}

func (g GetJoinTokenResult) String() string {
	return fmt.Sprintf("On the node you want to join call:\n\n  sudo k8s join-cluster %s\n\n", g.JoinToken)
}

func newGetJoinTokenCmd() *cobra.Command {
	getJoinTokenCmd := &cobra.Command{
		Use:     "get-join-token <node-name>",
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

			result := GetJoinTokenResult{
				JoinToken: joinToken,
			}
			f, err := formatter.New(getJoinTokenCmdOpts.outputFormat, cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}
			return f.Print(result)
		},
	}

	getJoinTokenCmd.PersistentFlags().StringVarP(&getJoinTokenCmdOpts.outputFormat, "output-format", "o", "plain", "Specify in which format the output should be printed. One of plain, json or yaml")
	getJoinTokenCmd.PersistentFlags().BoolVar(&getJoinTokenCmdOpts.worker, "worker", false, "generate a join token for a worker node")
	return getJoinTokenCmd
}
