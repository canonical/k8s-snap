package k8s

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	statusCmdOpts struct {
		debug bool
	}

	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Retrieve the current status of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if statusCmdOpts.debug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			logrus.Info("Placeholder k8s status command")
			return nil
		},
	}
)

func init() {
	statusCmd.Flags().BoolVar(&statusCmdOpts.debug, "debug", false, "debug logs")
	rootCmd.AddCommand(statusCmd)
}
