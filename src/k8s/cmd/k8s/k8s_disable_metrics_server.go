package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newDisableMetricsServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "metrics-server",
		Short:   "Disable the Metrics-Server component in the cluster.",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateMetricsServerComponentRequest{
				Status: api.ComponentDisable,
			}

			if err := k8sdClient.UpdateMetricsServerComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to disable Metrics-Server component: %w", err)
			}

			cmd.Println("Component 'Metrics-Server' disabled")
			return nil
		},
	}
}
