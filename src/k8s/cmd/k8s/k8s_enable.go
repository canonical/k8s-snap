package k8s

import (
	"fmt"
	"slices"
	"strings"

	api "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/utils/vals"
	"github.com/spf13/cobra"
)

type EnableResult struct {
	Functionalities []string `json:"functionalities" yaml:"functionalities"`
}

func (e EnableResult) String() string {
	return fmt.Sprintf("%s enabled.\n", strings.Join(e.Functionalities, ", "))
}

func newEnableCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "enable <functionality> ...",
		Short:  "Enable core cluster functionalities",
		Long:   fmt.Sprintf("Enable one of %s.", strings.Join(componentList, ", ")),
		Args:   cmdutil.MinimumNArgs(env, 1),
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			config := api.UserFacingClusterConfig{}
			functionalities := args
			for _, functionality := range functionalities {
				if !slices.Contains(componentList, functionality) {
					cmd.PrintErrf("Error: Cannot enable %q, must be one of: %s\n", functionality, strings.Join(componentList, ", "))
					env.Exit(1)
					return
				}

				switch functionality {
				case "network":
					config.Network = api.NetworkConfig{
						Enabled: vals.Pointer(true),
					}
				case "dns":
					config.DNS = api.DNSConfig{
						Enabled: vals.Pointer(true),
					}
				case "gateway":
					config.Gateway = api.GatewayConfig{
						Enabled: vals.Pointer(true),
					}
				case "ingress":
					config.Ingress = api.IngressConfig{
						Enabled: vals.Pointer(true),
					}
				case "local-storage":
					config.LocalStorage = api.LocalStorageConfig{
						Enabled: vals.Pointer(true),
					}
				case "load-balancer":
					config.LoadBalancer = api.LoadBalancerConfig{
						Enabled: vals.Pointer(true),
					}
				case "metrics-server":
					config.MetricsServer = api.MetricsServerConfig{
						Enabled: vals.Pointer(true),
					}
				}
			}
			request := api.UpdateClusterConfigRequest{
				Config: config,
			}

			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			cmd.PrintErrf("Enabling %s on the cluster. This may take a few seconds, please wait.\n", strings.Join(functionalities, ", "))
			if err := client.UpdateClusterConfig(cmd.Context(), request); err != nil {
				cmd.PrintErrf("Error: Failed to enable %s on the cluster.\n\nThe error was: %v\n", strings.Join(functionalities, ", "), err)
				env.Exit(1)
				return
			}

			if err := cmdutil.FormatterFromContext(cmd.Context()).Print(EnableResult{Functionalities: functionalities}); err != nil {
				cmd.PrintErrf("WARNING: Failed to print the enable result.\n\nThe error was: %v\n", err)
			}
		},
	}

	return cmd
}
