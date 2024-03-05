package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		logDebug     bool
		logVerbose   bool
		outputFormat string
		stateDir     string
	}
	k8sdClient client.Client
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "k8s",
		Short: "Canonical Kubernetes CLI",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			withRoot, err := utils.RunsWithRootPrivilege(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to check if command runs as root: %w", err)
			}
			if !withRoot {
				return fmt.Errorf("insufficient permissions: run the command with sudo")
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.stateDir, "state-dir", "", "directory with the dqlite datastore")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "show all debug messages")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "show all information messages")
	rootCmd.PersistentFlags().StringVarP(&rootCmdOpts.outputFormat, "output-format", "o", "plain", "set the output format to one of plain, json or yaml")

	// By default, the state dir is set to a fixed directory in the snap.
	// This shouldn't be overwritten by the user.
	rootCmd.PersistentFlags().MarkHidden("state-dir")
	rootCmd.PersistentFlags().MarkHidden("debug")
	rootCmd.PersistentFlags().MarkHidden("verbose")

	// General
	rootCmd.AddCommand(newStatusCmd())

	// Clustering
	rootCmd.AddCommand(newBootstrapCmd())
	rootCmd.AddCommand(newGetJoinTokenCmd())
	rootCmd.AddCommand(newJoinClusterCmd())
	rootCmd.AddCommand(newRemoveNodeCmd())

	// Components
	rootCmd.AddCommand(newEnableCmd())
	rootCmd.AddCommand(newDisableCmd())
	rootCmd.AddCommand(newSetCmd())
	rootCmd.AddCommand(newGetCmd())

	// internal
	rootCmd.AddCommand(newGenerateAuthTokenCmd())
	rootCmd.AddCommand(newKubeConfigCmd())
	rootCmd.AddCommand(newLocalNodeStatusCommand())
	rootCmd.AddCommand(newRevokeAuthTokenCmd())
	rootCmd.AddCommand(xPrintShimPidsCmd)

	// Those commands replace the executable - no need for error wrapping.
	rootCmd.AddCommand(newHelmCmd())
	rootCmd.AddCommand(newKubectlCmd())

	return rootCmd
}
