package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var disableDNSCmd = &cobra.Command{
	Use:   "dns",
	Short: "Disable the DNS component in the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
			StateDir: clusterCmdOpts.stateDir,
			Verbose:  rootCmdOpts.logVerbose,
			Debug:    rootCmdOpts.logDebug,
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		request := api.UpdateDNSComponentRequest{
			Status: api.ComponentDisable,
		}

		err = client.UpdateDNSComponent(cmd.Context(), request)
		if err != nil {
			return fmt.Errorf("failed to disable DNS component: %w", err)
		}

		cmd.Println("Component 'DNS' disabled")
		return nil
	},
}
