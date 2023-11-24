package k8s

import (
	"context"

	cluster "github.com/canonical/k8s/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	bootstrapClusterCmd = &cobra.Command{
		Use:   "bootstrap-cluster",
		Short: "Create new cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.flagLogDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			err := cluster.Bootstrap(context.Background(), cluster.ClusterOpts{
				Address:  clusterCmdOpts.flagAddress,
				StateDir: clusterCmdOpts.flagStateDir,
				Verbose:  rootCmdOpts.flagLogVerbose,
				Debug:    rootCmdOpts.flagLogDebug,
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
