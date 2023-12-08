package k8sd

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		logDebug   bool
		logVerbose bool
		storageDir string
		port       uint
	}

	rootCmd = &cobra.Command{
		Use:   "k8sd",
		Short: "Canonical Kubernetes orchestrator and clustering daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := app.New(cmd.Context(), app.Config{
				Debug:      rootCmdOpts.logDebug,
				Verbose:    rootCmdOpts.logVerbose,
				StateDir:   rootCmdOpts.storageDir,
				ListenPort: rootCmdOpts.port,
			})
			if err != nil {
				return fmt.Errorf("failed to initialize k8sd: %w", err)
			}

			if err := app.Run(); err != nil {
				return fmt.Errorf("failed to run k8sd: %w", err)
			}
			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "Show all debug messages")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "Show all information messages")
	rootCmd.PersistentFlags().UintVar(&rootCmdOpts.port, "port", client.DefaultPort, "Port on which the REST API is exposed")
	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.storageDir, "storage-dir", "", "Directory with the dqlite datastore")
}
