package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newDisableIngressCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "ingress",
		Short:             "Disable the Ingress component in the cluster.",
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateIngressComponentRequest{
				Status: api.ComponentDisable,
			}

			if err := k8sdClient.UpdateIngressComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to disable Ingress component: %w", err)
			}

			cmd.Println("Component 'Ingress' disabled")
			return nil
		},
	}
}
