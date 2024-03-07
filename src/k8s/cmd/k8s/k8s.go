package k8s

import (
	"context"
	"time"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

var (
	componentList = []string{"network", "dns", "gateway", "ingress", "local-storage", "load-balancer", "metrics-server"}
)

func NewRootCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var (
		opts struct {
			logDebug     bool
			logVerbose   bool
			outputFormat string
			stateDir     string
			timeout      time.Duration
		}
	)
	cmd := &cobra.Command{
		Use:   "k8s",
		Short: "Canonical Kubernetes CLI",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// set input/output streams
			cmd.SetIn(env.Stdin)
			cmd.SetOut(env.Stdout)
			cmd.SetErr(env.Stderr)

			// initialize context
			ctx := cmd.Context()

			// initialize formatter
			var err error
			formatter, err := cmdutil.NewFormatter(opts.outputFormat, cmd.OutOrStdout())
			if err != nil {
				cmd.PrintErrf("Error: Unknown --output-format %q. It must be one of %q (default), %q or %q.", opts.outputFormat, "plain", "json", "yaml")
				env.Exit(1)
				return
			}
			ctx = cmdutil.ContextWithFormatter(ctx, formatter)

			// configure command context timeout
			const minTimeout = 3 * time.Second
			if opts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", opts.timeout, minTimeout, minTimeout)
				opts.timeout = minTimeout
			}

			ctx, cancel := context.WithTimeout(ctx, opts.timeout)
			cobra.OnFinalize(cancel)

			cmd.SetContext(ctx)
		},
	}

	cmd.PersistentFlags().StringVar(&opts.stateDir, "state-dir", "", "directory with the dqlite datastore")
	cmd.PersistentFlags().BoolVarP(&opts.logDebug, "debug", "d", false, "show all debug messages")
	cmd.PersistentFlags().BoolVarP(&opts.logVerbose, "verbose", "v", true, "show all information messages")
	cmd.PersistentFlags().StringVarP(&opts.outputFormat, "output-format", "o", "plain", "set the output format to one of plain, json or yaml")
	cmd.PersistentFlags().DurationVarP(&opts.timeout, "timeout", "t", 90*time.Second, "the max time to wait for the command to execute")

	// By default, the state dir is set to a fixed directory in the snap.
	// This shouldn't be overwritten by the user.
	cmd.PersistentFlags().MarkHidden("state-dir")
	cmd.PersistentFlags().MarkHidden("debug")
	cmd.PersistentFlags().MarkHidden("verbose")

	// General
	cmd.AddCommand(newStatusCmd(env))

	// Clustering
	cmd.AddCommand(newBootstrapCmd(env))
	cmd.AddCommand(newGetJoinTokenCmd(env))
	cmd.AddCommand(newJoinClusterCmd(env))
	cmd.AddCommand(newRemoveNodeCmd(env))

	// Components
	cmd.AddCommand(newEnableCmd(env))
	cmd.AddCommand(newDisableCmd(env))
	cmd.AddCommand(newSetCmd(env))
	cmd.AddCommand(newGetCmd(env))

	// internal
	cmd.AddCommand(newGenerateAuthTokenCmd(env))
	cmd.AddCommand(newKubeConfigCmd(env))
	cmd.AddCommand(newLocalNodeStatusCommand(env))
	cmd.AddCommand(newRevokeAuthTokenCmd(env))
	cmd.AddCommand(newGenerateDocsCmd(env))
	cmd.AddCommand(xPrintShimPidsCmd)

	// Those commands replace the executable - no need for error wrapping.
	cmd.AddCommand(newHelmCmd(env))
	cmd.AddCommand(newKubectlCmd(env))

	cmd.DisableAutoGenTag = true
	return cmd
}
