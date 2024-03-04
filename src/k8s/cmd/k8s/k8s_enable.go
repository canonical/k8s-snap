package k8s

import (
	"fmt"
	"slices"
	"strings"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/pkg/utils/vals"
	"github.com/spf13/cobra"
)

var (
	componentList      = []string{"network", "dns", "gateway", "ingress", "local-storage", "load-balancer", "metrics-server"}
	enableCmdErrorMsgs = map[error]string{
		api.ErrUnknown: "An error occurred while calling enable:\n",
	}
)

func newEnableCmd() *cobra.Command {
	enableCmd := &cobra.Command{
		Use:     "enable <functionality>",
		Short:   "Enable a specific functionality in the cluster",
		Long:    fmt.Sprintf("Enable one of the specific functionalities: %s.", strings.Join(componentList, ", ")),
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, enableCmdErrorMsgs)

			if len(args) > 1 {
				return fmt.Errorf("Too many arguments. Please enable one thing at a time.")
			}
			if len(args) < 1 {
				return fmt.Errorf("Please provide the name of what need to be enabled.")
			}
			if !slices.Contains(componentList, args[0]) {
				return fmt.Errorf("Cannot enable %q. Please select one of: %s", args[0], strings.Join(componentList, ", "))
			}

			config := api.UserFacingClusterConfig{}
			switch args[0] {
			case "network":
				config.Network = &api.NetworkConfig{
					Enabled: vals.Pointer(true),
				}
			case "dns":
				config.DNS = &api.DNSConfig{
					Enabled: vals.Pointer(true),
				}
			case "gateway":
				config.Gateway = &api.GatewayConfig{
					Enabled: vals.Pointer(true),
				}
			case "ingress":
				config.Ingress = &api.IngressConfig{
					Enabled: vals.Pointer(true),
				}
			case "local-storage":
				config.LocalStorage = &api.LocalStorageConfig{
					Enabled: vals.Pointer(true),
				}
			case "load-balancer":
				config.LoadBalancer = &api.LoadBalancerConfig{
					Enabled: vals.Pointer(true),
				}
			case "metrics-server":
				config.MetricsServer = &api.MetricsServerConfig{
					Enabled: vals.Pointer(true),
				}
			}

			request := api.UpdateClusterConfigRequest{
				Config: config,
			}

			if err := k8sdClient.UpdateClusterConfig(cmd.Context(), request); err != nil {
				return fmt.Errorf("Failed to update cluster configuration: %w", err)
			}

			return nil
		},
	}

	return enableCmd
}
