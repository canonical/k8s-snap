package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

func chainPreRunHooks(hooks ...func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, hook := range hooks {
			err := hook(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func hookSetupClient(cmd *cobra.Command, args []string) error {
	var err error
	k8sdClient, err = client.NewClient(cmd.Context(), client.ClusterOpts{
		StateDir: rootCmdOpts.stateDir,
		Verbose:  rootCmdOpts.logVerbose,
		Debug:    rootCmdOpts.logDebug,
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	return nil
}
