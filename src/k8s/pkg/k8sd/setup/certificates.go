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
// It will delete the file if contents parameter is an empty string. Trying to ensure a inexistent file
// with an empty contents parameter does not result in an error.
// It returns true if any of these is true: the file's content changed, it was created or it was deleted.
// It also returns any error that occured.
func ensureFile(fname string, contents string, uid, gid int, mode fs.FileMode) (bool, error) {
	if contents == "" {
		if err := os.Remove(fname); err != nil {
			if !os.IsNotExist(err) {
				// File exists but failed to delete.
				return false, fmt.Errorf("failed to delete: %w", err)
			}
			// File does not exist, nothing to do.
			return false, nil
		}

		// File was deleted.
		return true, nil
	}

	origContent, err := os.ReadFile(fname)
	if err != nil && !os.IsNotExist(err) {
		// File exists but failed to read.
		return false, fmt.Errorf("failed to read: %w", err)
	}

	var contentChanged bool

	if contents != string(origContent) {
		if err := os.WriteFile(fname, []byte(contents), mode); err != nil {
			return false, fmt.Errorf("failed to write: %w", err)
		}
		contentChanged = true
	}

	if err := os.Chown(fname, uid, gid); err != nil {
		return false, fmt.Errorf("failed to chown: %w", err)
	}
	if err := os.Chmod(fname, mode); err != nil {
		return false, fmt.Errorf("failed to chmod: %w", err)
	}

	return contentChanged, nil
}

// ensureFiles calls ensureFile for many files.
// It returns true if one or more files were updated and any error that occured.
func ensureFiles(uid, gid int, mode fs.FileMode, files map[string]string) (bool, error) {
	var changed bool
	for fname, content := range files {
		if v, err := ensureFile(fname, content, uid, gid, mode); err != nil {
			return false, fmt.Errorf("failed to configure %s: %w", path.Base(fname), err)
		} else if v {
			changed = true
		}
	}
	return changed, nil
}

// EnsureExtDatastorePKI ensures the external datastore PKI files are present
// and have the correct content, permissions and ownership.
// It returns true if one or more files were updated and any error that occured.
func EnsureExtDatastorePKI(snap snap.Snap, certificates *pki.ExternalDatastorePKI) (bool, error) {
	return ensureFiles(snap.UID(), snap.GID(), 0600, map[string]string{
		path.Join(snap.EtcdPKIDir(), "ca.crt"):     certificates.DatastoreCACert,
		path.Join(snap.EtcdPKIDir(), "client.key"): certificates.DatastoreClientKey,
		path.Join(snap.EtcdPKIDir(), "client.crt"): certificates.DatastoreClientCert,
	})
}

// EnsureK8sDqlitePKI ensures the k8s dqlite PKI files are present
// and have the correct content, permissions and ownership.
// It returns true if one or more files were updated and any error that occured.
func EnsureK8sDqlitePKI(snap snap.Snap, certificates *pki.K8sDqlitePKI) (bool, error) {
	return ensureFiles(snap.UID(), snap.GID(), 0600, map[string]string{
		path.Join(snap.K8sDqliteStateDir(), "cluster.crt"): certificates.K8sDqliteCert,
		path.Join(snap.K8sDqliteStateDir(), "cluster.key"): certificates.K8sDqliteKey,
	})
}

// EnsureControlPlanePKI ensures the control plane PKI files are present
// and have the correct content, permissions and ownership.
// It returns true if one or more files were updated and any error that occured.
func EnsureControlPlanePKI(snap snap.Snap, certificates *pki.ControlPlanePKI) (bool, error) {
	return ensureFiles(snap.UID(), snap.GID(), 0600, map[string]string{
		path.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.crt"): certificates.APIServerKubeletClientCert,
		path.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.key"): certificates.APIServerKubeletClientKey,
		path.Join(snap.KubernetesPKIDir(), "apiserver.crt"):                certificates.APIServerCert,
		path.Join(snap.KubernetesPKIDir(), "apiserver.key"):                certificates.APIServerKey,
		path.Join(snap.KubernetesPKIDir(), "ca.crt"):                       certificates.CACert,
		path.Join(snap.KubernetesPKIDir(), "client-ca.crt"):                certificates.ClientCACert,
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

// EnsureWorkerPKI ensures the worker PKI files are present
// and have the correct content, permissions and ownership.
// It returns true if one or more files were updated and any error that occured.
func EnsureWorkerPKI(snap snap.Snap, certificates *pki.WorkerNodePKI) (bool, error) {
	return ensureFiles(snap.UID(), snap.GID(), 0600, map[string]string{
		path.Join(snap.KubernetesPKIDir(), "ca.crt"):        certificates.CACert,
		path.Join(snap.KubernetesPKIDir(), "client-ca.crt"): certificates.ClientCACert,
		path.Join(snap.KubernetesPKIDir(), "kubelet.crt"):   certificates.KubeletCert,
		path.Join(snap.KubernetesPKIDir(), "kubelet.key"):   certificates.KubeletKey,
	})
}

// EnsureEtcdPKI ensures the etcd PKI files are present.
// and have the correct content, permissions and ownership.
// It returns true if one or more files were updated and any error that occured.
func EnsureEtcdPKI(snap snap.Snap, certificates *pki.EtcdPKI) (bool, error) {
	return ensureFiles(snap.UID(), snap.GID(), 0600, map[string]string{
		path.Join(snap.EtcdPKIDir(), "ca.crt"):                          certificates.CACert,
		path.Join(snap.EtcdPKIDir(), "server.crt"):                      certificates.ServerCert,
		path.Join(snap.EtcdPKIDir(), "server.key"):                      certificates.ServerKey,
		path.Join(snap.EtcdPKIDir(), "peer.crt"):                        certificates.ServerPeerCert,
		path.Join(snap.EtcdPKIDir(), "peer.key"):                        certificates.ServerPeerKey,
		path.Join(snap.KubernetesPKIDir(), "apiserver-etcd-client.crt"): certificates.APIServerClientCert,
		path.Join(snap.KubernetesPKIDir(), "apiserver-etcd-client.key"): certificates.APIServerClientKey,
	})
}
