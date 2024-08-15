package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/utils"

	api "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
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
		timeout      time.Duration
	}
	cmd := &cobra.Command{
		Use:    fmt.Sprintf("enable [%s] ...", strings.Join(featureList, "|")),
		Short:  "Enable core cluster features",
		Long:   fmt.Sprintf("Enable one of %s.", strings.Join(featureList, ", ")),
		Args:   cmdutil.MinimumNArgs(env, 1),
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Run: func(cmd *cobra.Command, args []string) {
			config := api.UserFacingClusterConfig{}

			if opts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", opts.timeout, minTimeout, minTimeout)
				opts.timeout = minTimeout
			}

			for _, feature := range args {
				switch feature {
				case string(features.Network):
					config.Network = api.NetworkConfig{
						Enabled: utils.Pointer(true),
					}
				case string(features.DNS):
					config.DNS = api.DNSConfig{
						Enabled: utils.Pointer(true),
					}
				case string(features.Gateway):
					config.Gateway = api.GatewayConfig{
						Enabled: utils.Pointer(true),
					}
				case string(features.Ingress):
					config.Ingress = api.IngressConfig{
						Enabled: utils.Pointer(true),
					}
				case string(features.LocalStorage):
					config.LocalStorage = api.LocalStorageConfig{
						Enabled: utils.Pointer(true),
					}
				case string(features.LoadBalancer):
					config.LoadBalancer = api.LoadBalancerConfig{
						Enabled: utils.Pointer(true),
					}
				case string(features.MetricsServer):
					config.MetricsServer = api.MetricsServerConfig{
						Enabled: utils.Pointer(true),
					}
				default:
					cmd.PrintErrf("Error: Cannot enable %q, must be one of: %s\n", feature, strings.Join(featureList, ", "))
					env.Exit(1)
					return
				}
			}
			request := api.UpdateClusterConfigRequest{
				Config: config,
			}

			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			cmd.PrintErrf("Enabling %s on the cluster. This may take a few seconds, please wait.\n", strings.Join(args, ", "))
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			cobra.OnFinalize(cancel)
			if err := client.SetClusterConfig(ctx, request); err != nil {
				cmd.PrintErrf("Error: Failed to enable %s on the cluster.\n\nThe error was: %v\n", strings.Join(args, ", "), err)
				env.Exit(1)
				return
			}

			outputFormatter.Print(EnableResult{Features: args})
		},
	}

	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")

	return cmd
}
