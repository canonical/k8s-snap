package mock

import (
	"context"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/client/k8sd"
)

// Mock is a mock implementation of k8sd.Client.
type Mock struct {
	// k8sd.ClusterClient
	BootstrapClusterCalledWith apiv1.BootstrapClusterRequest
	BootstrapClusterResponse   apiv1.BootstrapClusterResponse
	BootstrapClusterErr        error
	GetJoinTokenCalledWith     apiv1.GetJoinTokenRequest
	GetJoinTokenResponse       apiv1.GetJoinTokenResponse
	GetJoinTokenErr            error
	JoinClusterCalledWith      apiv1.JoinClusterRequest
	JoinClusterErr             error
	RemoveNodeCalledWith       apiv1.RemoveNodeRequest
	RemoveNodeErr              error

	// k8sd.StatusClient
	NodeStatusResponse    apiv1.NodeStatusResponse
	NodeStatusInitialized bool
	NodeStatusErr         error
	ClusterStatusResponse apiv1.ClusterStatusResponse
	ClusterStatusErr      error

	// k8sd.ConfigClient
	GetClusterConfigResponse   apiv1.GetClusterConfigResponse
	GetClusterConfigErr        error
	SetClusterConfigCalledWith apiv1.SetClusterConfigRequest
	SetClusterConfigErr        error

	// k8sd.UserClient
	KubeConfigCalledWith apiv1.KubeConfigRequest
	KubeConfigResponse   apiv1.KubeConfigResponse
	KubeConfigErr        error

	// k8sd.ClusterAPIClient
	SetClusterAPIAuthTokenCalledWith apiv1.ClusterAPISetAuthTokenRequest
	SetClusterAPIAuthTokenErr        error
}

func (m *Mock) BootstrapCluster(_ context.Context, request apiv1.BootstrapClusterRequest) (apiv1.BootstrapClusterResponse, error) {
	m.BootstrapClusterCalledWith = request
	return m.BootstrapClusterResponse, m.BootstrapClusterErr
}
func (m *Mock) GetJoinToken(_ context.Context, request apiv1.GetJoinTokenRequest) (apiv1.GetJoinTokenResponse, error) {
	m.GetJoinTokenCalledWith = request
	return m.GetJoinTokenResponse, m.GetJoinTokenErr
}
func (m *Mock) JoinCluster(_ context.Context, request apiv1.JoinClusterRequest) error {
	m.JoinClusterCalledWith = request
	return m.JoinClusterErr
}
func (m *Mock) RemoveNode(_ context.Context, request apiv1.RemoveNodeRequest) error {
	m.RemoveNodeCalledWith = request
	return m.RemoveNodeErr
}

func (m *Mock) NodeStatus(_ context.Context) (apiv1.NodeStatusResponse, bool, error) {
	return m.NodeStatusResponse, m.NodeStatusInitialized, m.NodeStatusErr
}
func (m *Mock) ClusterStatus(_ context.Context, waitReady bool) (apiv1.ClusterStatusResponse, error) {
	return m.ClusterStatusResponse, m.ClusterStatusErr
}

func (m *Mock) GetClusterConfig(_ context.Context) (apiv1.GetClusterConfigResponse, error) {
	return m.GetClusterConfigResponse, m.GetClusterConfigErr
}
func (m *Mock) SetClusterConfig(_ context.Context, request apiv1.SetClusterConfigRequest) error {
	m.SetClusterConfigCalledWith = request
	return m.SetClusterConfigErr
}

func (m *Mock) KubeConfig(_ context.Context, request apiv1.KubeConfigRequest) (apiv1.KubeConfigResponse, error) {
	m.KubeConfigCalledWith = request
	return m.KubeConfigResponse, m.KubeConfigErr
}

func (m *Mock) SetClusterAPIAuthToken(_ context.Context, request apiv1.ClusterAPISetAuthTokenRequest) error {
	m.SetClusterAPIAuthTokenCalledWith = request
	return m.SetClusterAPIAuthTokenErr
}

var _ k8sd.Client = &Mock{}
