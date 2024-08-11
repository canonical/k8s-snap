package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/spf13/cobra"
)

func newGetCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		outputFormat string
		timeout      time.Duration
	}
	cmd := &cobra.Command{
		Use:    "get <feature.key>",
		Short:  "Get cluster configuration",
		Long:   fmt.Sprintf("Show configuration of one of %s.", strings.Join(featureList, ", ")),
		Args:   cmdutil.MaximumNArgs(env, 1),
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Run: func(cmd *cobra.Command, args []string) {
			if opts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", opts.timeout, minTimeout, minTimeout)
				opts.timeout = minTimeout
			}

			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			cobra.OnFinalize(cancel)

			config, err := client.GetClusterConfig(ctx)
			if err != nil {
				cmd.PrintErrf("Error: Failed to get the current cluster configuration.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			config.MetricsServer = apiv1.MetricsServerConfig{}
			config.CloudProvider = nil
			config.Annotations = nil

			var key string
			if len(args) == 1 {
				key = args[0]
			}

			var output any
			switch key {
			case "":
				output = config
			case string(features.Network):
				output = config.Network
			case string(features.DNS):
				output = config.DNS
			case string(features.Gateway):
				output = config.Gateway
			case string(features.Ingress):
				output = config.Ingress
			case string(features.LocalStorage):
				output = config.LocalStorage
			case string(features.LoadBalancer):
				output = config.LoadBalancer
			case fmt.Sprintf("%s.enabled", features.Network):
				output = config.Network.GetEnabled()
			case fmt.Sprintf("%s.enabled", features.DNS):
				output = config.DNS.GetEnabled()
			case fmt.Sprintf("%s.upstream-nameservers", features.DNS):
				output = config.DNS.GetUpstreamNameservers()
			case fmt.Sprintf("%s.cluster-domain", features.DNS):
				output = config.DNS.GetClusterDomain()
			case fmt.Sprintf("%s.service-ip", features.DNS):
				output = config.DNS.GetServiceIP()
			case fmt.Sprintf("%s.enabled", features.Gateway):
				output = config.Gateway.GetEnabled()
			case fmt.Sprintf("%s.enabled", features.Ingress):
				output = config.Ingress.GetEnabled()
			case fmt.Sprintf("%s.default-tls-secret", features.Ingress):
				output = config.Ingress.GetDefaultTLSSecret()
			case fmt.Sprintf("%s.enable-proxy-protocol", features.Ingress):
				output = config.Ingress.GetEnableProxyProtocol()
			case fmt.Sprintf("%s.enabled", features.LocalStorage):
				output = config.LocalStorage.GetEnabled()
			case fmt.Sprintf("%s.local-path", features.LocalStorage):
				output = config.LocalStorage.GetLocalPath()
			case fmt.Sprintf("%s.reclaim-policy", features.LocalStorage):
				output = config.LocalStorage.GetReclaimPolicy()
			case fmt.Sprintf("%s.default", features.LocalStorage):
				output = config.LocalStorage.GetDefault()
			case fmt.Sprintf("%s.enabled", features.LoadBalancer):
				output = config.LoadBalancer.GetEnabled()
			case fmt.Sprintf("%s.cidrs", features.LoadBalancer):
				output = config.LoadBalancer.GetCIDRs()
			case fmt.Sprintf("%s.l2-mode", features.LoadBalancer):
				output = config.LoadBalancer.GetL2Mode()
			case fmt.Sprintf("%s.l2-interfaces", features.LoadBalancer):
				output = config.LoadBalancer.GetL2Interfaces()
			case fmt.Sprintf("%s.bgp-mode", features.LoadBalancer):
				output = config.LoadBalancer.GetBGPMode()
			case fmt.Sprintf("%s.bgp-local-asn", features.LoadBalancer):
				output = config.LoadBalancer.GetBGPLocalASN()
			case fmt.Sprintf("%s.bgp-peer-address", features.LoadBalancer):
				output = config.LoadBalancer.GetBGPPeerAddress()
			case fmt.Sprintf("%s.bgp-peer-port", features.LoadBalancer):
				output = config.LoadBalancer.GetBGPPeerPort()
			case fmt.Sprintf("%s.bgp-peer-asn", features.LoadBalancer):
				output = config.LoadBalancer.GetBGPPeerASN()
			default:
				cmd.PrintErrf("Error: Unknown config key %q.\n", key)
				env.Exit(1)
				return
			}

			outputFormatter.Print(output)
		},
	}
	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")

	return cmd
}
