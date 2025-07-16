package setup

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
)

type apiserverAuthTokenWebhookTemplateConfig struct {
	URL    string
	CAPath string
}

var SupportedDatastores = []string{"k8s-dqlite", "external"}

var (
	apiserverAuthTokenWebhookTemplate = mustTemplate("apiserver", "auth-token-webhook.conf")

	apiserverTLSCipherSuites = []string{
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
	}
)

// KubeAPIServer configures kube-apiserver on the local node.
func KubeAPIServer(snap snap.Snap, securePort int, nodeIP net.IP, serviceCIDR string, authWebhookURL string, enableFrontProxy bool, datastore types.Datastore, authorizationMode string, extraArgs map[string]*string) error {
	authTokenWebhookConfigFile := filepath.Join(snap.ServiceExtraConfigDir(), "auth-token-webhook.conf")
	authTokenWebhookFile, err := os.OpenFile(authTokenWebhookConfigFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open auth-token-webhook.conf: %w", err)
	}

	if err := apiserverAuthTokenWebhookTemplate.Execute(authTokenWebhookFile, apiserverAuthTokenWebhookTemplateConfig{
		URL:    authWebhookURL,
		CAPath: filepath.Join(snap.K8sdStateDir(), "cluster.crt"),
	}); err != nil {
		return fmt.Errorf("failed to write auth-token-webhook.conf: %w", err)
	}
	defer authTokenWebhookFile.Close()

	args := map[string]string{
		"--anonymous-auth":                           "false",
		"--allow-privileged":                         "true",
		"--authentication-token-webhook-config-file": authTokenWebhookConfigFile,
		"--authorization-mode":                       authorizationMode,
		"--client-ca-file":                           filepath.Join(snap.KubernetesPKIDir(), "client-ca.crt"),
		"--enable-admission-plugins":                 "NodeRestriction",
		"--kubelet-certificate-authority":            filepath.Join(snap.KubernetesPKIDir(), "ca.crt"),
		"--kubelet-client-certificate":               filepath.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.crt"),
		"--kubelet-client-key":                       filepath.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.key"),
		"--kubelet-preferred-address-types":          "InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP",
		"--profiling":                                "false",
		"--request-timeout":                          "300s",
		"--secure-port":                              strconv.Itoa(securePort),
		"--service-account-issuer":                   "https://kubernetes.default.svc",
		"--service-account-key-file":                 filepath.Join(snap.KubernetesPKIDir(), "serviceaccount.key"),
		"--service-account-signing-key-file":         filepath.Join(snap.KubernetesPKIDir(), "serviceaccount.key"),
		"--service-cluster-ip-range":                 serviceCIDR,
		"--tls-cert-file":                            filepath.Join(snap.KubernetesPKIDir(), "apiserver.crt"),
		"--tls-cipher-suites":                        strings.Join(apiserverTLSCipherSuites, ","),
		"--tls-min-version":                          "VersionTLS12",
		"--tls-private-key-file":                     filepath.Join(snap.KubernetesPKIDir(), "apiserver.key"),
	}

	if nodeIP != nil && !nodeIP.IsLoopback() {
		args["--advertise-address"] = nodeIP.String()
	}

	switch datastore.GetType() {
	case "k8s-dqlite", "etcd", "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", datastore.GetType(), SupportedDatastores)
	}

	datastoreUpdateArgs, deleteArgs, err := datastore.ToKubeAPIServerArguments(snap)
	if err != nil {
		return fmt.Errorf("failed to get datastore arguments for kube-apiserver: %w", err)
	}

	for key, val := range datastoreUpdateArgs {
		args[key] = val
	}

	if enableFrontProxy {
		args["--requestheader-client-ca-file"] = filepath.Join(snap.KubernetesPKIDir(), "front-proxy-ca.crt")
		args["--requestheader-allowed-names"] = "front-proxy-client"
		args["--requestheader-extra-headers-prefix"] = "X-Remote-Extra-"
		args["--requestheader-group-headers"] = "X-Remote-Group"
		args["--requestheader-username-headers"] = "X-Remote-User"
		args["--proxy-client-cert-file"] = filepath.Join(snap.KubernetesPKIDir(), "front-proxy-client.crt")
		args["--proxy-client-key-file"] = filepath.Join(snap.KubernetesPKIDir(), "front-proxy-client.key")
	}
	if _, err := snaputil.UpdateServiceArguments(snap, "kube-apiserver", args, deleteArgs); err != nil {
		return fmt.Errorf("failed to render arguments file: %w", err)
	}

	// Apply extra arguments after the defaults, so they can override them.
	updateArgs, deleteArgs := utils.ServiceArgsFromMap(extraArgs)
	if _, err := snaputil.UpdateServiceArguments(snap, "kube-apiserver", updateArgs, deleteArgs); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}
	return nil
}
