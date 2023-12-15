package setup

import (
	"fmt"

	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/k8s/utils"
	"github.com/canonical/k8s/pkg/snap"
)

// InitKubeApiserver handles the setup of kube-apiserver.
//   - Sets up the token webhook authentication.
func InitKubeApiserver() error {
	defaultIp, err := utils.GetDefaultIP()
	if err != nil {
		return fmt.Errorf("failed to get default ip: %w", err)
	}

	utils.TemplateAndSave(snap.Path("k8s/config/apiserver-token-hook.tmpl"),
		struct {
			WebhookIp   string
			WebhookPort int
		}{
			WebhookIp:   defaultIp.String(),
			WebhookPort: config.DefaultPort,
		},
		"/etc/kubernetes/apiserver-token-hook.conf",
	)

	return nil
}
