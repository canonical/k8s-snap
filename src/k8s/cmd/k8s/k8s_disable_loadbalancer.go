package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newDisableLoadBalancerCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "loadbalancer",
		Short:   "Disable the LoadBalancer component in the cluster.",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateLoadBalancerComponentRequest{
				Status: api.ComponentDisable,
			}

			if err := k8sdClient.UpdateLoadBalancerComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to disable LoadBalancer component: %w", err)
			}

			cmd.Println("Component 'LoadBalancer' disabled")
			return nil
		},
	}
}
