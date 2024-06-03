package k8s

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/spf13/cobra"
)

func newXWaitForNetworkCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	waitForNetworkCmd := &cobra.Command{
		Use:    "x-wait-for-network",
		Short:  "Wait for Network to be ready",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			features.StatusChecks.CheckNetwork(cmd.Context(), env.Snap)
		},
	}

	return waitForNetworkCmd
}
