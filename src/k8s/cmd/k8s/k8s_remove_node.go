package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var (
	removeNodeCmdOpts struct {
		force   bool
		timeout time.Duration
	}

	removeNodeCmd = &cobra.Command{
		Use:   "remove-node <name>",
		Short: "Remove a node from the cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
				StateDir: clusterCmdOpts.stateDir,
				Verbose:  rootCmdOpts.logVerbose,
				Debug:    rootCmdOpts.logDebug,
			})
			if err != nil {
				return fmt.Errorf("failed to create cluster client: %w", err)
			}

			// TODO: Apply this check for all command where a timeout is required, do not repeat in each command.
			const minTimeout = 3 * time.Second
			if removeNodeCmdOpts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", removeNodeCmdOpts.timeout, minTimeout, minTimeout)
				removeNodeCmdOpts.timeout = minTimeout
			}

			timeoutCtx, cancel := context.WithTimeout(cmd.Context(), removeNodeCmdOpts.timeout)
			defer cancel()
			err = client.RemoveNode(timeoutCtx, name, removeNodeCmdOpts.force)
			if err != nil {
				return fmt.Errorf("failed to remove node from cluster: %w", err)
			}
			fmt.Printf("Removed %s from cluster.\n", name)
			return nil
		},
	}
)

func init() {
	removeNodeCmd.Flags().BoolVar(&removeNodeCmdOpts.force, "force", false, "Forcibly remove the cluster member")
	removeNodeCmd.PersistentFlags().DurationVar(&removeNodeCmdOpts.timeout, "timeout", 180*time.Second, "The max time to wait for the node to be removed.")

	rootCmd.AddCommand(removeNodeCmd)
}
