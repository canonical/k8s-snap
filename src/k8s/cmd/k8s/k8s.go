package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		logDebug   bool
		logVerbose bool
		stateDir   string
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

	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.stateDir, "state-dir", "", "Directory with the dqlite datastore")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "Show all debug messages")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "Show all information messages")

	// By default, the state dir is set to a fixed directory in the snap.
	// This shouldn't be overwritten by the user.
	rootCmd.PersistentFlags().MarkHidden("state-dir")

	// General
	rootCmd.AddCommand(newStatusCmd())

	// Clustering
	rootCmd.AddCommand(newBootstrapCmd())
	rootCmd.AddCommand(newAddNodeCmd())
	rootCmd.AddCommand(newJoinNodeCmd())
	rootCmd.AddCommand(newRemoveNodeCmd())

	// Components
	rootCmd.AddCommand(newEnableCmd())
	rootCmd.AddCommand(newDisableCmd())

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
