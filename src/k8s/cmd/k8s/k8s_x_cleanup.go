package k8s

import (
	"context"
	"time"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
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

			if err := features.Cleanup.CleanupNetwork(ctx, env.Snap); err != nil {
				cmd.PrintErrf("Error: failed to cleanup network: %v\n", err)
				env.Exit(1)
			}
		},
	}
	cleanupNetworkCmd.Flags().DurationVar(&opts.timeout, "timeout", 5*time.Minute, "the max time to wait for the command to execute")

	cmd := &cobra.Command{
		Use:    "x-cleanup",
		Short:  "Cleanup left-over resources from the cluster's features",
		Hidden: true,
	}

	cmd.AddCommand(cleanupNetworkCmd)

	return cmd
}
