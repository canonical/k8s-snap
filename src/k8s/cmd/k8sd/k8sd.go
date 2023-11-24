package k8sd

import (
	"context"
	"fmt"

	"github.com/canonical/microcluster/microcluster"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		version    bool
		logDebug   bool
		logVerbose bool
		stateDir   string
	}

	rootCmd = &cobra.Command{
		Use:   "k8sd",
		Short: "Canonical Kubernetes orchestrator and clustering daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := microcluster.App(
				context.Background(),
				microcluster.Args{
					StateDir: rootCmdOpts.stateDir,
					Verbose:  rootCmdOpts.logVerbose,
					Debug:    rootCmdOpts.logDebug,
				},
			)
			if err != nil {
				return fmt.Errorf("failed to initialize microcluster app: %w", err)
			}

			err = m.Start(nil, nil, nil)
			if err != nil {
				return fmt.Errorf("failed to start microcluster app: %w", err)
			}

			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "Show all debug messages")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "Show all information messages")

	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.stateDir, "state-dir", "", "Path to store state information")
}
