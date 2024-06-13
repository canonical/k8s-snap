package setup

import (
	"fmt"
	"os"
	"path"

	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
)

// KubeControllerManager configures kube-controller-manager on the local node.
func KubeControllerManager(snap snap.Snap, extraArgs map[string]*string) error {
	args := map[string]string{
		"--authentication-kubeconfig":        path.Join(snap.KubernetesConfigDir(), "controller.conf"),
		"--authorization-kubeconfig":         path.Join(snap.KubernetesConfigDir(), "controller.conf"),
		"--kubeconfig":                       path.Join(snap.KubernetesConfigDir(), "controller.conf"),
		"--leader-elect-lease-duration":      "30s",
		"--leader-elect-renew-deadline":      "15s",
		"--profiling":                        "false",
		"--root-ca-file":                     path.Join(snap.KubernetesPKIDir(), "ca.crt"),
		"--service-account-private-key-file": path.Join(snap.KubernetesPKIDir(), "serviceaccount.key"),
		"--use-service-account-credentials":  "true",
		"--terminated-pod-gc-threshold":      "12500",
	}
	// enable cluster-signing if certificates are available
	if _, err := os.Stat(path.Join(snap.KubernetesPKIDir(), "ca.key")); err == nil {
		args["--cluster-signing-cert-file"] = path.Join(snap.KubernetesPKIDir(), "ca.crt")
		args["--cluster-signing-key-file"] = path.Join(snap.KubernetesPKIDir(), "ca.key")
	}
	if _, err := snaputil.UpdateServiceArguments(snap, "kube-controller-manager", args, nil); err != nil {
		return fmt.Errorf("failed to render arguments file: %w", err)
	}
	// Apply extra arguments after the defaults, so they can override them.
	updateArgs, deleteArgs := utils.ServiceArgsFromMap(extraArgs)
	if _, err := snaputil.UpdateServiceArguments(snap, "kube-controller-manager", updateArgs, deleteArgs); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}
	return nil
}
