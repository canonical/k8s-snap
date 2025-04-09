package mock

import (
	"context"
	"strings"

	"github.com/canonical/k8s/pkg/client/dqlite"
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/client/k8sd"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

type Mock struct {
	Revision                    string
	RevisionErr                 error
	Strict                      bool
	OnLXD                       bool
	OnLXDErr                    error
	UID                         int
	GID                         int
	Hostname                    string
	KubernetesConfigDir         string
	KubernetesPKIDir            string
	EtcdPKIDir                  string
	KubeletRootDir              string
	CNIConfDir                  string
	CNIBinDir                   string
	CNIPlugins                  []string
	CNIPluginsBinary            string
	ContainerdConfigDir         string
	ContainerdExtraConfigDir    string
	ContainerdRegistryConfigDir string
	ContainerdBaseDir           string
	ContainerdRootDir           string
	ContainerdSocketDir         string
	ContainerdSocketPath        string
	ContainerdStateDir          string
	K8sCRDDir                   string
	K8sScriptsDir               string
	K8sInspectScriptPath        string
	K8sdStateDir                string
	K8sDqliteStateDir           string
	ServiceArgumentsDir         string
	ServiceExtraConfigDir       string
	LockFilesDir                string
	PostRefreshLockPath         string
	NodeTokenFile               string
	KubernetesClient            *kubernetes.Client
	KubernetesNodeClient        *kubernetes.Client
	HelmClient                  helm.Client
	K8sDqliteClient             *dqlite.Client
	K8sdClient                  k8sd.Client
	SnapctlGet                  map[string][]byte
}

// Snap is a mock implementation for snap.Snap.
type Snap struct {
	StartServicesCalledWith   [][]string
	StartServicesErr          error
	StopServicesCalledWith    [][]string
	StopServicesErr           error
	RestartServicesCalledWith [][]string
	RestartServicesErr        error

	RefreshCalledWith []types.RefreshOpts
	RefreshErr        error

	SnapctlSetCalledWith [][]string
	SnapctlSetErr        error
	SnapctlGetCalledWith [][]string
	SnapctlGetErr        error

	PreInitChecksCalledWith []types.ClusterConfig
	PreInitChecksErr        error

	Mock Mock
}

func (s *Snap) StartServices(ctx context.Context, names []string, extraSnapArgs ...string) error {
	if len(s.StartServicesCalledWith) == 0 {
		s.StartServicesCalledWith = [][]string{names}
	} else {
		s.StartServicesCalledWith = append(s.StartServicesCalledWith, names)
	}
	if len(extraSnapArgs) > 0 {
		s.StartServicesCalledWith = append(s.StartServicesCalledWith, extraSnapArgs)
	}
	return s.StartServicesErr
}

func (s *Snap) StopServices(ctx context.Context, names []string, extraSnapArgs ...string) error {
	if len(s.StopServicesCalledWith) == 0 {
		s.StopServicesCalledWith = [][]string{names}
	} else {
		s.StopServicesCalledWith = append(s.StopServicesCalledWith, names)
	}
	if len(extraSnapArgs) > 0 {
		s.StopServicesCalledWith = append(s.StopServicesCalledWith, extraSnapArgs)
	}
	return s.StopServicesErr
}

func (s *Snap) RestartServices(ctx context.Context, names []string, extraSnapArgs ...string) error {
	if len(s.RestartServicesCalledWith) == 0 {
		s.RestartServicesCalledWith = [][]string{names}
	} else {
		s.RestartServicesCalledWith = append(s.RestartServicesCalledWith, names)
	}
	if len(extraSnapArgs) > 0 {
		s.RestartServicesCalledWith = append(s.RestartServicesCalledWith, extraSnapArgs)
	}
	return s.RestartServicesErr
}

func (s *Snap) Refresh(ctx context.Context, opts types.RefreshOpts) (string, error) {
	if len(s.RefreshCalledWith) == 0 {
		s.RefreshCalledWith = []types.RefreshOpts{opts}
	} else {
		s.RefreshCalledWith = append(s.RefreshCalledWith, opts)
	}
	return "", s.RefreshErr
}

func (s *Snap) RefreshStatus(ctx context.Context, changeID string) (*types.RefreshStatus, error) {
	return nil, nil
}

func (s *Snap) PostRefreshLockPath() string {
	return s.Mock.PostRefreshLockPath
}

func (s *Snap) Revision(ctx context.Context) (string, error) {
	return s.Mock.Revision, s.Mock.RevisionErr
}

func (s *Snap) Strict() bool {
	return s.Mock.Strict
}

func (s *Snap) OnLXD(context.Context) (bool, error) {
	return s.Mock.OnLXD, s.Mock.OnLXDErr
}

func (s *Snap) UID() int {
	return s.Mock.UID
}

func (s *Snap) GID() int {
	return s.Mock.GID
}

func (s *Snap) Hostname() string {
	return s.Mock.Hostname
}

func (s *Snap) SetContainerdBaseDir(baseDir string) {
	s.Mock.ContainerdBaseDir = baseDir
}

func (s *Snap) GetContainerdBaseDir() string {
	return s.Mock.ContainerdBaseDir
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

func (s *Snap) ContainerdSocketPath() string {
	return s.Mock.ContainerdSocketPath
}

func (s *Snap) ContainerdExtraConfigDir() string {
	return s.Mock.ContainerdExtraConfigDir
}

func (s *Snap) ContainerdRegistryConfigDir() string {
	return s.Mock.ContainerdRegistryConfigDir
}

func (s *Snap) K8sCRDDir() string {
	return s.Mock.K8sCRDDir
}

func (s *Snap) K8sScriptsDir() string {
	return s.Mock.K8sScriptsDir
}

func (s *Snap) K8sInspectScriptPath() string {
	return s.Mock.K8sInspectScriptPath
}

func (s *Snap) KubernetesConfigDir() string {
	return s.Mock.KubernetesConfigDir
}

func (s *Snap) KubernetesPKIDir() string {
	return s.Mock.KubernetesPKIDir
}

func (s *Snap) EtcdPKIDir() string {
	return s.Mock.EtcdPKIDir
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

func (s *Snap) LockFilesDir() string {
	return s.Mock.LockFilesDir
}

func (s *Snap) NodeTokenFile() string {
	return s.Mock.NodeTokenFile
}

func (s *Snap) KubernetesClient(namespace string) (*kubernetes.Client, error) {
	return s.Mock.KubernetesClient, nil
}

func (s *Snap) KubernetesNodeClient(namespace string) (*kubernetes.Client, error) {
	return s.Mock.KubernetesNodeClient, nil
}

func (s *Snap) HelmClient() helm.Client {
	return s.Mock.HelmClient
}

func (s *Snap) K8sDqliteClient(context.Context) (*dqlite.Client, error) {
	return s.Mock.K8sDqliteClient, nil
}

func (s *Snap) K8sdClient(address string) (k8sd.Client, error) {
	return s.Mock.K8sdClient, nil
}

func (s *Snap) SnapctlGet(ctx context.Context, args ...string) ([]byte, error) {
	s.SnapctlGetCalledWith = append(s.SnapctlGetCalledWith, args)
	return s.Mock.SnapctlGet[strings.Join(args, " ")], s.SnapctlGetErr
}

func (s *Snap) SnapctlSet(ctx context.Context, args ...string) error {
	s.SnapctlSetCalledWith = append(s.SnapctlSetCalledWith, args)
	return s.SnapctlSetErr
}

func (s *Snap) PreInitChecks(ctx context.Context, config types.ClusterConfig, serviceConfigs types.K8sServiceConfigs, isControlPlane bool) error {
	s.PreInitChecksCalledWith = append(s.PreInitChecksCalledWith, config)
	return s.PreInitChecksErr
}

var _ snap.Snap = &Snap{}
