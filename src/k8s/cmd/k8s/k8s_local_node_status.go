package k8s

import (
	"fmt"

	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newLocalNodeStatusCommand() *cobra.Command {
	localNodeStatusCmd := &cobra.Command{
		Use:     "local-node-status",
		Short:   "Retrieve the current status of the local node",
		Hidden:  true,
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			clusterStatus, err := k8sdClient.NodeStatus(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get cluster status: %w", err)
			}

			fmt.Println(clusterStatus)
			return nil
		},
	}
	return localNodeStatusCmd
}
