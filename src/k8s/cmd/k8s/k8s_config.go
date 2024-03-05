package k8s

import (
	"fmt"

	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/cmd/k8s/formatter"
	"github.com/spf13/cobra"
)

var (
	configCmdOpts struct {
		server string
	}
)

type KubeConfigResult struct {
	KubeConfig string `json:"kube-config" yaml:"kube-config"`
}

func (k KubeConfigResult) String() string {
	return k.KubeConfig
}

func newKubeConfigCmd() *cobra.Command {
	kubeConfigCmd := &cobra.Command{
		Use:     "config",
		Short:   "Generate a kubeconfig that can be used to access the Kubernetes cluster",
		Hidden:  true,
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			config, err := k8sdClient.KubeConfig(cmd.Context(), configCmdOpts.server)
			if err != nil {
				return fmt.Errorf("failed to get admin config: %w", err)
			}

			f, err := formatter.New(rootCmdOpts.outputFormat, cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}
			return f.Print(KubeConfigResult{
				KubeConfig: config,
			})
		},
	}
	kubeConfigCmd.PersistentFlags().StringVar(&configCmdOpts.server, "server", "", "custom cluster server address")
	return kubeConfigCmd
}
