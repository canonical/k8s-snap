package setup

import (
	"fmt"
	"os"
	"path"

	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
)

// KubeControllerManager configures kube-controller-manager on the local node.
func KubeControllerManager(snap snap.Snap) error {
	args := map[string]string{
		"--kubeconfig":                       path.Join(snap.KubernetesConfigDir(), "controller.conf"),
		"--authorization-kubeconfig":         path.Join(snap.KubernetesConfigDir(), "controller.conf"),
		"--authentication-kubeconfig":        path.Join(snap.KubernetesConfigDir(), "controller.conf"),
		"--service-account-private-key-file": path.Join(snap.KubernetesPKIDir(), "serviceaccount.key"),
		"--root-ca-file":                     path.Join(snap.KubernetesPKIDir(), "ca.crt"),
		"--use-service-account-credentials":  "true",
		"--profiling":                        "false",
		"--leader-elect-lease-duration":      "30s",
		"--leader-elect-renew-deadline":      "15s",
	}
	// enable cluster-signing if certificates are available
	if _, err := os.Stat(path.Join(snap.KubernetesPKIDir(), "ca.key")); err == nil {
		args["--cluster-signing-cert-file"] = path.Join(snap.KubernetesPKIDir(), "ca.crt")
		args["--cluster-signing-key-file"] = path.Join(snap.KubernetesPKIDir(), "ca.key")
	}
	if _, err := snaputil.UpdateServiceArguments(snap, "kube-controller-manager", args, nil); err != nil {
		return fmt.Errorf("failed to render arguments file: %w", err)
	}
	return nil
}
