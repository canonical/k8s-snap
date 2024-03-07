package k8s

import (
	"fmt"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

type RemoveNodeResult struct {
	Name string `json:"name" yaml:"name"`
}

func (r RemoveNodeResult) String() string {
	return fmt.Sprintf("Removed %s from cluster.\n", r.Name)
}

func newRemoveNodeCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		force bool
	}
	cmd := &cobra.Command{
		Use:    "remove-node <node-name>",
		Short:  "Remove a node from the cluster",
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Args:   cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("ERROR: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			name := args[0]

			cmd.PrintErrf("Removing %q from the Kubernetes cluster. This may take a few seconds, please wait.\n", name)
			if err := client.RemoveNode(cmd.Context(), name, opts.force); err != nil {
				cmd.PrintErrf("ERROR: Failed to remove node %q from the cluster.\n\nThe error was: %v\n", name, err)
				env.Exit(1)
				return
			}

			if err := cmdutil.FormatterFromContext(cmd.Context()).Print(RemoveNodeResult{Name: name}); err != nil {
				cmd.PrintErrf("WARNING: Failed to print remove node result.\n\nThe error was: %v\n", err)
			}
		},
	}

	cmd.Flags().BoolVar(&opts.force, "force", false, "forcibly remove the cluster member")
	return cmd
}
