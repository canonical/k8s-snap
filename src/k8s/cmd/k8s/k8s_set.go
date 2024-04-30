package k8s

import (
	"fmt"
	"strings"

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

			if err := client.UpdateClusterConfig(cmd.Context(), request); err != nil {
				cmd.PrintErrf("Error: Failed to apply requested cluster configuration changes.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			outputFormatter.Print(SetResult{ClusterConfig: config})
		},
	}

	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")

	return cmd
}

var knownSetKeys = map[string]struct{}{
	"cloud-provider":                 struct{}{},
	"dns.cluster-domain":             struct{}{},
	"dns.enabled":                    struct{}{},
	"dns.service-ip":                 struct{}{},
	"dns.upstream-nameservers":       struct{}{},
	"gateway.enabled":                struct{}{},
	"ingress.default-tls-secret":     struct{}{},
	"ingress.enable-proxy-protocol":  struct{}{},
	"ingress.enabled":                struct{}{},
	"load-balancer.bgp-local-asn":    struct{}{},
	"load-balancer.bgp-mode":         struct{}{},
	"load-balancer.bgp-peer-address": struct{}{},
	"load-balancer.bgp-peer-asn":     struct{}{},
	"load-balancer.bgp-peer-port":    struct{}{},
	"load-balancer.cidrs":            struct{}{},
	"load-balancer.enabled":          struct{}{},
	"load-balancer.l2-interfaces":    struct{}{},
	"load-balancer.l2-mode":          struct{}{},
	"local-storage.default":          struct{}{},
	"local-storage.enabled":          struct{}{},
	"local-storage.local-path":       struct{}{},
	"local-storage.reclaim-policy":   struct{}{},
	"metrics-server.enabled":         struct{}{},
	"network.enabled":                struct{}{},
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
