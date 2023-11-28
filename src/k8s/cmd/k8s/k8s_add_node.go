package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8s/cluster"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	addNodeCmd = &cobra.Command{
		Use:   "add-node <name>",
		Short: "Create a connection token for a node to join the cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.logDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			name := args[0]

			client, err := cluster.NewClient(cmd.Context(), cluster.ClusterOpts{
				RemoteAddress: clusterCmdOpts.remoteAddress,
				StorageDir:    clusterCmdOpts.storageDir,
				Verbose:       rootCmdOpts.logVerbose,
				Debug:         rootCmdOpts.logDebug,
			})
			if err != nil {
				return fmt.Errorf("failed to create cluster client: %w", err)
			}

			token, err := client.GetToken(cmd.Context(), name)
			if err != nil {
				return fmt.Errorf("failed to retrieve token: %w", err)
			}

			fmt.Println(token)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(addNodeCmd)
}
