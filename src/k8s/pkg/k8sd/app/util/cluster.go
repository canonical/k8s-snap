package apputil

import (
	"fmt"
	"path/filepath"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
)

func SetupControlPlaneKubeconfigs(kubeConfigDir string, securePort int, pki pki.ControlPlanePKI) error {
	for _, kubeconfig := range []struct {
		file string
		crt  string
		key  string
	}{
		{file: "admin.conf", crt: pki.AdminClientCert, key: pki.AdminClientKey},
		{file: "controller.conf", crt: pki.KubeControllerManagerClientCert, key: pki.KubeControllerManagerClientKey},
		{file: "proxy.conf", crt: pki.KubeProxyClientCert, key: pki.KubeProxyClientKey},
		{file: "scheduler.conf", crt: pki.KubeSchedulerClientCert, key: pki.KubeSchedulerClientKey},
		{file: "kubelet.conf", crt: pki.KubeletClientCert, key: pki.KubeletClientKey},
	} {
		if err := setup.Kubeconfig(filepath.Join(kubeConfigDir, kubeconfig.file), fmt.Sprintf("127.0.0.1:%d", securePort), pki.CACert, kubeconfig.crt, kubeconfig.key); err != nil {
			return fmt.Errorf("failed to write kubeconfig %s: %w", kubeconfig.file, err)
		}
	}
	return nil

}
