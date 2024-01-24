package setup

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	apiImpl "github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/cert"
	"github.com/canonical/microcluster/state"
)

// InitKubeconfigs generates the kubeconfig files that services use to communicate with the apiserver.
func InitKubeconfigs(ctx context.Context, state *state.State, ca *cert.CertKeyPair, hostOverwrite *string, portOverwrite *int) error {
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
		{
			username: "kubernetes-admin",
			groups:   []string{"system:masters"},
			path:     "/etc/kubernetes/admin.conf",
		},
		{
			username: "system:kube-controller-manager",
			groups:   []string{},
			path:     "/etc/kubernetes/controller-manager.conf",
		},
		{
			username: "system:kube-proxy",
			groups:   []string{},
			path:     "/etc/kubernetes/proxy.conf",
		},
		{
			username: "system:kube-scheduler",
			groups:   []string{},
			path:     "/etc/kubernetes/scheduler.conf",
		},
		{
			username: fmt.Sprintf("system:node:%s", hostname),
			groups:   []string{"system:nodes"},
			path:     "/etc/kubernetes/kubelet.conf",
		},
	}

	for _, config := range configs {
		token, err := apiImpl.GetOrCreateAuthToken(ctx, state, config.username, config.groups)
		if err != nil {
			return fmt.Errorf("could not generate auth token for %s: %w", config.username, err)
		}

		err = renderKubeconfig(snap.SnapFromContext(state.Context), token, ca.CertPem, config.path, hostOverwrite, portOverwrite)
		if err != nil {
			return fmt.Errorf("failed to generate kubeconfig for %s: %w", config.username, err)
		}
	}

	return nil
}

// renderKubeconfig creates a kubeconfig file with the given token and CA data.
func renderKubeconfig(snap snap.Snap, token string, caCertPem []byte, path string, hostOverwrite *string, portOverwrite *int) error {
	port := apiServerPort(snap, portOverwrite)
	return utils.TemplateAndSave(snap.Path("k8s/config/kubeconfig-with-token.tmpl"),
		struct {
			CaData        string
			ApiServerIp   string
			ApiServerPort string
			Token         string
		}{
			CaData:        base64.StdEncoding.EncodeToString(caCertPem),
			ApiServerIp:   apiServerHost(hostOverwrite),
			ApiServerPort: port,
			Token:         token,
		},
		path,
	)
}

func apiServerHost(hostOverwrite *string) string {
	if hostOverwrite != nil {
		return *hostOverwrite
	}
	return "127.0.0.1"
}

func apiServerPort(s snap.Snap, portOverwrite *int) string {
	if portOverwrite != nil {
		return fmt.Sprintf("%d", *portOverwrite)
	} else {
		return snap.GetServiceArgument(s, "kube-apiserver", "--secure-port")
	}
}
