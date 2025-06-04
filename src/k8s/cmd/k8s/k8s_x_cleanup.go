package k8s

import (
	"context"
	"time"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/snap/util/cleanup"
	"github.com/spf13/cobra"
)

func newXCleanupCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		timeout time.Duration
	}

	cleanupNetworkCmd := &cobra.Command{
		Use:   string(features.Network),
		Short: "Cleanup left-over network resources",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			defer cancel()

			cleanup.RemoveKubeProxyRules(ctx, env.Snap)

			if err := features.Cleanup.CleanupNetwork(ctx, env.Snap); err != nil {
				cmd.PrintErrf("Error: failed to cleanup network: %v\n", err)
				env.Exit(1)
			}
		},
	}
	cleanupNetworkCmd.Flags().DurationVar(&opts.timeout, "timeout", 5*time.Minute, "the max time to wait for the command to execute")

	cleanupContainersCmd := &cobra.Command{
		Use:   "containers",
		Short: "Cleanup left-over containers",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			defer cancel()

			cleanup.TryCleanupContainers(ctx, env.Snap)
		},
	}
	cleanupContainersCmd.Flags().DurationVar(&opts.timeout, "timeout", 5*time.Minute, "the max time to wait for the command to execute")

	cleanupContainerdCmd := &cobra.Command{
		Use:   "containerd",
		Short: "Cleanup containerd left-over resources",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			defer cancel()

			cleanup.TryCleanupContainerdPaths(ctx, env.Snap)
		},
	}
	cleanupContainerdCmd.Flags().DurationVar(&opts.timeout, "timeout", 5*time.Minute, "the max time to wait for the command to execute")

	cmd := &cobra.Command{
		Use:    "x-cleanup",
		Short:  "Cleanup left-over resources from the cluster's features",
		Hidden: true,
	}

	cmd.AddCommand(cleanupNetworkCmd)
	cmd.AddCommand(cleanupContainersCmd)
	cmd.AddCommand(cleanupContainerdCmd)

	return cmd
}
