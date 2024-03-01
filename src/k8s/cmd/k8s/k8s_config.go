package k8s

import (
	"fmt"

	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

var (
	configCmdOpts struct {
		server string
	}
)

func newKubeConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:     "config --server <server>",
		Short:   "Generate a kubeconfig that can be used to access the Kubernetes cluster",
		Hidden:  true,
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			adminConfig, err := k8sdClient.KubeConfig(cmd.Context(), configCmdOpts.server)
			if err != nil {
				return fmt.Errorf("failed to get admin config: %w", err)
			}

			fmt.Println(adminConfig)
			return nil
		},
	}
	configCmd.PersistentFlags().StringVar(&configCmdOpts.server, "server", "", "Specify a custom cluster server address for the kubeconfig")
	return configCmd
}
