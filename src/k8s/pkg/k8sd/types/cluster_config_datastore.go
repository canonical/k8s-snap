package types

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Datastore struct {
	Type *string `json:"type,omitempty"`

	K8sDqlitePort *int    `json:"k8s-dqlite-port,omitempty"`
	K8sDqliteCert *string `json:"k8s-dqlite-crt,omitempty"`
	K8sDqliteKey  *string `json:"k8s-dqlite-key,omitempty"`

	ExternalServers    *[]string `json:"external-servers,omitempty"`
	ExternalCACert     *string   `json:"external-ca-crt,omitempty"`
	ExternalClientCert *string   `json:"external-client-crt,omitempty"`
	ExternalClientKey  *string   `json:"external-client-key,omitempty"`
}

func (c Datastore) GetType() string               { return getField(c.Type) }
func (c Datastore) GetK8sDqlitePort() int         { return getField(c.K8sDqlitePort) }
func (c Datastore) GetK8sDqliteCert() string      { return getField(c.K8sDqliteCert) }
func (c Datastore) GetK8sDqliteKey() string       { return getField(c.K8sDqliteKey) }
func (c Datastore) GetExternalServers() []string  { return getField(c.ExternalServers) }
func (c Datastore) GetExternalCACert() string     { return getField(c.ExternalCACert) }
func (c Datastore) GetExternalClientCert() string { return getField(c.ExternalClientCert) }
func (c Datastore) GetExternalClientKey() string  { return getField(c.ExternalClientKey) }
func (c Datastore) Empty() bool                   { return c == Datastore{} }

// DatastorePathsProvider is to avoid circular dependency for snap.Snap in Datastore.ToKubeAPIServerArguments().
type DatastorePathsProvider interface {
	K8sDqliteStateDir() string
	EtcdPKIDir() string
}

// ToKubeAPIServerArguments returns updateArgs, deleteArgs that can be used with snaputil.UpdateServiceArguments() for the kube-apiserver
// according the datastore configuration.
func (c Datastore) ToKubeAPIServerArguments(p DatastorePathsProvider) (map[string]string, []string) {
	var (
		updateArgs = make(map[string]string)
		deleteArgs []string
	)

	switch c.GetType() {
	case "k8s-dqlite":
		updateArgs["--etcd-servers"] = fmt.Sprintf("unix://%s", filepath.Join(p.K8sDqliteStateDir(), "k8s-dqlite.sock"))
		// The watch list feature is enable by default on >1.32, but k8s-dqlite currently doesn't support it.
		updateArgs["--feature-gates"] = "WatchList=false"
		deleteArgs = []string{"--etcd-cafile", "--etcd-certfile", "--etcd-keyfile"}
	case "external":
		updateArgs["--etcd-servers"] = strings.Join(c.GetExternalServers(), ",")

		// the certificates will be written by setup.EnsureExtDatastorePKI(), here we only set the paths
		for _, loop := range []struct {
			arg  string
			cert string
			path string
		}{
			{cert: c.GetExternalCACert(), arg: "--etcd-cafile", path: "ca.crt"},
			{cert: c.GetExternalClientCert(), arg: "--etcd-certfile", path: "client.crt"},
			{cert: c.GetExternalClientKey(), arg: "--etcd-keyfile", path: "client.key"},
		} {
			if loop.cert != "" {
				updateArgs[loop.arg] = filepath.Join(p.EtcdPKIDir(), loop.path)
			} else {
				deleteArgs = append(deleteArgs, loop.arg)
			}
		}
	}

	return updateArgs, deleteArgs
}
