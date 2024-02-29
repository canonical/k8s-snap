package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newDisableDNSCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "dns",
		Short:   "Disable the DNS component in the cluster.",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateDNSComponentRequest{
				Status: api.ComponentDisable,
			}

			if err := k8sdClient.UpdateDNSComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to disable DNS component: %w", err)
			}

			cmd.Println("Component 'DNS' disabled")
			return nil
		},
	}
}
