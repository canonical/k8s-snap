package snap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/canonical/k8s/pkg/utils"
	"gopkg.in/yaml.v2"
)

// snap implements the Snap interface.
type snap struct {
	snapDir       string
	snapCommonDir string
}

// NewSnap creates a new interface with the K8s snap.
// NewSnap accepts the $SNAP and $SNAP_COMMON directories
func NewSnap(snapDir, snapCommonDir string) *snap {
	s := &snap{
		snapDir:       snapDir,
		snapCommonDir: snapCommonDir,
	}

	return s
}

func (s *snap) path(parts ...string) string {
	return filepath.Join(append([]string{s.snapDir}, parts...)...)
}

func (s *snap) commonPath(parts ...string) string {
	return filepath.Join(append([]string{s.snapCommonDir}, parts...)...)
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
	return utils.RunCommand(ctx, "snapctl", "start", serviceName(name))
}

// StopService stops a k8s service. The name can be either prefixed or not.
func (s *snap) StopService(ctx context.Context, name string) error {
	return utils.RunCommand(ctx, "snapctl", "stop", serviceName(name))
}

// RestartService restarts a k8s service. The name can be either prefixed or not.
func (s *snap) RestartService(ctx context.Context, name string) error {
	return utils.RunCommand(ctx, "snapctl", "restart", serviceName(name))
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

func (s *snap) UID() int {
	return 0
}

func (s *snap) GID() int {
	return 0
}

func (s *snap) ContainerdConfigDir() string {
	return filepath.Join(s.snapCommonDir, "etc", "containerd")
}

func (s *snap) ContainerdRootDir() string {
	return filepath.Join(s.snapCommonDir, "var", "lib", "containerd")
}

func (s *snap) ContainerdSocketDir() string {
	return filepath.Join(s.snapCommonDir, "run")
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
	return filepath.Join(s.snapDir, "bin", "cni")
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

func (s *snap) KubeletRootDir() string {
	return "/var/lib/kubelet"
}

func (s *snap) K8sdStateDir() string {
	return filepath.Join(s.snapCommonDir, "var", "lib", "k8sd", "state")
}

func (s *snap) K8sDqliteStateDir() string {
	return filepath.Join(s.snapCommonDir, "var", "lib", "k8s-dqlite")
}

func (s *snap) ServiceArgumentsDir() string {
	return filepath.Join(s.snapCommonDir, "args")
}

func (s *snap) ServiceExtraConfigDir() string {
	return filepath.Join(s.snapCommonDir, "args", "conf.d")
}

func (s *snap) ContainerdExtraConfigDir() string {
	return filepath.Join(s.snapCommonDir, "etc", "containerd", "conf.d")
}

func (s *snap) ContainerdRegistryConfigDir() string {
	return filepath.Join(s.snapCommonDir, "etc", "containerd", "hosts.d")
}

var _ Snap = &snap{}
