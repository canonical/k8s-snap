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
	Features []string `json:"features" yaml:"features"`
}

func (d DisableResult) String() string {
	return fmt.Sprintf("%s disabled.\n", strings.Join(d.Features, ", "))
}

func newDisableCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		outputFormat string
	}
	cmd := &cobra.Command{
		Use:    "disable <feature> ...",
		Short:  "Disable core cluster features",
		Long:   fmt.Sprintf("Disable one of %s.", strings.Join(componentList, ", ")),
		Args:   cmdutil.MinimumNArgs(env, 1),
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, opts.outputFormat)),
		Run: func(cmd *cobra.Command, args []string) {
			config := api.UserFacingClusterConfig{}
			features := args
			for _, feature := range features {
				if !slices.Contains(componentList, feature) {
					cmd.PrintErrf("Error: Cannot disable %q, must be one of: %s\n", feature, strings.Join(componentList, ", "))
					env.Exit(1)
					return
				}

				switch feature {
				case "network":
					config.Network = api.NetworkConfig{
						Enabled: vals.Pointer(false),
					}
				case "dns":
					config.DNS = api.DNSConfig{
						Enabled: vals.Pointer(false),
					}
				case "gateway":
					config.Gateway = api.GatewayConfig{
						Enabled: vals.Pointer(false),
					}
				case "ingress":
					config.Ingress = api.IngressConfig{
						Enabled: vals.Pointer(false),
					}
				case "local-storage":
					config.LocalStorage = api.LocalStorageConfig{
						Enabled: vals.Pointer(false),
					}
				case "load-balancer":
					config.LoadBalancer = api.LoadBalancerConfig{
						Enabled: vals.Pointer(false),
					}
				case "metrics-server":
					config.MetricsServer = api.MetricsServerConfig{
						Enabled: vals.Pointer(false),
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

			cmd.PrintErrf("Disabling %s from the cluster. This may take a few seconds, please wait.\n", strings.Join(features, ", "))
			if err := client.UpdateClusterConfig(cmd.Context(), request); err != nil {
				cmd.PrintErrf("Error: Failed to disable %s from the cluster.\n\nThe error was: %v\n", strings.Join(features, ", "), err)
				env.Exit(1)
				return
			}

			globalFormatter.Print(DisableResult{Features: features})
		},
	}

	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")

	return cmd
}
