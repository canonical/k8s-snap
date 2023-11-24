package k8sd

import (
	"context"

	"github.com/canonical/microcluster/microcluster"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		flagVersion    bool
		flagLogDebug   bool
		flagLogVerbose bool
		flagStateDir   string
	}

	rootCmd = &cobra.Command{
		Use:   "k8sd",
		Short: "Canonical Kubernetes orchestrator and clustering daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := microcluster.App(
				context.Background(),
				microcluster.Args{
					StateDir: rootCmdOpts.flagStateDir,
					Verbose:  rootCmdOpts.flagLogVerbose,
					Debug:    rootCmdOpts.flagLogDebug,
				},
			)
			if err != nil {
				return err
			}

			return m.Start(nil, nil, nil)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.flagLogDebug, "debug", "d", false, "Show all debug messages")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.flagLogVerbose, "verbose", "v", true, "Show all information messages")

	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.flagStateDir, "state-dir", "", "Path to store state information")
}
