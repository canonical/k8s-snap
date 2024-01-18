package setup

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	apiImpl "github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/snap"
	snapPkg "github.com/canonical/k8s/pkg/snap"
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

		err = renderKubeconfig(snapPkg.SnapFromContext(state.Context), token, ca.CertPem, config.path, hostOverwrite, portOverwrite)
		if err != nil {
			return fmt.Errorf("failed to generate kubeconfig for %s: %w", config.username, err)
		}
	}

	return nil
}

// renderX509Kubeconfig creates a kubeconfig file with the given x509 certificate and key data.
func renderX509Kubeconfig(snap snapPkg.Snap, keyPem, certPem, caCertPem []byte, path string, hostOverwrite *string, portOverwrite *int) error {
	port, err := apiServerPort(snap, portOverwrite)
	if err != nil {
		return fmt.Errorf("failed to render kubeconfig: %w", err)
	}

	return utils.TemplateAndSave(snap.Path("k8s/config/kubeconfig-with-x509.tmpl"),
		struct {
			CaData        string
			ApiServerIp   string
			ApiServerPort int
			CertData      string
			KeyData       string
		}{
			CaData:        base64.StdEncoding.EncodeToString(caCertPem),
			ApiServerIp:   apiServerHost(hostOverwrite),
			ApiServerPort: port,
			CertData:      base64.StdEncoding.EncodeToString(certPem),
			KeyData:       base64.StdEncoding.EncodeToString(keyPem),
		},
		path,
	)
}

// renderKubeconfig creates a kubeconfig file with the given token and CA data.
func renderKubeconfig(snap snap.Snap, token string, caCertPem []byte, path string, hostOverwrite *string, portOverwrite *int) error {
	port, err := apiServerPort(snap, portOverwrite)
	if err != nil {
		return fmt.Errorf("failed to render kubeconfig: %w", err)
	}
	return utils.TemplateAndSave(snap.Path("k8s/config/kubeconfig-with-token.tmpl"),
		struct {
			CaData        string
			ApiServerIp   string
			ApiServerPort int
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

func apiServerPort(snap snap.Snap, portOverwrite *int) (port int, err error) {
	if portOverwrite != nil {
		port = *portOverwrite
	} else {
		port, err = strconv.Atoi(snapPkg.GetServiceArgument(
			snap,
			"kube-apiserver",
			"--secure-port",
		))
		if err != nil {
			return 0, fmt.Errorf("apiserver port is not an integer: %w", err)
		}
	}

	return port, err
}
