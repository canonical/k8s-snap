package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	addNodeCmd = &cobra.Command{
		Use:    "add-node <name>",
		Short:  "Create a connection token for a node to join the cluster",
		Args:   cobra.ExactArgs(1),
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.logDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			name := args[0]

			c, err := client.NewClient(cmd.Context(), client.ClusterOpts{
				StorageDir:    clusterCmdOpts.storageDir,
				RemoteAddress: clusterCmdOpts.remoteAddress,
				Port:          clusterCmdOpts.port,
				Verbose:       rootCmdOpts.logVerbose,
				Debug:         rootCmdOpts.logDebug,
			})
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			// Create a token that will be used by the joining
			// node to join the cluster.
			token, err := c.CreateJoinToken(cmd.Context(), name)
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
