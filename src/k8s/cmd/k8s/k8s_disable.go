package k8s

import (
	"fmt"
	"strings"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {

	disableCmd := &cobra.Command{
		Use:       "disable <component>",
		Short:     "Disable a specific component in the cluster",
		Long:      fmt.Sprintf("Disable one of the specific components: %s.", strings.Join(componentList, ",")),
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

			err = client.UpdateComponent(cmd.Context(), name, api.ComponentDisable)
			if err != nil {
				return fmt.Errorf("failed to %s %s: %w", name, api.ComponentDisable, err)
			}

			logrus.WithField("component", name).Info("Component disabled.")
			return nil
		},
	}

	rootCmd.AddCommand(disableCmd)
}
