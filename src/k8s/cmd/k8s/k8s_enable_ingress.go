package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

var enableIngressCmdOpts struct {
	DefaultTLSSecret    string
	EnableProxyProtocol bool
}

func newEnableIngressCmd() *cobra.Command {
	enableIngressCmd := &cobra.Command{
		Use:               "ingress",
		Short:             "Enable the Ingress component in the cluster",
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateIngressComponentRequest{
				Status: api.ComponentEnable,
				Config: api.IngressComponentConfig{
					DefaultTLSSecret:    enableIngressCmdOpts.DefaultTLSSecret,
					EnableProxyProtocol: enableIngressCmdOpts.EnableProxyProtocol,
				},
			}

			if err := k8sdClient.UpdateIngressComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to enable Ingress component: %w", err)
			}

			cmd.Println("Component 'Ingress' enabled")
			return nil
		},
	}
	enableIngressCmd.Flags().StringVar(&enableIngressCmdOpts.DefaultTLSSecret, "default-tls-secret", "", "Name of the TLS Secret in the kube-system namespace that will be used as the default Ingress certificate")
	enableIngressCmd.Flags().BoolVar(&enableIngressCmdOpts.EnableProxyProtocol, "enable-proxy-protocol", false, "If set, proxy protocol will be enabled for the Ingress")
	return enableIngressCmd
}
