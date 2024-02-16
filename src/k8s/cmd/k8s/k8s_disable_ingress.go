package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var disableIngressCmd = &cobra.Command{
	Use:   "ingress",
	Short: "Disable the Ingress component in the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
			StateDir: clusterCmdOpts.stateDir,
			Verbose:  rootCmdOpts.logVerbose,
			Debug:    rootCmdOpts.logDebug,
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		request := api.UpdateIngressComponentRequest{
			Status: api.ComponentDisable,
		}

		if err := client.UpdateIngressComponent(cmd.Context(), request); err != nil {
			return fmt.Errorf("failed to disable Ingress component: %w", err)
		}

		cmd.Println("Component 'Ingress' disabled")
		return nil
	},
}
