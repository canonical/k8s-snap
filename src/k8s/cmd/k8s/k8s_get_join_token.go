package k8s

import (
	"context"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newGetJoinTokenCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		worker  bool
		timeout time.Duration
		ttl     time.Duration
	}
	cmd := &cobra.Command{
		Use:    "get-join-token <node-name>",
		Short:  "Create a token for a node to join the cluster",
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Args:   cmdutil.MaximumNArgs(env, 1),
		Run: func(cmd *cobra.Command, args []string) {
			if !opts.worker && len(args) == 0 {
				cmd.PrintErrln("Error: A node name is required for control-plane nodes.")
				env.Exit(1)
				return
			}

			var name string
			if len(args) == 1 {
				name = args[0]
			}

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

			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			cobra.OnFinalize(cancel)
			token, err := client.GetJoinToken(ctx, apiv1.GetJoinTokenRequest{Name: name, Worker: opts.worker, TTL: opts.ttl})
			if err != nil {
				cmd.PrintErrf("Error: Could not generate a join token for %q.\n\nThe error was: %v\n", name, err)
				env.Exit(1)
				return
			}

			cmd.Println(token.EncodedToken)
		},
	}

	cmd.Flags().BoolVar(&opts.worker, "worker", false, "generate a join token for a worker node")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")
	// The CLI uses verbose names for flags instead of abbreviations. Internally and for the API, the common TTL (time-to-live) name is used.
	cmd.Flags().DurationVar(&opts.ttl, "expires-in", 24*time.Hour, "the time until the token expires")
	return cmd
}
