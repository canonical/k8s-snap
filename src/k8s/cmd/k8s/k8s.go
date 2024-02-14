package k8s

import (
	"fmt"

	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		logDebug   bool
		logVerbose bool
	}

	ew = errors.ErrorWrapper{}

	rootCmd = &cobra.Command{
		Use:   "k8s",
		Short: "Canonical Kubernetes CLI",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			withRoot, err := utils.RunsWithRootPrivilege(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to check if command runs as root: %w", err)
			}
			if !withRoot {
				return fmt.Errorf("You do not have enough permissions. Please run the command with sudo.")
			}
			return nil
		},
		PersistentPostRunE: ew.TransformToHumanError(),
		SilenceUsage:       true,
		SilenceErrors:      true,
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "Show all debug messages")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "Show all information messages")
}
