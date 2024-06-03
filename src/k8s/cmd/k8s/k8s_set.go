package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

type SetResult struct {
	ClusterConfig apiv1.UserFacingClusterConfig `json:"cluster-config" yaml:"cluster-config"`
}

func (s SetResult) String() string {
	return "Configuration updated."
}

func newSetCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		outputFormat string
		timeout      time.Duration
	}
	cmd := &cobra.Command{
		Use:    "set <feature.key=value> ...",
		Short:  "Set cluster configuration",
		Long:   fmt.Sprintf("Configure one of %s.\nUse `k8s get` to explore configuration options.", strings.Join(featureList, ", ")),
		Args:   cmdutil.MinimumNArgs(env, 1),
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Run: func(cmd *cobra.Command, args []string) {
			config := apiv1.UserFacingClusterConfig{}

			for _, arg := range args {
				if err := updateConfigMapstructure(&config, arg); err != nil {
					cmd.PrintErrf("Error: Invalid option %q.\n\nThe error was: %v\n", arg, err)
					env.Exit(1)
				}
			}

			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			request := apiv1.UpdateClusterConfigRequest{
				Config: config,
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			cobra.OnFinalize(cancel)

			if err := client.UpdateClusterConfig(ctx, request); err != nil {
				cmd.PrintErrf("Error: Failed to apply requested cluster configuration changes.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			outputFormatter.Print(SetResult{ClusterConfig: config})
		},
	}

	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")

	return cmd
}

var knownSetKeys = map[string]struct{}{
	"annotations":                    {},
	"cloud-provider":                 {},
	"dns.cluster-domain":             {},
	"dns.enabled":                    {},
	"dns.service-ip":                 {},
	"dns.upstream-nameservers":       {},
	"gateway.enabled":                {},
	"ingress.default-tls-secret":     {},
	"ingress.enable-proxy-protocol":  {},
	"ingress.enabled":                {},
	"load-balancer.bgp-local-asn":    {},
	"load-balancer.bgp-mode":         {},
	"load-balancer.bgp-peer-address": {},
	"load-balancer.bgp-peer-asn":     {},
	"load-balancer.bgp-peer-port":    {},
	"load-balancer.cidrs":            {},
	"load-balancer.enabled":          {},
	"load-balancer.l2-interfaces":    {},
	"load-balancer.l2-mode":          {},
	"local-storage.default":          {},
	"local-storage.enabled":          {},
	"local-storage.local-path":       {},
	"local-storage.reclaim-policy":   {},
	"metrics-server.enabled":         {},
	"network.enabled":                {},
}

func updateConfigMapstructure(config *apiv1.UserFacingClusterConfig, arg string) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "json",
		WeaklyTypedInput: true,
		ErrorUnused:      true,
		Result:           config,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			utils.YAMLToStringSliceHookFunc,
			utils.StringToFieldsSliceHookFunc(','),
			utils.YAMLToStringMapHookFunc,
			utils.StringToStringMapHookFunc,
		),
	})
	if err != nil {
		panic(fmt.Sprintf("failed to define decoder with error %v", err.Error()))
	}

	parts := strings.SplitN(arg, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("option not in <key>=<value> format")
	}
	key := parts[0]
	value := parts[1]

	if _, ok := knownSetKeys[key]; !ok {
		return fmt.Errorf("unknown option key %q", key)
	}

	if err := decoder.Decode(toRecursiveMap(key, value)); err != nil {
		return fmt.Errorf("invalid option %q: %w", arg, err)
	}
	return nil
}

func toRecursiveMap(key, value string) map[string]any {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) == 2 {
		return map[string]any{parts[0]: toRecursiveMap(parts[1], value)}
	}
	return map[string]any{key: value}
}
