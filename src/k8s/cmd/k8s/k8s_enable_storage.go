package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newEnableStorageCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "storage",
		Short:   "Enable the Storage component in the cluster.",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateStorageComponentRequest{
				Status: api.ComponentEnable,
			}

			if err := k8sdClient.UpdateStorageComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to enable Storage component: %w", err)
			}

			cmd.Println("Component 'Storage' enabled")
			return nil
		},
	}
}
