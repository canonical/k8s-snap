package setup

import (
	"fmt"
	"net"
	"path"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
)

var kubeletTLSCipherSuites = []string{
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
	"TLS_RSA_WITH_AES_128_GCM_SHA256",
	"TLS_RSA_WITH_AES_256_GCM_SHA384",
}

// Kubelet configures kubelet on the local node.
func Kubelet(snap snap.Snap, hostname string, nodeIP net.IP, clusterDNS string, clusterDomain string, cloudProvider string) error {
	args := map[string]string{
		"--anonymous-auth":               "false",
		"--authentication-token-webhook": "true",
		"--cert-dir":                     snap.KubernetesPKIDir(),
		"--client-ca-file":               path.Join(snap.KubernetesPKIDir(), "ca.crt"),
		"--container-runtime-endpoint":   path.Join(snap.ContainerdSocketDir(), "containerd.sock"),
		"--containerd":                   path.Join(snap.ContainerdSocketDir(), "containerd.sock"),
		"--eviction-hard":                "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'",
		"--fail-swap-on":                 "false",
		"--hostname-override":            hostname,
		"--kubeconfig":                   path.Join(snap.KubernetesConfigDir(), "kubelet.conf"),
		"--read-only-port":               "0",
		"--root-dir":                     snap.KubeletRootDir(),
		"--serialize-image-pulls":        "false",
		"--tls-cipher-suites":            strings.Join(kubeletTLSCipherSuites, ","),
	}
	if cloudProvider != "" {
		args["--cloud-provider"] = cloudProvider
	}
	if clusterDNS != "" {
		args["--cluster-dns"] = clusterDNS
	}
	if clusterDomain != "" {
		args["--cluster-domain"] = clusterDomain
	}
	if nodeIP != nil {
		args["--node-ip"] = nodeIP.String()
	}
	if _, err := snaputil.UpdateServiceArguments(snap, "kubelet", args, nil); err != nil {
		return fmt.Errorf("failed to render arguments file: %w", err)
	}
	return nil
}
