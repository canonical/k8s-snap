package k8sd

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/spf13/cobra"
)

var rootCmdOpts struct {
	logDebug                            bool
	logVerbose                          bool
	stateDir                            string
	pprofAddress                        string
	disableNodeConfigController         bool
	disableControlPlaneConfigController bool
}

func NewRootCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8sd",
		Short: "Canonical Kubernetes orchestrator and clustering daemon",
		Run: func(cmd *cobra.Command, args []string) {
			app, err := app.New(app.Config{
				Debug:                               rootCmdOpts.logDebug,
				Verbose:                             rootCmdOpts.logVerbose,
				StateDir:                            rootCmdOpts.stateDir,
				Snap:                                env.Snap,
				PprofAddress:                        rootCmdOpts.pprofAddress,
				DisableNodeConfigController:         rootCmdOpts.disableNodeConfigController,
				DisableControlPlaneConfigController: rootCmdOpts.disableControlPlaneConfigController,
			})
			if err != nil {
				cmd.PrintErrf("Error: Failed to initialize k8sd: %v", err)
				env.Exit(1)
				return
			}

			if err := app.Run(cmd.Context(), nil); err != nil {
				cmd.PrintErrf("Error: Failed to run k8sd: %v", err)
				env.Exit(1)
				return
			}
		},
	}

	cmd.SetIn(env.Stdin)
	cmd.SetOut(env.Stdout)
	cmd.SetErr(env.Stderr)

	cmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "Show all debug messages")
	cmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "Show all information messages")
	cmd.PersistentFlags().StringVar(&rootCmdOpts.stateDir, "state-dir", "", "Directory with the dqlite datastore")
	cmd.PersistentFlags().StringVar(&rootCmdOpts.pprofAddress, "pprof-address", "", "Listen address for pprof endpoints, e.g. \"127.0.0.1:4217\"")
	cmd.PersistentFlags().BoolVar(&rootCmdOpts.disableNodeConfigController, "disable-node-config-controller", false, "Disable the Node Config Controller (defaults to `false`)")
	cmd.PersistentFlags().BoolVar(&rootCmdOpts.disableControlPlaneConfigController, "disable-control-plane-config-controller", false, "Disable the Control Plane Config Controller (defaults to `false`)")

	cmd.Flags().Uint("port", 0, "Default port for the HTTP API")
	cmd.Flags().MarkDeprecated("port", "this flag does not have any effect, and will be removed in a future version")

	cmd.AddCommand(newSqlCmd(env))

	return cmd
}
