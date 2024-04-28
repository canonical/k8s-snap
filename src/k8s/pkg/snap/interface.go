package snap

import (
	"context"

	"github.com/canonical/k8s/pkg/client/dqlite"
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/client/kubernetes"
)

// Snap abstracts file system paths and interacting with the k8s services.
type Snap interface {
	Strict() bool                        // Strict returns true if the snap is installed with strict confinement.
	OnLXD(context.Context) (bool, error) // OnLXD returns true if the host runs on LXD.

	UID() int // UID is the user ID to set on config files.
	GID() int // GID is the group ID to set on config files.

	StartService(ctx context.Context, serviceName string) error   // snapctl start $service
	StopService(ctx context.Context, serviceName string) error    // snapctl stop $service
	RestartService(ctx context.Context, serviceName string) error // snapctl restart $service

	CNIConfDir() string       // /etc/cni/net.d
	CNIBinDir() string        // /opt/cni/bin
	CNIPluginsBinary() string // /snap/k8s/current/bin/cni
	CNIPlugins() []string     // cni plugins built into the cni binary

	KubernetesConfigDir() string // /etc/kubernetes
	KubernetesPKIDir() string    // /etc/kubernetes/pki
	EtcdPKIDir() string          // /etc/kubernetes/pki/etcd
	KubeletRootDir() string      // /var/lib/kubelet

	ContainerdConfigDir() string         // /var/snap/k8s/common/etc/containerd
	ContainerdExtraConfigDir() string    // /var/snap/k8s/common/etc/containerd/conf.d
	ContainerdRegistryConfigDir() string // /var/snap/k8s/common/etc/containerd/hosts.d
	ContainerdRootDir() string           // /var/snap/k8s/common/var/lib/containerd
	ContainerdSocketDir() string         // /var/snap/k8s/common/run
	ContainerdStateDir() string          // /run/containerd

	K8sdStateDir() string      // /var/snap/k8s/common/var/lib/k8sd/state
	K8sDqliteStateDir() string // /var/snap/k8s/common/var/lib/k8s-dqlite

	ServiceArgumentsDir() string   // /var/snap/k8s/common/args
	ServiceExtraConfigDir() string // /var/snap/k8s/common/args/conf.d

	LockFilesDir() string // /var/snap/k8s/common/lock

	KubernetesClient(namespace string) (*kubernetes.Client, error)     // admin kubernetes client
	KubernetesNodeClient(namespace string) (*kubernetes.Client, error) // node kubernetes client

	HelmClient() helm.Client // admin helm client

	K8sDqliteClient(ctx context.Context) (*dqlite.Client, error) // go-dqlite client for k8s-dqlite
}
