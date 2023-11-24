package k8s

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		logDebug   bool
		logVerbose bool
	}

	rootCmd = &cobra.Command{
		Use:          "k8s",
		Short:        "Canonical Kubernetes CLI",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.logDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "Show all debug messages")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "Show all information messages")
}
