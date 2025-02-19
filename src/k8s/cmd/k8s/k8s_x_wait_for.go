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
		Use:   string(features.DNS),
		Short: "Wait for DNS to be ready",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			defer cancel()
			if err := control.WaitUntilReady(ctx, func() (bool, error) {
				err := features.StatusChecks.CheckDNS(cmd.Context(), env.Snap)
				if err != nil {
					cmd.PrintErrf("DNS not ready yet: %v\n", err.Error())
				}
				return err == nil, nil
			}); err != nil {
				cmd.PrintErrf("Error: DNS did not become ready: %v\n", err)
				env.Exit(1)
			}
		},
	}
	waitForDNSCmd.Flags().DurationVar(&opts.timeout, "timeout", 5*time.Minute, "maximum time to wait")

	waitForNetworkCmd := &cobra.Command{
		Use:   string(features.Network),
		Short: "Wait for Network to be ready",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			defer cancel()
			if err := control.WaitUntilReady(ctx, func() (bool, error) {
				err := features.StatusChecks.CheckNetwork(cmd.Context(), env.Snap)
				if err != nil {
					cmd.PrintErrf("network not ready yet: %v\n", err.Error())
				}
				return err == nil, nil
			}); err != nil {
				cmd.PrintErrf("Error: network did not become ready: %v\n", err)
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
