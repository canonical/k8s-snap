package setup

import (
	"context"
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/k8s/certutils"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/k8s/pkg/k8s/utils"
)

// InitKubeconfigs generates the kubeconfig files that services use to communicate with the apiserver.
func InitKubeconfigs(ctx context.Context, client *client.Client, ca *certutils.CertKeyPair) error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	type KubeconfigArgs struct {
		username string
		groups   []string
		path     string
	}

	configs := []KubeconfigArgs{
		KubeconfigArgs{
			username: "kubernetes-admin",
			groups:   []string{"system:masters"},
			path:     "/etc/kubernetes/admin.conf",
		},
		KubeconfigArgs{
			username: "system:kube-controller-manager",
			groups:   []string{},
			path:     "/etc/kubernetes/controller-manager.conf",
		},
		KubeconfigArgs{
			username: "system:kube-proxy",
			groups:   []string{},
			path:     "/etc/kubernetes/proxy.conf",
		},
		KubeconfigArgs{
			username: "system:kube-scheduler",
			groups:   []string{},
			path:     "/etc/kubernetes/scheduler.conf",
		},
		KubeconfigArgs{
			username: fmt.Sprintf("system:node:%s", hostname),
			groups:   []string{"system:nodes"},
			path:     "/etc/kubernetes/kubelet.conf",
		},
	}

	for _, config := range configs {
		token, err := client.GenerateAuthToken(ctx, config.username, config.groups)
		if err != nil {
			return fmt.Errorf("could not generate auth token for %s: %w", config.username, err)
		}

		err = utils.GenerateKubeconfig(token, ca.CertPem, config.path)
		if err != nil {
			return fmt.Errorf("failed to generate kubeconfig for %s: %w", config.username, err)
		}
	}

	return nil
}
