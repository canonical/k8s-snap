package k8s

import (
	"context"
	"fmt"
	"time"

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
		timeout      time.Duration
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

			const minTimeout = 3 * time.Second
			if rootCmdOpts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", rootCmdOpts.timeout, minTimeout, minTimeout)
				rootCmdOpts.timeout = minTimeout
			}

			timeoutCtx, cancel := context.WithTimeout(context.Background(), rootCmdOpts.timeout)
			cobra.OnFinalize(func() {
				// Use OnFinalize because PostRun is not executed on error.
				cancel()
			})

			cmd.SetContext(timeoutCtx)
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.stateDir, "state-dir", "", "directory with the dqlite datastore")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "show all debug messages")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "show all information messages")
	rootCmd.PersistentFlags().StringVarP(&rootCmdOpts.outputFormat, "output-format", "o", "plain", "set the output format to one of plain, json or yaml")
	rootCmd.PersistentFlags().DurationVarP(&rootCmdOpts.timeout, "timeout", "t", 90*time.Second, "the max time to wait for the command to execute")

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

	rootCmd.DisableAutoGenTag = true
	return rootCmd
}
