package setup

import (
	"fmt"
	"path/filepath"

	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/k8s/pkg/k8s/utils"
)

// InitKubeApiserver handles the setup of kube-apiserver.
//   - Sets up the token webhook authentication.
func InitKubeApiserver() error {
	defaultIp, err := utils.GetDefaultIP()
	if err != nil {
		return fmt.Errorf("failed to get default ip: %w", err)
	}

	utils.TemplateAndSave(filepath.Join(utils.SNAP, "k8s/config/apiserver-token-hook.tmpl"),
		struct {
			WebhookIp   string
			WebhookPort int
		}{
			WebhookIp:   defaultIp.String(),
			WebhookPort: client.DefaultPort,
		},
		"/etc/kubernetes/apiserver-token-hook.conf",
	)

	return nil
}
