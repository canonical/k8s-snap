package k8s

import (
	"fmt"
	"github.com/canonical/k8s/pkg/utils"
	"strings"

	api "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
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
		Long:   fmt.Sprintf("Disable one of %s.", strings.Join(featureList, ", ")),
		Args:   cmdutil.MinimumNArgs(env, 1),
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Run: func(cmd *cobra.Command, args []string) {
			config := api.UserFacingClusterConfig{}

			for _, feature := range args {
				switch feature {
				case "network":
					config.Network = api.NetworkConfig{
						Enabled: utils.Pointer(false),
					}
				case "dns":
					config.DNS = api.DNSConfig{
						Enabled: utils.Pointer(false),
					}
				case "gateway":
					config.Gateway = api.GatewayConfig{
						Enabled: utils.Pointer(false),
					}
				case "ingress":
					config.Ingress = api.IngressConfig{
						Enabled: utils.Pointer(false),
					}
				case "local-storage":
					config.LocalStorage = api.LocalStorageConfig{
						Enabled: utils.Pointer(false),
					}
				case "load-balancer":
					config.LoadBalancer = api.LoadBalancerConfig{
						Enabled: utils.Pointer(false),
					}
				case "metrics-server":
					config.MetricsServer = api.MetricsServerConfig{
						Enabled: utils.Pointer(false),
					}
				default:
					cmd.PrintErrf("Error: Cannot disable %q, must be one of: %s\n", feature, strings.Join(featureList, ", "))
					env.Exit(1)
					return
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

			cmd.PrintErrf("Disabling %s from the cluster. This may take a few seconds, please wait.\n", strings.Join(args, ", "))
			if err := client.UpdateClusterConfig(cmd.Context(), request); err != nil {
				cmd.PrintErrf("Error: Failed to disable %s from the cluster.\n\nThe error was: %v\n", strings.Join(args, ", "), err)
				env.Exit(1)
				return
			}

			outputFormatter.Print(DisableResult{Features: args})
		},
	}

	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")

	return cmd
}
