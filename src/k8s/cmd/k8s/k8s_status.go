package k8s

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Retrieve the current status of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.flagLogDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			logrus.Info("Placeholder k8s status command")
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(statusCmd)
}
