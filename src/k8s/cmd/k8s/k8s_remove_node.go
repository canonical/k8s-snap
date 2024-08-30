package k8s

import (
	"fmt"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
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
		force        bool
		outputFormat string
		timeout      time.Duration
	}
	cmd := &cobra.Command{
		Use:    "remove-node <node-name>",
		Short:  "Remove a node from the cluster",
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Args:   cmdutil.ExactArgs(env, 1),
		Run: func(cmd *cobra.Command, args []string) {
			if opts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", opts.timeout, minTimeout, minTimeout)
				opts.timeout = minTimeout
			}

			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			name := args[0]

			cmd.PrintErrf("Removing %q from the Kubernetes cluster. This may take a few seconds, please wait.\n", name)
			if err := client.RemoveNode(cmd.Context(), apiv1.RemoveNodeRequest{Name: name, Force: opts.force, Timeout: opts.timeout}); err != nil {
				cmd.PrintErrf("Error: Failed to remove node %q from the cluster.\n\nThe error was: %v\n", name, err)
				env.Exit(1)
				return
			}

			outputFormatter.Print(RemoveNodeResult{Name: name})
		},
	}

	cmd.Flags().BoolVar(&opts.force, "force", false, "forcibly remove the cluster member")
	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")

	return cmd
}
