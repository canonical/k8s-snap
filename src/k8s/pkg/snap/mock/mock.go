package mock

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type Mock struct {
	Strict                      bool
	UID                         int
	GID                         int
	KubernetesConfigDir         string
	KubernetesPKIDir            string
	KubeletRootDir              string
	CNIConfDir                  string
	CNIBinDir                   string
	CNIPlugins                  []string
	CNIPluginsBinary            string
	ContainerdConfigDir         string
	ContainerdExtraConfigDir    string
	ContainerdRegistryConfigDir string
	ContainerdRootDir           string
	ContainerdSocketDir         string
	ContainerdStateDir          string
	K8sdStateDir                string
	K8sDqliteStateDir           string
	ServiceArgumentsDir         string
	ServiceExtraConfigDir       string
	Components                  map[string]types.Component
	KubernetesRESTClientGetter  genericclioptions.RESTClientGetter
}

// Snap is a mock implementation for snap.Snap.
type Snap struct {
	StartServiceCalledWith   []string
	StartServiceErr          error
	StopServiceCalledWith    []string
	StopServiceErr           error
	RestartServiceCalledWith []string
	RestartServiceErr        error

	Mock Mock
}

func (s *Snap) StartService(ctx context.Context, name string) error {
	if len(s.StartServiceCalledWith) == 0 {
		s.StartServiceCalledWith = []string{name}
	} else {
		s.StartServiceCalledWith = append(s.StartServiceCalledWith, name)
	}
	return s.StartServiceErr
}
func (s *Snap) StopService(ctx context.Context, name string) error {
	if len(s.StopServiceCalledWith) == 0 {
		s.StopServiceCalledWith = []string{name}
	} else {
		s.StopServiceCalledWith = append(s.StopServiceCalledWith, name)
	}
	return s.StopServiceErr
}
func (s *Snap) RestartService(ctx context.Context, name string) error {
	if len(s.RestartServiceCalledWith) == 0 {
		s.RestartServiceCalledWith = []string{name}
	} else {
		s.RestartServiceCalledWith = append(s.RestartServiceCalledWith, name)
	}
	return s.RestartServiceErr
}

func (s *Snap) Strict() bool {
	return s.Mock.Strict
}
func (s *Snap) UID() int {
	return s.Mock.UID
}
func (s *Snap) GID() int {
	return s.Mock.GID
}
func (s *Snap) ContainerdConfigDir() string {
	return s.Mock.ContainerdConfigDir
}
func (s *Snap) ContainerdRootDir() string {
	return s.Mock.ContainerdRootDir
}
func (s *Snap) ContainerdStateDir() string {
	return s.Mock.ContainerdStateDir
}
func (s *Snap) ContainerdSocketDir() string {
	return s.Mock.ContainerdSocketDir
}
func (s *Snap) ContainerdExtraConfigDir() string {
	return s.Mock.ContainerdExtraConfigDir
}
func (s *Snap) ContainerdRegistryConfigDir() string {
	return s.Mock.ContainerdRegistryConfigDir
}
func (s *Snap) KubernetesConfigDir() string {
	return s.Mock.KubernetesConfigDir
}
func (s *Snap) KubernetesPKIDir() string {
	return s.Mock.KubernetesPKIDir
}
func (s *Snap) KubeletRootDir() string {
	return s.Mock.KubeletRootDir
}
func (s *Snap) CNIConfDir() string {
	return s.Mock.CNIConfDir
}
func (s *Snap) CNIBinDir() string {
	return s.Mock.CNIBinDir
}
func (s *Snap) CNIPluginsBinary() string {
	return s.Mock.CNIPluginsBinary
}
func (s *Snap) CNIPlugins() []string {
	return s.Mock.CNIPlugins
}
func (s *Snap) K8sdStateDir() string {
	return s.Mock.K8sdStateDir
}
func (s *Snap) K8sDqliteStateDir() string {
	return s.Mock.K8sDqliteStateDir
}
func (s *Snap) ServiceArgumentsDir() string {
	return s.Mock.ServiceArgumentsDir
}
func (s *Snap) ServiceExtraConfigDir() string {
	return s.Mock.ServiceExtraConfigDir
}
func (s *Snap) Components() map[string]types.Component {
	return s.Mock.Components
}
func (s *Snap) KubernetesRESTClientGetter() genericclioptions.RESTClientGetter {
	return s.Mock.KubernetesRESTClientGetter
}

var _ snap.Snap = &Snap{}
