package k8s

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/spf13/cobra"
)

func newXWaitForDNSCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	waitForDNSCmd := &cobra.Command{
		Use:    "x-wait-for-dns",
		Short:  "Wait for DNS to be ready",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			features.StatusChecks.CheckDNS(cmd.Context(), env.Snap)
		},
	}

	return waitForDNSCmd
}
