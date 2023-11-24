package k8s

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		flagLogDebug   bool
		flagLogVerbose bool
	}

	rootCmd = &cobra.Command{
		Use:   "k8s",
		Short: "Canonical Kubernetes CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.flagLogDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.flagLogDebug, "debug", "d", false, "Show all debug messages")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.flagLogVerbose, "verbose", "v", true, "Show all information messages")
	rootCmd.SilenceUsage = true
}
