package setup

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"k8s.io/client-go/tools/clientcmd"
)

// ensureFile creates fname with the specified contents, mode and owner bits.
// It will delete the file if contents parameter is an empty string. Trying to ensure a inexistent file
// with an empty contents parameter does not result in an error.
// It returns true if any of these is true: the file's content changed, it was created or it was deleted.
// It also returns any error that occurred.
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
		if err := utils.WriteFile(fname, []byte(contents), mode); err != nil {
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
// It returns true if one or more files were updated and any error that occurred.
func ensureFiles(uid, gid int, mode fs.FileMode, files map[string]string) (bool, error) {
	var changed bool
	for fname, content := range files {
		if v, err := ensureFile(fname, content, uid, gid, mode); err != nil {
			return false, fmt.Errorf("failed to configure %s: %w", filepath.Base(fname), err)
		} else if v {
			changed = true
		}
	}
	return changed, nil
}

// EnsureExtDatastorePKI ensures the external datastore PKI files are present
// and have the correct content, permissions and ownership.
// It returns true if one or more files were updated and any error that occurred.
func EnsureExtDatastorePKI(snap snap.Snap, certificates *pki.ExternalDatastorePKI) (bool, error) {
	return ensureFiles(snap.UID(), snap.GID(), 0o600, map[string]string{
		filepath.Join(snap.EtcdPKIDir(), "ca.crt"):     certificates.DatastoreCACert,
		filepath.Join(snap.EtcdPKIDir(), "client.key"): certificates.DatastoreClientKey,
		filepath.Join(snap.EtcdPKIDir(), "client.crt"): certificates.DatastoreClientCert,
	})
}

// EnsureK8sDqlitePKI ensures the k8s dqlite PKI files are present
// and have the correct content, permissions and ownership.
// It returns true if one or more files were updated and any error that occurred.
func EnsureK8sDqlitePKI(snap snap.Snap, certificates *pki.K8sDqlitePKI) (bool, error) {
	return ensureFiles(snap.UID(), snap.GID(), 0o600, map[string]string{
		filepath.Join(snap.K8sDqliteStateDir(), "cluster.crt"): certificates.K8sDqliteCert,
		filepath.Join(snap.K8sDqliteStateDir(), "cluster.key"): certificates.K8sDqliteKey,
	})
}

// EnsureControlPlanePKI ensures the control plane PKI files are present
// and have the correct content, permissions and ownership.
// It returns true if one or more files were updated and any error that occurred.
func EnsureControlPlanePKI(snap snap.Snap, certificates *pki.ControlPlanePKI) (bool, error) {
	return ensureFiles(snap.UID(), snap.GID(), 0o600, map[string]string{
		filepath.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.crt"): certificates.APIServerKubeletClientCert,
		filepath.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.key"): certificates.APIServerKubeletClientKey,
		filepath.Join(snap.KubernetesPKIDir(), "apiserver.crt"):                certificates.APIServerCert,
		filepath.Join(snap.KubernetesPKIDir(), "apiserver.key"):                certificates.APIServerKey,
		filepath.Join(snap.KubernetesPKIDir(), "ca.crt"):                       certificates.CACert,
		filepath.Join(snap.KubernetesPKIDir(), "ca.key"):                       certificates.CAKey,
		filepath.Join(snap.KubernetesPKIDir(), "client-ca.crt"):                certificates.ClientCACert,
		filepath.Join(snap.KubernetesPKIDir(), "client-ca.key"):                certificates.ClientCAKey,
		filepath.Join(snap.KubernetesPKIDir(), "front-proxy-ca.crt"):           certificates.FrontProxyCACert,
		filepath.Join(snap.KubernetesPKIDir(), "front-proxy-ca.key"):           certificates.FrontProxyCAKey,
		filepath.Join(snap.KubernetesPKIDir(), "front-proxy-client.crt"):       certificates.FrontProxyClientCert,
		filepath.Join(snap.KubernetesPKIDir(), "front-proxy-client.key"):       certificates.FrontProxyClientKey,
		filepath.Join(snap.KubernetesPKIDir(), "kubelet.crt"):                  certificates.KubeletCert,
		filepath.Join(snap.KubernetesPKIDir(), "kubelet.key"):                  certificates.KubeletKey,
		filepath.Join(snap.KubernetesPKIDir(), "serviceaccount.key"):           certificates.ServiceAccountKey,
	})
}

// EnsureWorkerPKI ensures the worker PKI files are present
// and have the correct content, permissions and ownership.
// It returns true if one or more files were updated and any error that occurred.
func EnsureWorkerPKI(snap snap.Snap, certificates *pki.WorkerNodePKI) (bool, error) {
	return ensureFiles(snap.UID(), snap.GID(), 0o600, map[string]string{
		filepath.Join(snap.KubernetesPKIDir(), "ca.crt"):        certificates.CACert,
		filepath.Join(snap.KubernetesPKIDir(), "client-ca.crt"): certificates.ClientCACert,
		filepath.Join(snap.KubernetesPKIDir(), "kubelet.crt"):   certificates.KubeletCert,
		filepath.Join(snap.KubernetesPKIDir(), "kubelet.key"):   certificates.KubeletKey,
	})
}

// ReadControlPlanePKI reads the existing control plane PKI files and kubeconfig files,
// populating a ControlPlanePKI structure with their contents.
// The readManaged parameter controls which certificates to read:
// - If readManaged=true: only reads certificates where CA keys are present (internally managed)
// - If readManaged=false: only reads certificates where CA keys are missing (externally managed).
func ReadControlPlanePKI(snap snap.Snap, certificates *pki.ControlPlanePKI, readManaged bool) error {
	caKeyPath := filepath.Join(snap.KubernetesPKIDir(), "ca.key")
	caKeyExists, err := utils.FileExists(caKeyPath)
	if err != nil {
		return fmt.Errorf("failed to check if ca.key exists: %w", err)
	}

	clientCAKeyPath := filepath.Join(snap.KubernetesPKIDir(), "client-ca.key")
	clientCAKeyExists, err := utils.FileExists(clientCAKeyPath)
	if err != nil {
		return fmt.Errorf("failed to check if client-ca.key exists: %w", err)
	}

	frontProxyCAKeyPath := filepath.Join(snap.KubernetesPKIDir(), "front-proxy-ca.key")
	frontProxyCAKeyExists, err := utils.FileExists(frontProxyCAKeyPath)
	if err != nil {
		return fmt.Errorf("failed to check if front-proxy-ca.key: %w", err)
	}

	fileFields := map[string]struct {
		field *string
		read  bool
	}{
		// CA certificates
		filepath.Join(snap.KubernetesPKIDir(), "ca.crt"): {
			field: &certificates.CACert,
			read:  true,
		},
		filepath.Join(snap.KubernetesPKIDir(), "client-ca.crt"): {
			field: &certificates.ClientCACert,
			read:  true,
		},
		filepath.Join(snap.KubernetesPKIDir(), "front-proxy-ca.crt"): {
			field: &certificates.FrontProxyCACert,
			read:  true,
		},

		// CA keys
		filepath.Join(snap.KubernetesPKIDir(), "ca.key"): {
			field: &certificates.CAKey,
			read:  readManaged && caKeyExists,
		},
		filepath.Join(snap.KubernetesPKIDir(), "client-ca.key"): {
			field: &certificates.ClientCAKey,
			read:  readManaged && clientCAKeyExists,
		},
		filepath.Join(snap.KubernetesPKIDir(), "front-proxy-ca.key"): {
			field: &certificates.FrontProxyCAKey,
			read:  readManaged && frontProxyCAKeyExists,
		},

		// Certificates
		// NOTE: read is using the XNOR operation.
		filepath.Join(snap.KubernetesPKIDir(), "apiserver.crt"): {
			field: &certificates.APIServerCert,
			read:  readManaged == caKeyExists,
		},
		filepath.Join(snap.KubernetesPKIDir(), "apiserver.key"): {
			field: &certificates.APIServerKey,
			read:  readManaged == caKeyExists,
		},

		filepath.Join(snap.KubernetesPKIDir(), "kubelet.crt"): {
			field: &certificates.KubeletCert,
			read:  readManaged == caKeyExists,
		},
		filepath.Join(snap.KubernetesPKIDir(), "kubelet.key"): {
			field: &certificates.KubeletKey,
			read:  readManaged == caKeyExists,
		},

		filepath.Join(snap.KubernetesPKIDir(), "front-proxy-client.crt"): {
			field: &certificates.FrontProxyClientCert,
			read:  readManaged == frontProxyCAKeyExists,
		},
		filepath.Join(snap.KubernetesPKIDir(), "front-proxy-client.key"): {
			field: &certificates.FrontProxyClientKey,
			read:  readManaged == frontProxyCAKeyExists,
		},

		filepath.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.crt"): {
			field: &certificates.APIServerKubeletClientCert,
			read:  readManaged == clientCAKeyExists,
		},
		filepath.Join(snap.KubernetesPKIDir(), "apiserver-kubelet-client.key"): {
			field: &certificates.APIServerKubeletClientKey,
			read:  readManaged == clientCAKeyExists,
		},
		filepath.Join(snap.KubernetesPKIDir(), "serviceaccount.key"): {
			field: &certificates.ServiceAccountKey,
			read:  true,
		},
	}

	for filePath, info := range fileFields {
		if !info.read {
			// NOTE: Skip files that are not marked for reading
			// based on the presence of required keys and the
			// readManaged flag. This ensures only relevant PKI
			// files are loaded.
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
		if info.field != nil {
			*info.field = string(content)
		}
	}

	kubeconfigFields := map[string]struct {
		certField *string
		keyField  *string
		read      bool
	}{
		"admin.conf": {
			certField: &certificates.AdminClientCert,
			keyField:  &certificates.AdminClientKey,
			read:      readManaged == clientCAKeyExists,
		},
		"controller.conf": {
			certField: &certificates.KubeControllerManagerClientCert,
			keyField:  &certificates.KubeControllerManagerClientKey,
			read:      readManaged == clientCAKeyExists,
		},
		"scheduler.conf": {
			certField: &certificates.KubeSchedulerClientCert,
			keyField:  &certificates.KubeSchedulerClientKey,
			read:      readManaged == clientCAKeyExists,
		},
		"proxy.conf": {
			certField: &certificates.KubeProxyClientCert,
			keyField:  &certificates.KubeProxyClientKey,
			read:      readManaged == clientCAKeyExists,
		},
		"kubelet.conf": {
			certField: &certificates.KubeletClientCert,
			keyField:  &certificates.KubeletClientKey,
			read:      readManaged == clientCAKeyExists,
		},
	}

	for configName, info := range kubeconfigFields {
		if !info.read {
			// NOTE: Skip kubeconfig files that are not marked for
			// reading based on the presence of the client-ca key
			// and the readManaged flag.
			continue
		}

		configPath := filepath.Join(snap.KubernetesConfigDir(), configName)

		kubeConfig, err := clientcmd.LoadFromFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to load kubeconfig %s: %w", configName, err)
		}

		authInfo, exists := kubeConfig.AuthInfos["k8s-user"]
		if !exists {
			return fmt.Errorf("user 'k8s-user' not found in kubeconfig %s", configName)
		}

		if authInfo.ClientCertificateData != nil && info.certField != nil {
			*info.certField = string(authInfo.ClientCertificateData)
		}

		if authInfo.ClientKeyData != nil && info.keyField != nil {
			*info.keyField = string(authInfo.ClientKeyData)
		}
	}

	return nil
}

// EnsureEtcdPKI ensures the etcd PKI files are present.
// and have the correct content, permissions and ownership.
// It returns true if one or more files were updated and any error that occured.
func EnsureEtcdPKI(snap snap.Snap, certificates *pki.EtcdPKI) (bool, error) {
	return ensureFiles(snap.UID(), snap.GID(), 0o600, map[string]string{
		filepath.Join(snap.EtcdPKIDir(), "ca.crt"):                          certificates.CACert,
		filepath.Join(snap.EtcdPKIDir(), "server.crt"):                      certificates.ServerCert,
		filepath.Join(snap.EtcdPKIDir(), "server.key"):                      certificates.ServerKey,
		filepath.Join(snap.EtcdPKIDir(), "peer.crt"):                        certificates.ServerPeerCert,
		filepath.Join(snap.EtcdPKIDir(), "peer.key"):                        certificates.ServerPeerKey,
		filepath.Join(snap.KubernetesPKIDir(), "apiserver-etcd-client.crt"): certificates.APIServerClientCert,
		filepath.Join(snap.KubernetesPKIDir(), "apiserver-etcd-client.key"): certificates.APIServerClientKey,
	})
}
