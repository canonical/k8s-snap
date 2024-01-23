package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var (
	configCmd = &cobra.Command{
		Use:    "config",
		Short:  "Generate a kubeconfig that can be used to access the Kubernetes cluster",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.NewClient(cmd.Context(), client.ClusterOpts{
				StateDir: clusterCmdOpts.stateDir,
				Verbose:  rootCmdOpts.logVerbose,
				Debug:    rootCmdOpts.logDebug,
			})
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			adminConfig, err := c.KubeConfig(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get admin config: %w", err)
			}

			fmt.Println(adminConfig)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(configCmd)
}
