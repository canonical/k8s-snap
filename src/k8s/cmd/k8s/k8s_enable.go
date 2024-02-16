package k8s

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	componentList = []string{"network", "dns", "gateway", "ingress", "rbac", "storage", "loadbalancer"}

	enableCmd = &cobra.Command{
		Use:       "enable <component>",
		Short:     "Enable a specific component in the cluster",
		Long:      fmt.Sprintf("Enable one of the specific components: %s.", strings.Join(componentList, ",")),
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: componentList,
	}
)

func init() {
	rootCmd.AddCommand(enableCmd)
	enableCmd.AddCommand(enableDNSCmd)
	enableCmd.AddCommand(enableNetworkCmd)
	enableCmd.AddCommand(enableStorageCmd)
	enableCmd.AddCommand(enableIngressCmd)
	enableCmd.AddCommand(enableGatewayCmd)
	enableCmd.AddCommand(enableLoadBalancerCmd)
}
