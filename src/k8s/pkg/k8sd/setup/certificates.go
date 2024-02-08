package setup

import (
	"fmt"
	"os"
	"path"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/snap"
)

func EnsureControlPlanePKI(snap snap.Snap, certificates *pki.ControlPlanePKI) error {
	toWrite := map[string]string{
		path.Join(snap.KubernetesPKIDir(), "ca.crt"):                       certificates.CACert,
		path.Join(snap.KubernetesPKIDir(), "front-proxy-ca.crt"):           certificates.FrontProxyCACert,
		path.Join(snap.KubernetesPKIDir(), "front-proxy-client.crt"):       certificates.FrontProxyClientCert,
		path.Join(snap.KubernetesPKIDir(), "front-proxy-client.key"):       certificates.FrontProxyClientKey,
		path.Join(snap.KubernetesPKIDir(), "apiserver.crt"):                certificates.APIServerCert,
		path.Join(snap.KubernetesPKIDir(), "apiserver.key"):                certificates.APIServerKey,
		path.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.crt"): certificates.APIServerKubeletClientCert,
		path.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.key"): certificates.APIServerKubeletClientKey,
		path.Join(snap.KubernetesPKIDir(), "kubelet.crt"):                  certificates.KubeletCert,
		path.Join(snap.KubernetesPKIDir(), "kubelet.key"):                  certificates.KubeletKey,
		path.Join(snap.KubernetesPKIDir(), "serviceaccount.key"):           certificates.ServiceAccountKey,
		path.Join(snap.K8sDqliteStateDir(), "cluster.crt"):                 certificates.K8sDqliteCert,
		path.Join(snap.K8sDqliteStateDir(), "cluster.key"):                 certificates.K8sDqliteKey,
	}

	if certificates.CAKey != "" {
		toWrite[path.Join(snap.KubernetesPKIDir(), "ca.key")] = certificates.CAKey
	}
	if certificates.FrontProxyCAKey != "" {
		toWrite[path.Join(snap.KubernetesPKIDir(), "front-proxy-ca.key")] = certificates.FrontProxyCACert
	}

	for fname, cert := range toWrite {
		if err := os.WriteFile(fname, []byte(cert), 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", path.Base(fname), err)
		}
		if err := os.Chown(fname, snap.UID(), snap.GID()); err != nil {
			return fmt.Errorf("failed to chown %s: %w", fname, err)
		}
		if err := os.Chmod(fname, 0600); err != nil {
			return fmt.Errorf("failed to chmod %s: %w", fname, err)
		}
	}

	return nil
}

func EnsureWorkerPKI(snap snap.Snap, certificates *pki.WorkerNodePKI) error {
	toWrite := map[string]string{
		path.Join(snap.KubernetesPKIDir(), "ca.crt"): certificates.CACert,
	}

	if certificates.KubeletCert != "" {
		toWrite[path.Join(snap.KubernetesPKIDir(), "kubelet.crt")] = certificates.KubeletCert
	}
	if certificates.KubeletKey != "" {
		toWrite[path.Join(snap.KubernetesPKIDir(), "kubelet.crt")] = certificates.KubeletKey
	}

	for fname, cert := range toWrite {
		if err := os.WriteFile(fname, []byte(cert), 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", path.Base(fname), err)
		}
		if err := os.Chown(fname, snap.UID(), snap.GID()); err != nil {
			return fmt.Errorf("failed to chown %s: %w", fname, err)
		}
		if err := os.Chmod(fname, 0600); err != nil {
			return fmt.Errorf("failed to chmod %s: %w", fname, err)
		}
	}

	return nil
}
