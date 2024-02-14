package k8s

import (
	"fmt"

	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newKubeConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "config",
		Short:             "Generate a kubeconfig that can be used to access the Kubernetes cluster",
		Hidden:            true,
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			adminConfig, err := k8sdClient.KubeConfig(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get admin config: %w", err)
			}

			fmt.Println(adminConfig)
			return nil
		},
	}
}
