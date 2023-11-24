package k8s

import (
	"context"

	"github.com/canonical/k8s/pkg/k8s/cluster"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	bootstrapClusterCmd = &cobra.Command{
		Use:   "bootstrap-cluster",
		Short: "Create new cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.logDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			err := cluster.Bootstrap(context.Background(), cluster.ClusterOpts{
				Address:  clusterCmdOpts.address,
				StateDir: clusterCmdOpts.stateDir,
				Verbose:  rootCmdOpts.logVerbose,
				Debug:    rootCmdOpts.logDebug,
			})
			if err == nil {
				logrus.Info("Cluster created.")
			}

			return err
		},
	}
)

func init() {
	rootCmd.AddCommand(bootstrapClusterCmd)
}
