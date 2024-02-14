package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

var enableDNSCmdConfig struct {
	ServiceIP           string
	ClusterDomain       string
	UpstreamNameservers []string
}

func newEnableDNSCmd() *cobra.Command {
	enableDNSCmd := &cobra.Command{
		Use:               "dns",
		Short:             "Enable the DNS component in the cluster.",
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateDNSComponentRequest{
				Status: api.ComponentEnable,
				Config: api.DNSComponentConfig{
					ServiceIP:           enableDNSCmdConfig.ServiceIP,
					UpstreamNameservers: enableDNSCmdConfig.UpstreamNameservers,
					ClusterDomain:       enableDNSCmdConfig.ClusterDomain,
				},
			}

			if err := k8sdClient.UpdateDNSComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to enable DNS component: %w", err)
			}

			cmd.Println("Component 'DNS' enabled")
			return nil
		},
	}
	enableDNSCmd.Flags().StringVar(&enableDNSCmdConfig.ServiceIP, "service-ip", "", "IP address to assign to the DNS service")
	enableDNSCmd.Flags().StringSliceVar(&enableDNSCmdConfig.UpstreamNameservers, "upstream-nameservers", []string{}, "Upstream nameservers for the DNS service")
	enableDNSCmd.Flags().StringVar(&enableDNSCmdConfig.ClusterDomain, "cluster-domain", "", "Cluster DNS domain")
	return enableDNSCmd
}
