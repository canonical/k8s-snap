package k8sd

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/canonical/k8s/pkg/log"
	"github.com/spf13/cobra"
)

var rootCmdOpts struct {
	logDebug                            bool
	logVerbose                          bool
	logLevel                            int
	stateDir                            string
	pprofAddress                        string
	disableNodeConfigController         bool
	disableNodeLabelController          bool
	disableControlPlaneConfigController bool
	disableFeatureController            bool
	disableUpdateNodeConfigController   bool
	disableCSRSigningController         bool
	featureControllerMaxRetryAttempts   int
}

func addCommands(root *cobra.Command, group *cobra.Group, commands ...*cobra.Command) {
	if group != nil {
		root.AddGroup(group)
		for _, command := range commands {
			command.GroupID = group.ID
		}
	}

	root.AddCommand(commands...)
}

func NewRootCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8sd",
		Short: "Canonical Kubernetes orchestrator and clustering daemon",
		Run: func(cmd *cobra.Command, args []string) {
			// configure logging
			log.Configure(log.Options{
				LogLevel:     rootCmdOpts.logLevel,
				AddDirHeader: true,
			})

			app, err := app.New(app.Config{
				Debug:                               rootCmdOpts.logDebug,
				Verbose:                             rootCmdOpts.logVerbose,
				StateDir:                            rootCmdOpts.stateDir,
				Snap:                                env.Snap,
				PprofAddress:                        rootCmdOpts.pprofAddress,
				DisableNodeConfigController:         rootCmdOpts.disableNodeConfigController,
				DisableNodeLabelController:          rootCmdOpts.disableNodeLabelController,
				DisableControlPlaneConfigController: rootCmdOpts.disableControlPlaneConfigController,
				DisableUpdateNodeConfigController:   rootCmdOpts.disableUpdateNodeConfigController,
				DisableFeatureController:            rootCmdOpts.disableFeatureController,
				DisableCSRSigningController:         rootCmdOpts.disableCSRSigningController,
				FeatureControllerMaxRetryAttempts:   rootCmdOpts.featureControllerMaxRetryAttempts,
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

	cmd.PersistentFlags().IntVarP(&rootCmdOpts.logLevel, "log-level", "l", 0, "k8sd log level")
	cmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "Show all debug messages")
	cmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "Show all information messages")
	cmd.PersistentFlags().StringVar(&rootCmdOpts.stateDir, "state-dir", "", "Directory with the dqlite datastore")
	cmd.PersistentFlags().StringVar(&rootCmdOpts.pprofAddress, "pprof-address", "", "Listen address for pprof endpoints, e.g. \"127.0.0.1:4217\"")
	cmd.PersistentFlags().BoolVar(&rootCmdOpts.disableNodeConfigController, "disable-node-config-controller", false, "Disable the Node Config Controller")
	cmd.PersistentFlags().BoolVar(&rootCmdOpts.disableNodeLabelController, "disable-node-label-controller", false, "Disable the Node Label Controller")
	cmd.PersistentFlags().BoolVar(&rootCmdOpts.disableControlPlaneConfigController, "disable-control-plane-config-controller", false, "Disable the Control Plane Config Controller")
	cmd.PersistentFlags().BoolVar(&rootCmdOpts.disableUpdateNodeConfigController, "disable-update-node-config-controller", false, "Disable the Update Node Config Controller")
	cmd.PersistentFlags().BoolVar(&rootCmdOpts.disableFeatureController, "disable-feature-controller", false, "Disable the Feature Controller")
	cmd.PersistentFlags().BoolVar(&rootCmdOpts.disableCSRSigningController, "disable-csrsigning-controller", false, "Disable the CSR signing controller")

	cmd.Flags().Uint("port", 0, "Default port for the HTTP API")
	cmd.Flags().MarkDeprecated("port", "this flag does not have any effect, and will be removed in a future version")
	cmd.Flags().IntVar(&rootCmdOpts.featureControllerMaxRetryAttempts, "feature-controller-max-retry-attempts", 15, "Maximum number of retry attempts for the feature controller before giving up. Zero or negative values mean no limit.")

	cmd.AddCommand(newSqlCmd(env))

	addCommands(
		cmd,
		&cobra.Group{ID: "cluster", Title: "K8sd clustering commands:"},
		newClusterRecoverCmd(env),
	)

	return cmd
}
