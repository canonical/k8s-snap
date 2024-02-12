package k8sd

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		logDebug   bool
		logVerbose bool
		stateDir   string
		port       uint
	}

	rootCmd = &cobra.Command{
		Use:   "k8sd",
		Short: "Canonical Kubernetes orchestrator and clustering daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			snap := snap.NewSnap(os.Getenv("SNAP"), os.Getenv("SNAP_COMMON"))
			app, err := app.New(cmd.Context(), app.Config{
				Debug:      rootCmdOpts.logDebug,
				Verbose:    rootCmdOpts.logVerbose,
				StateDir:   rootCmdOpts.stateDir,
				ListenPort: rootCmdOpts.port,
				Snap:       snap,
			})
			if err != nil {
				return fmt.Errorf("failed to initialize k8sd: %w", err)
			}

			if err := app.Run(nil); err != nil {
				return fmt.Errorf("failed to run k8sd: %w", err)
			}
			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "Show all debug messages")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "Show all information messages")
	rootCmd.PersistentFlags().UintVar(&rootCmdOpts.port, "port", config.DefaultPort, "Port on which the REST API is exposed")
	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.stateDir, "state-dir", "", "Directory with the dqlite datastore")
}
