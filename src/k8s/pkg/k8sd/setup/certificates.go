package setup

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/snap"
)

// ensureFile creates fname with the specified contents, mode and owner bits.
// ensureFile will delete the file if contents is an empty string.
func ensureFile(fname string, contents string, uid, gid int, mode fs.FileMode) error {
	if contents == "" {
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete: %w", err)
		}
		return nil
	}

	if err := os.WriteFile(fname, []byte(contents), mode); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}
	if err := os.Chown(fname, uid, gid); err != nil {
		return fmt.Errorf("failed to chown: %w", err)
	}
	if err := os.Chmod(fname, mode); err != nil {
		return fmt.Errorf("failed to chmod: %w", err)
	}

	return nil
}

// ensureFiles calls ensureFile for many files
func ensureFiles(uid, gid int, mode fs.FileMode, files map[string]string) error {
	for fname, cert := range files {
		if err := ensureFile(fname, cert, uid, gid, mode); err != nil {
			return fmt.Errorf("failed to configure %s: %w", path.Base(fname), err)
		}
	}
	return nil
}

func EnsureExtDatastorePKI(snap snap.Snap, certificates *pki.ExternalDatastorePKI) error {
	return ensureFiles(snap.UID(), snap.GID(), 0600, map[string]string{
		path.Join(snap.EtcdPKIDir(), "ca.crt"):     certificates.DatastoreCACert,
		path.Join(snap.EtcdPKIDir(), "client.key"): certificates.DatastoreClientKey,
		path.Join(snap.EtcdPKIDir(), "client.crt"): certificates.DatastoreClientCert,
	})
}

func EnsureK8sDqlitePKI(snap snap.Snap, certificates *pki.K8sDqlitePKI) error {
	return ensureFiles(snap.UID(), snap.GID(), 0600, map[string]string{
		path.Join(snap.K8sDqliteStateDir(), "cluster.crt"): certificates.K8sDqliteCert,
		path.Join(snap.K8sDqliteStateDir(), "cluster.key"): certificates.K8sDqliteKey,
	})
}

func EnsureControlPlanePKI(snap snap.Snap, certificates *pki.ControlPlanePKI) error {
	return ensureFiles(snap.UID(), snap.GID(), 0600, map[string]string{
		path.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.crt"): certificates.APIServerKubeletClientCert,
		path.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.key"): certificates.APIServerKubeletClientKey,
		path.Join(snap.KubernetesPKIDir(), "apiserver.crt"):                certificates.APIServerCert,
		path.Join(snap.KubernetesPKIDir(), "apiserver.key"):                certificates.APIServerKey,
		path.Join(snap.KubernetesPKIDir(), "ca.crt"):                       certificates.CACert,
		path.Join(snap.KubernetesPKIDir(), "ca.key"):                       certificates.CAKey,
		path.Join(snap.KubernetesPKIDir(), "front-proxy-ca.crt"):           certificates.FrontProxyCACert,
		path.Join(snap.KubernetesPKIDir(), "front-proxy-ca.key"):           certificates.FrontProxyCAKey,
		path.Join(snap.KubernetesPKIDir(), "front-proxy-client.crt"):       certificates.FrontProxyClientCert,
		path.Join(snap.KubernetesPKIDir(), "front-proxy-client.key"):       certificates.FrontProxyClientKey,
		path.Join(snap.KubernetesPKIDir(), "kubelet.crt"):                  certificates.KubeletCert,
		path.Join(snap.KubernetesPKIDir(), "kubelet.key"):                  certificates.KubeletKey,
		path.Join(snap.KubernetesPKIDir(), "serviceaccount.key"):           certificates.ServiceAccountKey,
	})
}

func EnsureWorkerPKI(snap snap.Snap, certificates *pki.WorkerNodePKI) error {
	return ensureFiles(snap.UID(), snap.GID(), 0600, map[string]string{
		path.Join(snap.KubernetesPKIDir(), "ca.crt"):      certificates.CACert,
		path.Join(snap.KubernetesPKIDir(), "kubelet.crt"): certificates.KubeletCert,
		path.Join(snap.KubernetesPKIDir(), "kubelet.key"): certificates.KubeletKey,
	})
}
