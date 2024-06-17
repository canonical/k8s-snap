package snap

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/dqlite"
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/moby/sys/mountinfo"
	"gopkg.in/yaml.v2"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type SnapOpts struct {
	SnapDir       string
	SnapCommonDir string
	RunCommand    func(ctx context.Context, command []string, opts ...func(c *exec.Cmd)) error
}

// snap implements the Snap interface.
type snap struct {
	snapDir       string
	snapCommonDir string
	runCommand    func(ctx context.Context, command []string, opts ...func(c *exec.Cmd)) error
}

// NewSnap creates a new interface with the K8s snap.
// NewSnap accepts the $SNAP, $SNAP_COMMON, directories, and a number of options.
func NewSnap(opts SnapOpts) *snap {
	runCommand := utils.RunCommand
	if opts.RunCommand != nil {
		runCommand = opts.RunCommand
	}
	s := &snap{
		snapDir:       opts.SnapDir,
		snapCommonDir: opts.SnapCommonDir,
		runCommand:    runCommand,
	}

	return s
}

// StartService starts a k8s service. The name can be either prefixed or not.
func (s *snap) StartService(ctx context.Context, name string) error {
	return s.runCommand(ctx, []string{"snapctl", "start", "--enable", serviceName(name)})
}

// StopService stops a k8s service. The name can be either prefixed or not.
func (s *snap) StopService(ctx context.Context, name string) error {
	return s.runCommand(ctx, []string{"snapctl", "stop", "--disable", serviceName(name)})
}

// RestartService restarts a k8s service. The name can be either prefixed or not.
func (s *snap) RestartService(ctx context.Context, name string) error {
	return s.runCommand(ctx, []string{"snapctl", "restart", serviceName(name)})
}

type snapcraftYml struct {
	Confinement string `yaml:"confinement"`
}

func (s *snap) Strict() bool {
	var meta snapcraftYml
	contents, err := os.ReadFile(filepath.Join(s.snapDir, "meta", "snap.yaml"))
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

func (s *snap) Hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "dev"
	}
	return hostname
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

func (s *snap) restClientGetter(path string, namespace string) genericclioptions.RESTClientGetter {
	flags := &genericclioptions.ConfigFlags{
		KubeConfig: utils.Pointer(path),
	}
	if namespace != "" {
		flags.Namespace = &namespace
	}
	return flags
}

func (s *snap) KubernetesClient(namespace string) (*kubernetes.Client, error) {
	return kubernetes.NewClient(s.restClientGetter(path.Join(s.KubernetesConfigDir(), "admin.conf"), namespace))
}

func (s *snap) KubernetesNodeClient(namespace string) (*kubernetes.Client, error) {
	return kubernetes.NewClient(s.restClientGetter(path.Join(s.KubernetesConfigDir(), "kubelet.conf"), namespace))
}

func (s *snap) HelmClient() helm.Client {
	return helm.NewClient(
		filepath.Join(s.snapDir, "k8s", "manifests"),
		func(namespace string) genericclioptions.RESTClientGetter {
			return s.restClientGetter(path.Join(s.KubernetesConfigDir(), "admin.conf"), namespace)
		},
	)
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

func (s *snap) SnapctlGet(ctx context.Context, args ...string) ([]byte, error) {
	var b bytes.Buffer
	if err := s.runCommand(ctx, append([]string{"snapctl", "get"}, args...), func(c *exec.Cmd) { c.Stdout = &b }); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (s *snap) SnapctlSet(ctx context.Context, args ...string) error {
	return s.runCommand(ctx, append([]string{"snapctl", "set"}, args...))
}

func (s *snap) PreInitChecks(ctx context.Context, config types.ClusterConfig) error {
	// TODO: check for available ports for k8s-dqlite, apiserver, containerd, etc

	// NOTE(neoaggelos): in some environments the Kubernetes might hang when running for the first time
	// This works around the issue by running them once during the install hook
	for _, binary := range []string{"kube-apiserver", "kube-controller-manager", "kube-scheduler", "kube-proxy", "kubelet"} {
		if err := s.runCommand(ctx, []string{filepath.Join(s.snapDir, "bin", binary), "--version"}); err != nil {
			return fmt.Errorf("%q binary could not run: %w", binary, err)
		}
	}

	return nil
}

var _ Snap = &snap{}
