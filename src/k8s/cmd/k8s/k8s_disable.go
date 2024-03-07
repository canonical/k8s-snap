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

type DisableResult struct {
	Functionalities []string `json:"functionalities" yaml:"functionalities"`
}

func (d DisableResult) String() string {
	return fmt.Sprintf("%s disabled.\n", strings.Join(d.Functionalities, ", "))
}

func newDisableCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "disable <functionality> ...",
		Short:  "Disable core cluster functionalities",
		Long:   fmt.Sprintf("Disable one of %s.", strings.Join(componentList, ", ")),
		Args:   cobra.MinimumNArgs(1),
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			config := api.UserFacingClusterConfig{}
			functionalities := args
			for _, functionality := range functionalities {
				if !slices.Contains(componentList, functionality) {
					cmd.PrintErrf("ERROR: Cannot disable %q, must be one of: %s\n", functionality, strings.Join(componentList, ", "))
					env.Exit(1)
					return
				}

				switch functionality {
				case "network":
					config.Network = &api.NetworkConfig{
						Enabled: vals.Pointer(false),
					}
				case "dns":
					config.DNS = &api.DNSConfig{
						Enabled: vals.Pointer(false),
					}
				case "gateway":
					config.Gateway = &api.GatewayConfig{
						Enabled: vals.Pointer(false),
					}
				case "ingress":
					config.Ingress = &api.IngressConfig{
						Enabled: vals.Pointer(false),
					}
				case "local-storage":
					config.LocalStorage = &api.LocalStorageConfig{
						Enabled: vals.Pointer(false),
					}
				case "load-balancer":
					config.LoadBalancer = &api.LoadBalancerConfig{
						Enabled: vals.Pointer(false),
					}
				case "metrics-server":
					config.MetricsServer = &api.MetricsServerConfig{
						Enabled: vals.Pointer(false),
					}
				}
			}
			request := api.UpdateClusterConfigRequest{
				Config: config,
			}

			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("ERROR: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			cmd.PrintErrf("Disabling %s from the cluster. This may take a few seconds, please wait.\n", strings.Join(functionalities, ", "))
			if err := client.UpdateClusterConfig(cmd.Context(), request); err != nil {
				cmd.PrintErrf("ERROR: Failed to disable %s from the cluster.\n\nThe error was: %v\n", strings.Join(functionalities, ", "), err)
				env.Exit(1)
				return
			}

			if err := cmdutil.FormatterFromContext(cmd.Context()).Print(DisableResult{Functionalities: functionalities}); err != nil {
				cmd.PrintErrf("WARNING: Failed to print the disable result.\n\nThe error was: %v\n", err)
			}
		},
	}

	return cmd
}
