package k8s

import (
	"fmt"

	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/cmd/k8s/formatter"
	"github.com/spf13/cobra"
)

var (
	removeNodeCmdOpts struct {
		force bool
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

			fmt.Fprintf(cmd.ErrOrStderr(), "Removing %q from the cluster. This may take some time, please wait.", name)
			if err := k8sdClient.RemoveNode(cmd.Context(), name, removeNodeCmdOpts.force); err != nil {
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
	removeNodeCmd.FlagErrorFunc()
	return removeNodeCmd
}
