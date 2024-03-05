package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/cmd/k8s/formatter"
	"github.com/spf13/cobra"
)

var (
	removeNodeCmdOpts struct {
		force   bool
		timeout time.Duration
	}
)

type RemoveNodeResult struct {
	Name string `json:"name" yaml:"name"`
}

func (r RemoveNodeResult) String() string {
	return fmt.Sprintf("Removed %s from cluster.\n", r.Name)
}

func newRemoveNodeCmd() *cobra.Command {
	removeNodeCmd := &cobra.Command{
		Use:     "remove-node <node-name>",
		Short:   "Remove a node from the cluster",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) > 1 {
				return fmt.Errorf("too many arguments: provide only the name of the node to remove")
			}
			if len(args) < 1 {
				return fmt.Errorf("missing argument: provide the name of the node to remove")
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

			fmt.Fprintf(cmd.ErrOrStderr(), "Removing %q from the cluster. This may take some time, please wait.", name)
			if err := k8sdClient.RemoveNode(timeoutCtx, name, removeNodeCmdOpts.force); err != nil {
				return fmt.Errorf("failed to remove node from cluster: %w", err)
			}
			f, err := formatter.New(rootCmdOpts.outputFormat, cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}
			return f.Print(RemoveNodeResult{
				Name: name,
			})
		},
	}
	removeNodeCmd.Flags().BoolVar(&removeNodeCmdOpts.force, "force", false, "forcibly remove the cluster member")
	removeNodeCmd.PersistentFlags().DurationVar(&removeNodeCmdOpts.timeout, "timeout", 180*time.Second, "the max time to wait for the node to be removed")
	removeNodeCmd.FlagErrorFunc()
	return removeNodeCmd
}
