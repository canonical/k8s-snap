package k8s

import (
	"fmt"
	"slices"
	"strings"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/cmd/k8s/formatter"
	"github.com/canonical/k8s/pkg/utils/vals"
	"github.com/spf13/cobra"
)

var (
	componentList = []string{"network", "dns", "gateway", "ingress", "local-storage", "load-balancer", "metrics-server"}
)

type EnableResult struct {
	Functionalities []string `json:"functionalities" yaml:"functionalities"`
}

func (e EnableResult) String() string {
	return fmt.Sprintf("%s enabled.\n", strings.Join(e.Functionalities, ", "))
}

func newEnableCmd() *cobra.Command {
	enableCmd := &cobra.Command{
		Use:     "enable <functionality1> ...",
		Short:   "Enable core cluster functionalities",
		Long:    fmt.Sprintf("Enable one of %s.", strings.Join(componentList, ", ")),
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			if len(args) < 1 {
				return fmt.Errorf("missing argument: provide the names of the functionalites that should be enabled")
			}

			config := api.UserFacingClusterConfig{}
			functionalities := args
			for _, functionality := range functionalities {
				if !slices.Contains(componentList, functionality) {
					return fmt.Errorf("unknown functionality %q; needs to be one of: %s", args[0], strings.Join(componentList, ", "))
				}

				switch functionality {
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
			}

			request := api.UpdateClusterConfigRequest{
				Config: config,
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Enabling %s. This may take some time, please wait.\n", strings.Join(functionalities, ", "))
			if err := k8sdClient.UpdateClusterConfig(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to update cluster configuration: %w", err)
			}

			f, err := formatter.New(rootCmdOpts.outputFormat, cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}
			return f.Print(EnableResult{
				Functionalities: functionalities,
			})

		},
	}

	return enableCmd
}
