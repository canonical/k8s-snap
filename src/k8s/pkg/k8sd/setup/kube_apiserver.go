package setup

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
)

type apiserverAuthTokenWebhookTemplateConfig struct {
	URL    string
	CAPath string
}

var SupportedDatastores = []string{"k8s-dqlite", "external"}

var (
	apiserverAuthTokenWebhookTemplate = mustTemplate("apiserver", "auth-token-webhook.conf")

	apiserverTLSCipherSuites = []string{
		"TLS_AES_128_GCM_SHA256",
		"TLS_AES_256_GCM_SHA384",
		"TLS_CHACHA20_POLY1305_SHA256",
		"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
		"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
		"TLS_RSA_WITH_3DES_EDE_CBC_SHA",
		"TLS_RSA_WITH_AES_128_CBC_SHA",
		"TLS_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_RSA_WITH_AES_256_CBC_SHA",
		"TLS_RSA_WITH_AES_256_GCM_SHA384",
	}
)

// KubeAPIServer configures kube-apiserver on the local node.
func KubeAPIServer(snap snap.Snap, serviceCIDR string, authWebhookURL string, enableFrontProxy bool, datastore types.Datastore, authorizationMode string) error {
	authTokenWebhookConfigFile := path.Join(snap.ServiceExtraConfigDir(), "auth-token-webhook.conf")
	authTokenWebhookFile, err := os.OpenFile(authTokenWebhookConfigFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open auth-token-webhook.conf: %w", err)
	}

	if err := apiserverAuthTokenWebhookTemplate.Execute(authTokenWebhookFile, apiserverAuthTokenWebhookTemplateConfig{
		URL:    authWebhookURL,
		CAPath: path.Join(snap.K8sdStateDir(), "cluster.crt"),
	}); err != nil {
		return fmt.Errorf("failed to write auth-token-webhook.conf: %w", err)
	}
	defer authTokenWebhookFile.Close()

	args := map[string]string{
		"--allow-privileged":                         "true",
		"--authentication-token-webhook-config-file": authTokenWebhookConfigFile,
		"--authorization-mode":                       authorizationMode,
		"--client-ca-file":                           path.Join(snap.KubernetesPKIDir(), "client-ca.crt"),
		"--enable-admission-plugins":                 "NodeRestriction",
		"--kubelet-certificate-authority":            path.Join(snap.KubernetesPKIDir(), "ca.crt"),
		"--kubelet-client-certificate":               path.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.crt"),
		"--kubelet-client-key":                       path.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.key"),
		"--kubelet-preferred-address-types":          "InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP",
		"--secure-port":                              "6443",
		"--service-account-issuer":                   "https://kubernetes.default.svc",
		"--service-account-key-file":                 path.Join(snap.KubernetesPKIDir(), "serviceaccount.key"),
		"--service-account-signing-key-file":         path.Join(snap.KubernetesPKIDir(), "serviceaccount.key"),
		"--service-cluster-ip-range":                 serviceCIDR,
		"--tls-cert-file":                            path.Join(snap.KubernetesPKIDir(), "apiserver.crt"),
		"--tls-cipher-suites":                        strings.Join(apiserverTLSCipherSuites, ","),
		"--tls-private-key-file":                     path.Join(snap.KubernetesPKIDir(), "apiserver.key"),
		"--anonymous-auth":                           "false",
		"--profiling":                                "false",
		"--requests-timeout":                         "300s",
	}

	switch datastore.GetType() {
	case "k8s-dqlite", "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", datastore.GetType(), SupportedDatastores)
	}

	datastoreUpdateArgs, deleteArgs := datastore.ToKubeAPIServerArguments(snap)
	for key, val := range datastoreUpdateArgs {
		args[key] = val
	}

	if enableFrontProxy {
		args["--requestheader-client-ca-file"] = path.Join(snap.KubernetesPKIDir(), "front-proxy-ca.crt")
		args["--requestheader-allowed-names"] = "front-proxy-client"
		args["--requestheader-extra-headers-prefix"] = "X-Remote-Extra-"
		args["--requestheader-group-headers"] = "X-Remote-Group"
		args["--requestheader-username-headers"] = "X-Remote-User"
		args["--proxy-client-cert-file"] = path.Join(snap.KubernetesPKIDir(), "front-proxy-client.crt")
		args["--proxy-client-key-file"] = path.Join(snap.KubernetesPKIDir(), "front-proxy-client.key")
	}
	if _, err := snaputil.UpdateServiceArguments(snap, "kube-apiserver", args, deleteArgs); err != nil {
		return fmt.Errorf("failed to render arguments file: %w", err)
	}
	return nil
}
