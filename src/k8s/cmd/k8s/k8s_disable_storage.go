package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newDisableStorageCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "storage",
		Short:             "Disable the Network component in the cluster.",
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateStorageComponentRequest{
				Status: api.ComponentDisable,
			}

			if err := k8sdClient.UpdateStorageComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to disable Storage component: %w", err)
			}

			cmd.Println("Component 'Storage' disabled")
			return nil
		},
	}
}
