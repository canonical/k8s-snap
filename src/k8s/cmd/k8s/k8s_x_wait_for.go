package k8s

import (
	"context"
	"time"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/spf13/cobra"
)

func newXWaitForCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		timeout time.Duration
	}
	waitForDNSCmd := &cobra.Command{
		Use:   "dns",
		Short: "Wait for DNS to be ready",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			defer cancel()
			if err := control.WaitUntilReady(ctx, func() (bool, error) {
				return features.StatusChecks.CheckDNS(cmd.Context(), env.Snap)
			}); err != nil {
				cmd.PrintErrf("Error: failed to wait for DNS to be ready: %v\n", err)
				env.Exit(1)
			}
		},
	}
	waitForDNSCmd.Flags().DurationVar(&opts.timeout, "timeout", 5*time.Minute, "maximum time to wait")

	waitForNetworkCmd := &cobra.Command{
		Use:   "network",
		Short: "Wait for Network to be ready",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			defer cancel()
			if err := control.WaitUntilReady(ctx, func() (bool, error) {
				return features.StatusChecks.CheckNetwork(cmd.Context(), env.Snap)
			}); err != nil {
				cmd.PrintErrf("Error: failed to wait for network to be ready: %v\n", err)
				env.Exit(1)
			}
		},
	}
	waitForNetworkCmd.Flags().DurationVar(&opts.timeout, "timeout", 5*time.Minute, "maximum time to wait")

	cmd := &cobra.Command{
		Use:    "x-wait-for",
		Short:  "Wait for the cluster's feature to be in a ready state",
		Hidden: true,
	}

	cmd.AddCommand(waitForDNSCmd)
	cmd.AddCommand(waitForNetworkCmd)

	return cmd
}
