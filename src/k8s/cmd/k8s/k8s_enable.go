package k8s

import (
	"fmt"
	"strings"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var componentList = []string{"network", "dns", "gateway", "ingress", "rbac", "storage"}

func init() {
	enableCmd := &cobra.Command{
		Use:       "enable <component>",
		Short:     "Enable a specific component in the cluster",
		Long:      fmt.Sprintf("Enable one of the specific components: %s.", strings.Join(componentList, ",")),
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: componentList,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
				RemoteAddress: clusterCmdOpts.remoteAddress,
				StorageDir:    clusterCmdOpts.storageDir,
				Verbose:       rootCmdOpts.logVerbose,
				Debug:         rootCmdOpts.logDebug,
			})
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			err = client.UpdateComponent(cmd.Context(), name, api.ComponentEnable)
			if err != nil {
				return fmt.Errorf("failed to %s %s: %w", name, api.ComponentEnable, err)
			}

			logrus.WithField("component", name).Info("Component enabled.")
			return nil
		},
	}

	rootCmd.AddCommand(enableCmd)
	enableCmd.AddCommand(enableDNSCmd)
}
