package k8s

import (
	"time"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/spf13/cobra"
)

var (
	featureList = features.Public()

	outputFormatter cmdutil.Formatter
)

const minTimeout = 3 * time.Second

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
			logDebug   bool
			logVerbose bool
			stateDir   string
		}
	)
	cmd := &cobra.Command{
		Use:   "k8s",
		Short: "Canonical Kubernetes CLI",
	}

	// set input/output streams
	cmd.SetIn(env.Stdin)
	cmd.SetOut(env.Stdout)
	cmd.SetErr(env.Stderr)

	cmd.PersistentFlags().StringVar(&opts.stateDir, "state-dir", "", "directory with the dqlite datastore")
	cmd.PersistentFlags().BoolVarP(&opts.logDebug, "debug", "d", false, "show all debug messages")
	cmd.PersistentFlags().BoolVarP(&opts.logVerbose, "verbose", "v", true, "show all information messages")

	// By default, the state dir is set to a fixed directory in the snap.
	// This shouldn't be overwritten by the user.
	cmd.PersistentFlags().MarkHidden("state-dir")
	cmd.PersistentFlags().MarkHidden("debug")
	cmd.PersistentFlags().MarkHidden("verbose")

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
		newLocalNodeStatusCommand(env),
		newGenerateDocsCmd(env),
		newHelmCmd(env),
		xPrintShimPidsCmd,
		newXSnapdConfigCmd(env),
		newXWaitForCmd(env),
		newXCAPICmd(env),
		newListImagesCmd(env),
		newXCleanupCmd(env),
	)

	cmd.DisableAutoGenTag = true
	return cmd
}
