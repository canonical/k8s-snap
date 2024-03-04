package k8s

import (
	"fmt"

	v1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

var (
	localStatusCmdErrorMsgs = map[error]string{
		v1.ErrUnknown: "An error occurred while retrieving the node's status:\n",
	}
)

func newLocalNodeStatusCommand() *cobra.Command {
	localNodeStatusCmd := &cobra.Command{
		Use:     "local-node-status",
		Short:   "Retrieve the current status of the local node",
		Hidden:  true,
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, localStatusCmdErrorMsgs)

			clusterStatus, err := k8sdClient.NodeStatus(cmd.Context())
			if err != nil {
				return fmt.Errorf("Failed to get cluster status: %w", err)
			}

			fmt.Println(clusterStatus)
			return nil
		},
	}
	return localNodeStatusCmd
}
