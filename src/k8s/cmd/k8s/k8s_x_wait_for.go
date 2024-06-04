package k8s

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/spf13/cobra"
)

func newXWaitForCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	waitForDNSCmd := &cobra.Command{
		Use:    "dns",
		Short:  "Wait for DNS to be ready",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			err := control.WaitUntilReady(cmd.Context(), func() (bool, error) {
				return features.StatusChecks.CheckDNS(cmd.Context(), env.Snap)
			})
			if err != nil {
				cmd.PrintErrf("Error: failed to wait for DNS to be ready: %v\n", err)
				env.Exit(1)
			}
		},
	}

	waitForNetworkCmd := &cobra.Command{
		Use:    "network",
		Short:  "Wait for Network to be ready",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			err := control.WaitUntilReady(cmd.Context(), func() (bool, error) {
				return features.StatusChecks.CheckNetwork(cmd.Context(), env.Snap)
			})
			if err != nil {
				cmd.PrintErrf("Error: failed to wait for DNS to be ready: %v\n", err)
				env.Exit(1)
			}
		},
	}

	cmd := &cobra.Command{
		Use:    "x-wait-for",
		Short:  "Wait for the cluster's feature to be in a ready state",
		Hidden: true,
	}

	cmd.AddCommand(waitForDNSCmd)
	cmd.AddCommand(waitForNetworkCmd)

	return cmd
}
