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
		storageDir string
		port       string
	}

	rootCmd = &cobra.Command{
		Use:   "k8sd",
		Short: "Canonical Kubernetes orchestrator and clustering daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := microcluster.App(
				context.Background(),
				microcluster.Args{
					ListenPort: rootCmdOpts.port,
					StateDir:   rootCmdOpts.storageDir,
					Verbose:    rootCmdOpts.logVerbose,
					Debug:      rootCmdOpts.logDebug,
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

	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.port, "port", "6444", "Port on which the REST-API is exposed")

	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.storageDir, "storage-dir", "", "directory with the dqlite datastore")
}
