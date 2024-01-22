package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var enableIngressCmdConfig struct {
	DefaultTLSSecret    string
	EnableProxyProtocol bool
}
var enableIngressCmd = &cobra.Command{
	Use:   "ingress",
	Short: "Enable the Ingress component in the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
			StateDir: clusterCmdOpts.stateDir,
			Verbose:  rootCmdOpts.logVerbose,
			Debug:    rootCmdOpts.logDebug,
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		request := api.UpdateIngressComponentRequest{
			Status: api.ComponentEnable,
			Config: api.IngressComponentConfig{
				DefaultTLSSecret:    enableIngressCmdConfig.DefaultTLSSecret,
				EnableProxyProtocol: enableIngressCmdConfig.EnableProxyProtocol,
			},
		}

		err = client.UpdateIngressComponent(cmd.Context(), request)
		if err != nil {
			return fmt.Errorf("failed to enable Ingress component: %w", err)
		}

		cmd.Println("Component 'Ingress' enabled")
		return nil
	},
}

func init() {
	enableIngressCmd.Flags().StringVar(&enableIngressCmdConfig.DefaultTLSSecret, "default-tls-secret", "", "Name of the TLS Secret in the kube-system namespace that will be used as the default Ingress certificate")
	enableIngressCmd.Flags().BoolVar(&enableIngressCmdConfig.EnableProxyProtocol, "enable-proxy-protocol", false, "If set, proxy protocol will be enabled for the Ingress")
}
