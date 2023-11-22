package k8sd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		testFlag string
		debug    bool
	}

	rootCmd = &cobra.Command{
		Use:   "k8sd",
		Short: "Canonical Kubernetes orchestrator and clustering daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.debug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			logrus.WithField("flag", rootCmdOpts.testFlag).Info("Placeholder k8sd command")
			return nil
		},
	}
)

func init() {
	rootCmd.Flags().StringVar(&rootCmdOpts.testFlag, "flag", "value", "test flag (TODO: remove)")
	rootCmd.Flags().BoolVar(&rootCmdOpts.debug, "debug", false, "debug logs")
}
