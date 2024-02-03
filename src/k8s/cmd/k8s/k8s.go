package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		logDebug   bool
		logVerbose bool
	}

	rootCmd = &cobra.Command{
		Use:   "k8s",
		Short: "Canonical Kubernetes CLI",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			withRoot, err := utils.RunsWithRootPrivilege(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to check if command runs as root: %w", err)
			}
			if !withRoot {
				return fmt.Errorf("k8s CLI needs to run with root priviledge.")
			}
			return nil
		},
		SilenceUsage: true,
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "Show all debug messages")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "Show all information messages")
}
