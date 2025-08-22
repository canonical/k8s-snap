package snap

import (
	"context"

	"github.com/canonical/k8s/pkg/client/dqlite"
	"github.com/canonical/k8s/pkg/client/etcd"
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/client/k8sd"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

// Snap abstracts file system paths and interacting with the k8s services.
type Snap interface {
	Revision(ctx context.Context) (string, error) // Revision returns the snap revision.
	Strict() bool                                 // Strict returns true if the snap is installed with strict confinement.
	OnLXD(context.Context) (bool, error)          // OnLXD returns true if the host runs on LXD.

	UID() int         // UID is the user ID to set on config files.
	GID() int         // GID is the group ID to set on config files.
	Hostname() string // Hostname is the name of the node.

	StartServices(ctx context.Context, services []string, extraSnapArgs ...string) error   // snap start $service
	StopServices(ctx context.Context, services []string, extraSnapArgs ...string) error    // snap stop $service
	RestartServices(ctx context.Context, services []string, extraSnapArgs ...string) error // snap restart $service

	SnapctlGet(ctx context.Context, args ...string) ([]byte, error) // snapctl get $args...
	SnapctlSet(ctx context.Context, args ...string) error           // snapctl set $args...

	Refresh(ctx context.Context, to types.RefreshOpts) (string, error)                // snap refresh --no-wait [k8s --channel $track | k8s --revision $revision | $path ]
	RefreshStatus(ctx context.Context, changeID string) (*types.RefreshStatus, error) // snap tasks $changeID
	PostRefreshLockPath() string                                                      // /var/snap/k8s/common/lock/post-refresh - lock file to indicate the first run after a snap refresh

	SystemTuningConfigDir() string             //  /etc/sysctl.d
	SystemConfigDirs() []string                // /etc/sysctl.d/, /run/sysctl.d/, /usr/local/lib/sysctl.d/, /usr/lib/sysctl.d/, /lib/sysctl.d/, /etc/sysctl.conf
	SystemMinConfig() map[string]string        // system limits: fs.inotify parameters
	SystemComplianceConfig() map[string]string // system compliance config: vm.overcommit_memory kernel.panic kernel.panic_on_oops

	CNIConfDir() string       // /etc/cni/net.d
	CNIBinDir() string        // /opt/cni/bin
	CNIPluginsBinary() string // /snap/k8s/current/bin/cni
	CNIPlugins() []string     // cni plugins built into the cni binary

	KubernetesConfigDir() string // /etc/kubernetes
	KubernetesPKIDir() string    // /etc/kubernetes/pki
	EtcdPKIDir() string          // /etc/kubernetes/pki/etcd
	KubeletRootDir() string      // /var/lib/kubelet

	SetContainerdBaseDir(baseDir string) // sets the containerd base directory.
	GetContainerdBaseDir() string        // gets the containerd base directory.
	ContainerdConfigDir() string         // classic confinement: /etc/containerd, strict confinement: /var/snap/k8s/common/etc/containerd
	ContainerdExtraConfigDir() string    // classic confinement: /etc/containerd/conf.d, strict confinement: /var/snap/k8s/common/etc/containerd/conf.d
	ContainerdRegistryConfigDir() string // classic confinement: /etc/containerd/hosts.d, strict confinement: /var/snap/k8s/common/etc/containerd/hosts.d
	ContainerdRootDir() string           // classic confinement: /var/lib/containerd, strict confinement: /var/snap/k8s/common/var/lib/containerd
	ContainerdSocketDir() string         // classic confinement: /run/containerd, strict confinement: /var/snap/k8s/common/run/containerd
	ContainerdSocketPath() string        // classic confinement: /run/containerd/containerd.sock, strict confinement: /var/snap/k8s/common/run/containerd/containerd.sock
	ContainerdStateDir() string          // classic confinement: /run/containerd, strict confinement: /var/snap/k8s/common/run/containerd

	K8sCRDDir() string            //  /snap/k8s/current/k8s/crds
	K8sScriptsDir() string        //  /snap/k8s/current/k8s/scripts
	K8sBinDir() string            //  /snap/k8s/current/bin
	K8sInspectScriptPath() string //  /snap/k8s/current/k8s/scripts/inspect.sh

	K8sdStateDir() string      // /var/snap/k8s/common/var/lib/k8sd/state
	K8sDqliteStateDir() string // /var/snap/k8s/common/var/lib/k8s-dqlite
	EtcdDir() string           // /var/snap/k8s/common/var/lib/etcd

	ServiceArgumentsDir() string   // /var/snap/k8s/common/args
	ServiceExtraConfigDir() string // /var/snap/k8s/common/args/conf.d

	EtcDir() string       // /var/snap/k8s/common/etc
	LockFilesDir() string // /var/snap/k8s/common/lock

	NodeTokenFile() string                                     // /var/snap/k8s/common/node-token
	NodeKubernetesVersion(ctx context.Context) (string, error) // The Kubernetes version of the node as set in the snap. Can be queried without running k8s services.

	KubernetesClient(namespace string) (*kubernetes.Client, error)     // admin kubernetes client
	KubernetesNodeClient(namespace string) (*kubernetes.Client, error) // node kubernetes client

	HelmClient() helm.Client // admin helm client

	K8sDqliteClient(ctx context.Context) (*dqlite.Client, error) // go-dqlite client for k8s-dqlite

	EtcdClient(endpoints []string) (*etcd.Client, error) // client for the managed etcd cluster

	K8sdClient(address string) (k8sd.Client, error) // k8sd client

	PreInitChecks(ctx context.Context, config types.ClusterConfig, serviceConfigs types.K8sServiceConfigs, isControlPlane bool) error // pre-init checks before k8s-snap can start
}
