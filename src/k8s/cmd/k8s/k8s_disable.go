package k8s

import (
	"context"
	"fmt"
	"strings"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type disableFunc func(ctx context.Context, client *client.Client) error

var disableActions = map[string]disableFunc{
	"dns":     disableDns,
	"network": disableNetwork,
}

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

			action, ok := disableActions[name]
			if !ok {
				return fmt.Errorf("unsupported component: %s", name)
			}

			err = action(cmd.Context(), client)
			if err != nil {
				return fmt.Errorf("failed to disable %s component: %w", name, err)
			}

			logrus.WithField("component", name).Info("Component disabled.")
			return nil
		},
	}

	rootCmd.AddCommand(disableCmd)
}

func disableDns(ctx context.Context, client *client.Client) error {
	request := api.UpdateDNSComponentRequest{
		Status: api.ComponentDisable,
	}
	return client.UpdateDNSComponent(ctx, request)
}

func disableNetwork(ctx context.Context, client *client.Client) error {
	request := api.UpdateNetworkComponentRequest{
		Status: api.ComponentDisable,
	}
	return client.UpdateNetworkComponent(ctx, request)
}
