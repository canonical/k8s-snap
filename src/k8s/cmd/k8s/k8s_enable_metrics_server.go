package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newEnableMetricsServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "metrics-server",
		Short:   "Enable the Metrics-Server component in the cluster.",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateMetricsServerComponentRequest{
				Status: api.ComponentEnable,
			}

			if err := k8sdClient.UpdateMetricsServerComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to enable Metrics-Server component: %w", err)
			}

			cmd.Println("Component 'Metrics-Server' enabled")
			return nil
		},
	}
}
