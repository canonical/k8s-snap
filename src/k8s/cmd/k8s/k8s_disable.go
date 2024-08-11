package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api-v1/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/utils"
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
		timeout      time.Duration
	}
	cmd := &cobra.Command{
		Use:    fmt.Sprintf("disable [%s] ...", strings.Join(featureList, "|")),
		Short:  "Disable core cluster features",
		Long:   fmt.Sprintf("Disable one of %s.", strings.Join(featureList, ", ")),
		Args:   cmdutil.MinimumNArgs(env, 1),
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Run: func(cmd *cobra.Command, args []string) {
			config := apiv1.UserFacingClusterConfig{}

			if opts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", opts.timeout, minTimeout, minTimeout)
				opts.timeout = minTimeout
			}

			for _, feature := range args {
				switch feature {
				case string(features.Network):
					config.Network = apiv1.NetworkConfig{
						Enabled: utils.Pointer(false),
					}
				case string(features.DNS):
					config.DNS = apiv1.DNSConfig{
						Enabled: utils.Pointer(false),
					}
				case string(features.Gateway):
					config.Gateway = apiv1.GatewayConfig{
						Enabled: utils.Pointer(false),
					}
				case string(features.Ingress):
					config.Ingress = apiv1.IngressConfig{
						Enabled: utils.Pointer(false),
					}
				case string(features.LocalStorage):
					config.LocalStorage = apiv1.LocalStorageConfig{
						Enabled: utils.Pointer(false),
					}
				case string(features.LoadBalancer):
					config.LoadBalancer = apiv1.LoadBalancerConfig{
						Enabled: utils.Pointer(false),
					}
				case string(features.MetricsServer):
					config.MetricsServer = apiv1.MetricsServerConfig{
						Enabled: utils.Pointer(false),
					}
				default:
					cmd.PrintErrf("Error: Cannot disable %q, must be one of: %s\n", feature, strings.Join(featureList, ", "))
					env.Exit(1)
					return
				}
			}

			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			cmd.PrintErrf("Disabling %s from the cluster. This may take a few seconds, please wait.\n", strings.Join(args, ", "))
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			cobra.OnFinalize(cancel)
			if err := client.SetClusterConfig(ctx, apiv1.SetClusterConfigRequest{Config: config}); err != nil {
				cmd.PrintErrf("Error: Failed to disable %s from the cluster.\n\nThe error was: %v\n", strings.Join(args, ", "), err)
				env.Exit(1)
				return
			}

			outputFormatter.Print(DisableResult{Features: args})
		},
	}

	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")

	return cmd
}
