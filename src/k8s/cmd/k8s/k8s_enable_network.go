package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newEnableNetworkCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "network",
		Short:             "Enable the Network component in the cluster.",
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateNetworkComponentRequest{
				Status: api.ComponentEnable,
			}

			if err := k8sdClient.UpdateNetworkComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to enable Network component: %w", err)
			}

			cmd.Println("Component 'Network' enabled")
			return nil
		},
	}
}
