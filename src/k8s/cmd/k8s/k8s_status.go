package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	statusCmd = &cobra.Command{
		Use:    "status",
		Short:  "Retrieve the current status of the cluster",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.logDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

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

			clusterStatus, err := c.ClusterStatus(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get cluster status: %w", err)
			}

			// TODO: make this nice and bright
			fmt.Printf("Number of nodes in the cluster: %d\n", len(clusterStatus.Members))
			fmt.Printf("HA cluster: %t\n", clusterStatus.HaClusterFormed())

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(statusCmd)
}
