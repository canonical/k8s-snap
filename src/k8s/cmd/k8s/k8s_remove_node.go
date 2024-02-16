package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

var (
	removeNodeCmdOpts struct {
		force   bool
		timeout time.Duration
	}
)

func newRemoveNodeCmd() *cobra.Command {
	removeNodeCmd := &cobra.Command{
		Use:               "remove-node <name>",
		Short:             "Remove a node from the cluster",
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) > 1 {
				return fmt.Errorf("Too many arguments. Please, only provide the name of the node to remove.")
			}
			if len(args) < 1 {
				return fmt.Errorf("Not enough arguments. Please, provide the name of the node to remove.")
			}

			defer errors.Transform(&err, nil)

			name := args[0]

			// TODO: Apply this check for all command where a timeout is required, do not repeat in each command.
			const minTimeout = 3 * time.Second
			if removeNodeCmdOpts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", removeNodeCmdOpts.timeout, minTimeout, minTimeout)
				removeNodeCmdOpts.timeout = minTimeout
			}

			timeoutCtx, cancel := context.WithTimeout(cmd.Context(), removeNodeCmdOpts.timeout)
			defer cancel()
			if err := k8sdClient.RemoveNode(timeoutCtx, name, removeNodeCmdOpts.force); err != nil {
				return fmt.Errorf("failed to remove node from cluster: %w", err)
			}
			fmt.Printf("Removed %s from cluster.\n", name)
			return nil
		},
	}
	removeNodeCmd.Flags().BoolVar(&removeNodeCmdOpts.force, "force", false, "Forcibly remove the cluster member")
	removeNodeCmd.PersistentFlags().DurationVar(&removeNodeCmdOpts.timeout, "timeout", 180*time.Second, "The max time to wait for the node to be removed.")
	removeNodeCmd.FlagErrorFunc()
	return removeNodeCmd
}
