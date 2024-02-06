package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var disableGatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Disable the Gateway component in the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
			StateDir: clusterCmdOpts.stateDir,
			Verbose:  rootCmdOpts.logVerbose,
			Debug:    rootCmdOpts.logDebug,
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		request := api.UpdateGatewayComponentRequest{
			Status: api.ComponentDisable,
		}

		if err := client.UpdateGatewayComponent(cmd.Context(), request); err != nil {
			return fmt.Errorf("failed to disable Gateway component: %w", err)
		}

		cmd.Println("Component 'Gateway' disabled")
		return nil
	},
}
