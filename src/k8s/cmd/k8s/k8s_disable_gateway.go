package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newDisableGatewayCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "gateway",
		Short:             "Disable the Gateway component in the cluster.",
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateGatewayComponentRequest{
				Status: api.ComponentDisable,
			}

			if err := k8sdClient.UpdateGatewayComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to disable Gateway component: %w", err)
			}

			cmd.Println("Component 'Gateway' disabled")
			return nil
		},
	}
}
