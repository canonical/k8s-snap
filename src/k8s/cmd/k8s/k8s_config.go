package k8s

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
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
			snap := snap.NewSnap(os.Getenv("SNAP"), os.Getenv("SNAP_COMMON"))

			isWorker, err := snaputil.IsWorker(snap)
			if err != nil {
				return fmt.Errorf("failed to check if node is a worker: %w", err)
			}

			if isWorker {
				return fmt.Errorf("this action is restricted on workers")
			}

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
