package k8s

import (
	"fmt"

	v1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/cmd/k8s/formatter"
	"github.com/spf13/cobra"
)

var (
	statusCmdOpts struct {
		waitReady bool
	}
)

func newStatusCmd() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:     "status",
		Short:   "Retrieve the current status of the cluster",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			// fail fast if we're not bootstrapped
			if !k8sdClient.IsBootstrapped(cmd.Context()) {
				return v1.ErrNotBootstrapped
			}
			// fail fast if we're not explicitly waiting and we can't get kube-apiserver endpoints
			if !statusCmdOpts.waitReady {
				if ready := k8sdClient.IsKubernetesAPIServerReady(cmd.Context()); !ready {
					return fmt.Errorf("failed to get kube-apiserver endpoints; cluster status is unavailable")
				}
			}

			clusterStatus, err := k8sdClient.ClusterStatus(cmd.Context(), statusCmdOpts.waitReady)
			if err != nil {
				return fmt.Errorf("failed to get cluster status: %w", err)
			}

			f, err := formatter.New(rootCmdOpts.outputFormat, cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}
			return f.Print(clusterStatus)
		},
	}
	statusCmd.PersistentFlags().BoolVar(&statusCmdOpts.waitReady, "wait-ready", false, "wait until at least one cluster node is ready")
	return statusCmd
}
