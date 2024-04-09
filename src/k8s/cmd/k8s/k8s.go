package k8s

import (
	"context"
	"time"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

var (
	componentList = []string{"network", "dns", "gateway", "ingress", "local-storage", "load-balancer", "metrics-server"}

	globalFormatter cmdutil.Formatter
)

func addCommands(root *cobra.Command, group *cobra.Group, commands ...*cobra.Command) {
	if group != nil {
		root.AddGroup(group)
		for _, command := range commands {
			command.GroupID = group.ID
		}
	}

	root.AddCommand(commands...)
}

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
			// initialize context
			ctx := cmd.Context()

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

	// set input/output streams
	cmd.SetIn(env.Stdin)
	cmd.SetOut(env.Stdout)
	cmd.SetErr(env.Stderr)

	cmd.PersistentFlags().StringVar(&opts.stateDir, "state-dir", "", "directory with the dqlite datastore")
	cmd.PersistentFlags().BoolVarP(&opts.logDebug, "debug", "d", false, "show all debug messages")
	cmd.PersistentFlags().BoolVarP(&opts.logVerbose, "verbose", "v", true, "show all information messages")
	cmd.PersistentFlags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")

	// By default, the state dir is set to a fixed directory in the snap.
	// This shouldn't be overwritten by the user.
	cmd.PersistentFlags().MarkHidden("state-dir")
	cmd.PersistentFlags().MarkHidden("debug")
	cmd.PersistentFlags().MarkHidden("verbose")
	cmd.PersistentFlags().MarkHidden("timeout")

	// General
	addCommands(
		cmd,
		&cobra.Group{ID: "general", Title: "General Commands:"},
		newStatusCmd(env),
		newKubeConfigCmd(env),
		newKubectlCmd(env),
	)

	// Clustering
	addCommands(
		cmd,
		&cobra.Group{ID: "cluster", Title: "Clustering Commands:"},
		newBootstrapCmd(env),
		newGetJoinTokenCmd(env),
		newJoinClusterCmd(env),
		newRemoveNodeCmd(env),
	)

	// Management
	addCommands(
		cmd,
		&cobra.Group{ID: "management", Title: "Management Commands:"},
		newEnableCmd(env),
		newDisableCmd(env),
		newSetCmd(env),
		newGetCmd(env),
	)

	// hidden commands
	addCommands(
		cmd,
		nil,
		newGenerateAuthTokenCmd(env),
		newLocalNodeStatusCommand(env),
		newRevokeAuthTokenCmd(env),
		newGenerateDocsCmd(env),
		xPrintShimPidsCmd,
		newHelmCmd(env),
	)

	cmd.DisableAutoGenTag = true
	return cmd
}
