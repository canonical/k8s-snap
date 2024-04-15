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
	Features []string `json:"features" yaml:"features"`
}

func (e EnableResult) String() string {
	return fmt.Sprintf("%s enabled.\n", strings.Join(e.Features, ", "))
}

func newEnableCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		outputFormat string
	}
	cmd := &cobra.Command{
		Use:    "enable <feature> ...",
		Short:  "Enable core cluster features",
		Long:   fmt.Sprintf("Enable one of %s.", strings.Join(componentList, ", ")),
		Args:   cmdutil.MinimumNArgs(env, 1),
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Run: func(cmd *cobra.Command, args []string) {
			config := api.UserFacingClusterConfig{}
			features := args
			for _, feature := range features {
				if !slices.Contains(componentList, feature) {
					cmd.PrintErrf("Error: Cannot enable %q, must be one of: %s\n", feature, strings.Join(componentList, ", "))
					env.Exit(1)
					return
				}

				switch feature {
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

			cmd.PrintErrf("Enabling %s on the cluster. This may take a few seconds, please wait.\n", strings.Join(features, ", "))
			if err := client.UpdateClusterConfig(cmd.Context(), request); err != nil {
				cmd.PrintErrf("Error: Failed to enable %s on the cluster.\n\nThe error was: %v\n", strings.Join(features, ", "), err)
				env.Exit(1)
				return
			}

			outputFormatter.Print(EnableResult{Features: features})
		},
	}

	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")

	return cmd
}
