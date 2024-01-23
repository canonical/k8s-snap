package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var disableNetworkCmd = &cobra.Command{
	Use:   "network",
	Short: "Disable the Network component in the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
			StorageDir: clusterCmdOpts.storageDir,
			Verbose:    rootCmdOpts.logVerbose,
			Debug:      rootCmdOpts.logDebug,
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		request := api.UpdateNetworkComponentRequest{
			Status: api.ComponentDisable,
		}

		err = client.UpdateNetworkComponent(cmd.Context(), request)
		if err != nil {
			return fmt.Errorf("failed to disable Network component: %w", err)
		}

		cmd.Println("Component 'Network' disabled")
		return nil
	},
}
