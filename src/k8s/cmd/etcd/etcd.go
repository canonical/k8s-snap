package etcd

import (
	"context"
	"os"
	"os/signal"
	"time"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/etcd"
	"github.com/canonical/k8s/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

var rootCmdOpts struct {
	logDebug   bool
	logVerbose bool
	logLevel   int
	stateDir   string
}

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
	cmd := &cobra.Command{
		Use:   "k8sd",
		Short: "Canonical Kubernetes orchestrator and clustering daemon",
		Run: func(cmd *cobra.Command, args []string) {
			// configure logging
			log.Configure(log.Options{
				LogLevel:     rootCmdOpts.logLevel,
				AddDirHeader: true,
			})

			instance, err := etcd.New(rootCmdOpts.stateDir)
			if err != nil {
				cmd.PrintErrf("Error: Failed to initialize etcd: %v", err)
				env.Exit(1)
				return
			}

			ctx, cancel := context.WithCancel(cmd.Context())
			if err := instance.Start(ctx); err != nil {
				logrus.WithError(err).Fatal("Server failed to start")
			}

			// Cancel context if we receive an exit signal
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, unix.SIGPWR)
			signal.Notify(ch, unix.SIGINT)
			signal.Notify(ch, unix.SIGQUIT)
			signal.Notify(ch, unix.SIGTERM)

			select {
			case <-ch:
			case <-instance.MustStop():
			}
			cancel()

			// Create a separate context with 30 seconds to cleanup
			stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := instance.Shutdown(stopCtx); err != nil {
				logrus.WithError(err).Fatal("Failed to shutdown server")
			}
		},
	}

	cmd.SetIn(env.Stdin)
	cmd.SetOut(env.Stdout)
	cmd.SetErr(env.Stderr)

	cmd.PersistentFlags().IntVarP(&rootCmdOpts.logLevel, "log-level", "l", 0, "etcd log level")
	cmd.PersistentFlags().BoolVarP(&rootCmdOpts.logDebug, "debug", "d", false, "Show all debug messages")
	cmd.PersistentFlags().BoolVarP(&rootCmdOpts.logVerbose, "verbose", "v", true, "Show all information messages")
	cmd.PersistentFlags().StringVar(&rootCmdOpts.stateDir, "state-dir", "", "Directory with the etcd datastore")

	return cmd
}
