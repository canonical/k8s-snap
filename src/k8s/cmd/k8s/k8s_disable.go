package k8s

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:       "disable <component>",
	Short:     "Disable a specific component in the cluster",
	Long:      fmt.Sprintf("Disable one of the specific components: %s.", strings.Join(componentList, ",")),
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: componentList,
}

func init() {
	rootCmd.AddCommand(disableCmd)
	disableCmd.AddCommand(disableDNSCmd)
	disableCmd.AddCommand(disableNetworkCmd)
	disableCmd.AddCommand(disableStorageCmd)
	disableCmd.AddCommand(disableIngressCmd)
	disableCmd.AddCommand(disableGatewayCmd)
	disableCmd.AddCommand(disableLoadBalancerCmd)
}
