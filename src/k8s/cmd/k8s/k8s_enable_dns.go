package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var dnsConfig struct {
	ServiceIP           string
	ClusterDomain       string
	UpstreamNameservers []string
}
var enableDNSCmd = &cobra.Command{
	Use:   "dns",
	Short: "Enable the DNS component in the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
			StorageDir: clusterCmdOpts.storageDir,
			Verbose:    rootCmdOpts.logVerbose,
			Debug:      rootCmdOpts.logDebug,
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		request := api.UpdateDNSComponentRequest{
			Status: api.ComponentEnable,
			Config: api.DNSComponentConfig{
				ServiceIP:           dnsConfig.ServiceIP,
				UpstreamNameservers: dnsConfig.UpstreamNameservers,
				ClusterDomain:       dnsConfig.ClusterDomain,
			},
		}

		err = client.UpdateDNSComponent(cmd.Context(), request)
		if err != nil {
			return fmt.Errorf("failed to enable DNS component: %w", err)
		}

		cmd.Println("DNS component enabled")
		return nil
	},
}

func init() {
	enableDNSCmd.Flags().StringVar(&dnsConfig.ServiceIP, "service-ip", "", "IP address to assign to the DNS service")
	enableDNSCmd.Flags().StringSliceVar(&dnsConfig.UpstreamNameservers, "upstream-nameservers", []string{}, "Upstream nameservers for the DNS service")
	enableDNSCmd.Flags().StringVar(&dnsConfig.ClusterDomain, "cluster-domain", "", "Cluster DNS domain")
}
