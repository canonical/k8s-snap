package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var enableNetworkCmd = &cobra.Command{
	Use:   "network",
	Short: "Enable the Network component in the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
			StateDir: clusterCmdOpts.stateDir,
			Verbose:  rootCmdOpts.logVerbose,
			Debug:    rootCmdOpts.logDebug,
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		request := api.UpdateNetworkComponentRequest{
			Status: api.ComponentEnable,
		}

		if err := client.UpdateNetworkComponent(cmd.Context(), request); err != nil {
			return fmt.Errorf("failed to enable Network component: %w", err)
		}

		cmd.Println("Component 'Network' enabled")
		return nil
	},
}
