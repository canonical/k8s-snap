package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var enableStorageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Enable the Storage component in the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
			StorageDir: clusterCmdOpts.storageDir,
			Verbose:    rootCmdOpts.logVerbose,
			Debug:      rootCmdOpts.logDebug,
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		request := api.UpdateStorageComponentRequest{
			Status: api.ComponentEnable,
		}

		err = client.UpdateStorageComponent(cmd.Context(), request)
		if err != nil {
			return fmt.Errorf("failed to enable Storage component: %w", err)
		}

		cmd.Println("Component 'Storage' enabled")
		return nil
	},
}
