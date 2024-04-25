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

var kubeletControlPlaneLabels = []string{
	"node-role.kubernetes.io/control-plane=",
}

var kubeletWorkerLabels = []string{
	"node-role.kubernetes.io/worker=",
}

// KubeletControlPlane configures kubelet on a control plane node.
func KubeletControlPlane(snap snap.Snap, hostname string, nodeIP net.IP, clusterDNS string, clusterDomain string, cloudProvider string, registerWithTaints []string) error {
	return kubelet(snap, hostname, nodeIP, clusterDNS, clusterDomain, cloudProvider, registerWithTaints, append(kubeletControlPlaneLabels, kubeletWorkerLabels...))
}

// KubeletWorker configures kubelet on a worker node.
func KubeletWorker(snap snap.Snap, hostname string, nodeIP net.IP, clusterDNS string, clusterDomain string, cloudProvider string) error {
	return kubelet(snap, hostname, nodeIP, clusterDNS, clusterDomain, cloudProvider, nil, kubeletWorkerLabels)
}

// kubelet configures kubelet on the local node.
func kubelet(snap snap.Snap, hostname string, nodeIP net.IP, clusterDNS string, clusterDomain string, cloudProvider string, taints []string, labels []string) error {
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
		"--node-labels":                  strings.Join(labels, ","),
		"--read-only-port":               "0",
		"--register-with-taints":         strings.Join(taints, ","),
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
	if nodeIP != nil && !nodeIP.IsLoopback() {
		args["--node-ip"] = nodeIP.String()
	}
	if _, err := snaputil.UpdateServiceArguments(snap, "kubelet", args, nil); err != nil {
		return fmt.Errorf("failed to render arguments file: %w", err)
	}
	return nil
}
