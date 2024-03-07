package k8sd

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/spf13/cobra"
)

var rootCmdOpts struct {
	logDebug   bool
	logVerbose bool
	stateDir   string
	port       uint
}

func NewRootCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8sd",
		Short: "Canonical Kubernetes orchestrator and clustering daemon",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetIn(env.Stdin)
			cmd.SetOut(env.Stdout)
			cmd.SetErr(env.Stderr)
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			app, err := app.New(cmd.Context(), app.Config{
				Debug:      rootCmdOpts.logDebug,
				Verbose:    rootCmdOpts.logVerbose,
				StateDir:   rootCmdOpts.stateDir,
				ListenPort: rootCmdOpts.port,
				Snap:       env.Snap,
			})
			if err != nil {
				cmd.PrintErrf("ERROR: Failed to initialize k8sd: %v", err)
				env.Exit(1)
				return
			}

			if err := app.Run(nil); err != nil {
				cmd.PrintErrf("ERROR: Failed to run k8sd: %v", err)
				env.Exit(1)
				return
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "Show all debug messages")
	cmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "Show all information messages")
	cmd.PersistentFlags().UintVar(&rootCmdOpts.port, "port", config.DefaultPort, "Port on which the REST API is exposed")
	cmd.PersistentFlags().StringVar(&rootCmdOpts.stateDir, "state-dir", "", "Directory with the dqlite datastore")

	cmd.AddCommand(newSqlCmd(env))

	return cmd
}
