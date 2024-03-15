package snap

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/canonical/k8s/pkg/client/dqlite"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/vals"
	"github.com/moby/sys/mountinfo"
	"gopkg.in/yaml.v2"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// snap implements the Snap interface.
type snap struct {
	snapDir       string
	snapCommonDir string
	runCommand    func(ctx context.Context, command ...string) error
}

// NewSnap creates a new interface with the K8s snap.
// NewSnap accepts the $SNAP, $SNAP_COMMON, directories, and a number of options.
func NewSnap(snapDir, snapCommonDir string, options ...func(s *snap)) *snap {
	s := &snap{
		snapDir:       snapDir,
		snapCommonDir: snapCommonDir,
		runCommand:    utils.RunCommand,
	}

	for _, option := range options {
		option(s) // Apply each passed option to the snap instance
	}

	return s
}

func (s *snap) path(parts ...string) string {
	return path.Join(append([]string{s.snapDir}, parts...)...)
}

// serviceName infers the name of the snapctl daemon from the service name.
// if the serviceName is the snap name `k8s` (=referes to all services) it will return it as is.
func serviceName(serviceName string) string {
	if strings.HasPrefix(serviceName, "k8s.") || serviceName == "k8s" {
		return serviceName
	}
	return fmt.Sprintf("k8s.%s", serviceName)
}

// StartService starts a k8s service. The name can be either prefixed or not.
func (s *snap) StartService(ctx context.Context, name string) error {
	return s.runCommand(ctx, "snapctl", "start", "--enable", serviceName(name))
}

// StopService stops a k8s service. The name can be either prefixed or not.
func (s *snap) StopService(ctx context.Context, name string) error {
	return s.runCommand(ctx, "snapctl", "stop", "--disable", serviceName(name))
}

// RestartService restarts a k8s service. The name can be either prefixed or not.
func (s *snap) RestartService(ctx context.Context, name string) error {
	return s.runCommand(ctx, "snapctl", "restart", serviceName(name))
}

type snapcraftYml struct {
	Confinement string `yaml:"confinement"`
}

func (s *snap) Strict() bool {
	var meta snapcraftYml
	contents, err := os.ReadFile(s.path("meta", "snap.yaml"))
	if err != nil {
		return false
	}
	if err := yaml.Unmarshal([]byte(contents), &meta); err != nil {
		return false
	}
	return meta.Confinement == "strict"
}

func (s *snap) OnLXD(ctx context.Context) (bool, error) {
	mounts, err := mountinfo.GetMounts(mountinfo.FSTypeFilter("fuse.lxcfs"))
	if err != nil {
		return false, fmt.Errorf("failed to check for lxcfs mounts: %w", err)
	}
	return len(mounts) > 0, nil
}

func (s *snap) UID() int {
	return 0
}

func (s *snap) GID() int {
	return 0
}

func (s *snap) ContainerdConfigDir() string {
	return path.Join(s.snapCommonDir, "etc", "containerd")
}

func (s *snap) ContainerdRootDir() string {
	return path.Join(s.snapCommonDir, "var", "lib", "containerd")
}

func (s *snap) ContainerdSocketDir() string {
	return path.Join(s.snapCommonDir, "run")
}

func (s *snap) ContainerdStateDir() string {
	return "/run/containerd"
}

func (s *snap) CNIConfDir() string {
	return "/etc/cni/net.d"
}

func (s *snap) CNIBinDir() string {
	return "/opt/cni/bin"
}

func (s *snap) CNIPluginsBinary() string {
	return path.Join(s.snapDir, "bin", "cni")
}

func (s *snap) CNIPlugins() []string {
	return []string{
		"dhcp",
		"host-local",
		"static",
		"bridge",
		"host-device",
		"ipvlan",
		"loopback",
		"macvlan",
		"ptp",
		"vlan",
		"bandwidth",
		"firewall",
		"portmap",
		"sbr",
		"tuning",
		"vrf",
	}
}

func (s *snap) KubernetesConfigDir() string {
	return "/etc/kubernetes"
}

func (s *snap) KubernetesPKIDir() string {
	return "/etc/kubernetes/pki"
}

func (s *snap) EtcdPKIDir() string {
	return "/etc/kubernetes/pki/etcd"
}

func (s *snap) KubeletRootDir() string {
	return "/var/lib/kubelet"
}

func (s *snap) K8sdStateDir() string {
	return path.Join(s.snapCommonDir, "var", "lib", "k8sd", "state")
}

func (s *snap) K8sDqliteStateDir() string {
	return path.Join(s.snapCommonDir, "var", "lib", "k8s-dqlite")
}

func (s *snap) ServiceArgumentsDir() string {
	return path.Join(s.snapCommonDir, "args")
}

func (s *snap) ServiceExtraConfigDir() string {
	return path.Join(s.snapCommonDir, "args", "conf.d")
}

func (s *snap) LockFilesDir() string {
	return path.Join(s.snapCommonDir, "lock")
}

func (s *snap) ContainerdExtraConfigDir() string {
	return path.Join(s.snapCommonDir, "etc", "containerd", "conf.d")
}

func (s *snap) ContainerdRegistryConfigDir() string {
	return path.Join(s.snapCommonDir, "etc", "containerd", "hosts.d")
}

func (s *snap) Components() map[string]types.Component {
	return map[string]types.Component{
		"network": {
			ReleaseName:  "ck-network",
			ManifestPath: path.Join(s.snapDir, "k8s", "components", "charts", "cilium-1.14.1.tgz"),
			Namespace:    "kube-system",
		},
		"dns": {
			ReleaseName: "ck-dns",
			// TODO: fork coredns helm chart so that we can set custom args needed for the rock
			ManifestPath: path.Join(s.snapDir, "k8s", "components", "charts", "coredns-1.29.0"),
			Namespace:    "kube-system",
		},
		"local-storage": {
			ReleaseName:  "ck-storage",
			ManifestPath: path.Join(s.snapDir, "k8s", "components", "charts", "rawfile-csi-0.8.0.tgz"),
			Namespace:    "kube-system",
		},
		"ingress": {},
		"gateway": {
			ReleaseName:  "ck-gateway",
			ManifestPath: path.Join(s.snapDir, "k8s", "components", "charts", "gateway-api-0.7.1.tgz"),
			Namespace:    "kube-system",
		},
		"load-balancer": {
			ReleaseName:  "ck-loadbalancer",
			ManifestPath: path.Join(s.snapDir, "k8s", "components", "charts", "ck-loadbalancer"),
			Namespace:    "kube-system",
		},
		"metrics-server": {
			ReleaseName:  "metrics-server",
			ManifestPath: path.Join(s.snapDir, "k8s", "components", "charts", "metrics-server-3.12.0.tgz"),
			Namespace:    "kube-system",
		},
	}
}

func (s *snap) KubernetesRESTClientGetter(namespace string) genericclioptions.RESTClientGetter {
	flags := &genericclioptions.ConfigFlags{
		KubeConfig: &[]string{"/etc/kubernetes/admin.conf"}[0],
	}
	if namespace != "" {
		flags.Namespace = &namespace
	}
	return flags
}

func (s *snap) KubernetesNodeRESTClientGetter(namespace string) genericclioptions.RESTClientGetter {
	flags := &genericclioptions.ConfigFlags{
		KubeConfig: vals.Pointer(path.Join(s.KubernetesConfigDir(), "kubelet.conf")),
	}
	if namespace != "" {
		flags.Namespace = &namespace
	}
	return flags
}

func (s *snap) K8sDqliteClient(ctx context.Context) (*dqlite.Client, error) {
	client, err := dqlite.NewClient(ctx, dqlite.ClientOpts{
		ClusterYAML: path.Join(s.snapCommonDir, "var", "lib", "k8s-dqlite", "cluster.yaml"),
		ClusterCert: path.Join(s.snapCommonDir, "var", "lib", "k8s-dqlite", "cluster.crt"),
		ClusterKey:  path.Join(s.snapCommonDir, "var", "lib", "k8s-dqlite", "cluster.key"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create default k8s-dqlite client: %w", err)
	}
	return client, nil
}

var _ Snap = &snap{}
