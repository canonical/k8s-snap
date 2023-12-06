package k8s

import (
	"github.com/canonical/k8s/pkg/k8s/component"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {

	enableCmd := &cobra.Command{
		Use:       "enable <component>",
		Short:     "Enable a specific component in the cluster",
		Long:      "Enable one of the specific components: cni, dns, gateway, ingress, rbac or storage.",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"cni", "dns", "gateway", "ingress", "rbac", "storage"},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := component.EnableComponent(name, map[string]any{}); err != nil {
				return err
			}

			logrus.WithField("component", name).Info("Component enabled")
			return nil

		},
	}

	rootCmd.AddCommand(enableCmd)
}
