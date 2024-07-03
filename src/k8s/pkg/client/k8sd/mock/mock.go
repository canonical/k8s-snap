package mock

import (
	"context"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/client/k8sd"
)

// Mock is a mock implementation of k8sd.Client.
type Mock struct {
	// k8sd.ClusterClient
	BootstrapClusterCalledWith apiv1.PostClusterBootstrapRequest
	BootstrapClusterResult     apiv1.NodeStatus
	BootstrapClusterErr        error
	GetJoinTokenCalledWith     apiv1.GetJoinTokenRequest
	GetJoinTokenResult         apiv1.GetJoinTokenResponse
	GetJoinTokenErr            error
	JoinClusterCalledWith      apiv1.JoinClusterRequest
	JoinClusterErr             error
	RemoveNodeCalledWith       apiv1.RemoveNodeRequest
	RemoveNodeErr              error

	// k8sd.StatusClient
	NodeStatusResult    apiv1.NodeStatus
	NodeStatusErr       error
	ClusterStatusResult apiv1.ClusterStatus
	ClusterStatusErr    error

	// k8sd.ConfigClient
	GetClusterConfigResult     apiv1.UserFacingClusterConfig
	GetClusterConfigErr        error
	SetClusterConfigCalledWith apiv1.UpdateClusterConfigRequest
	SetClusterConfigErr        error

	// k8sd.UserClient
	KubeConfigCalledWith apiv1.GetKubeConfigRequest
	KubeConfigResult     string
	KubeConfigErr        error

	// k8sd.ClusterAPIClient
	SetClusterAPIAuthTokenCalledWith apiv1.SetClusterAPIAuthTokenRequest
	SetClusterAPIAuthTokenErr        error
}

func (m *Mock) BootstrapCluster(_ context.Context, request apiv1.PostClusterBootstrapRequest) (apiv1.NodeStatus, error) {
	m.BootstrapClusterCalledWith = request
	return m.BootstrapClusterResult, m.BootstrapClusterErr
}
func (m *Mock) GetJoinToken(_ context.Context, request apiv1.GetJoinTokenRequest) (apiv1.GetJoinTokenResponse, error) {
	m.GetJoinTokenCalledWith = request
	return m.GetJoinTokenResult, m.GetJoinTokenErr
}
func (m *Mock) JoinCluster(_ context.Context, request apiv1.JoinClusterRequest) error {
	m.JoinClusterCalledWith = request
	return m.JoinClusterErr
}
func (m *Mock) RemoveNode(_ context.Context, request apiv1.RemoveNodeRequest) error {
	m.RemoveNodeCalledWith = request
	return m.RemoveNodeErr
}

func (m *Mock) NodeStatus(_ context.Context) (apiv1.NodeStatus, error) {
	return m.NodeStatusResult, m.NodeStatusErr
}
func (m *Mock) ClusterStatus(_ context.Context, waitReady bool) (apiv1.ClusterStatus, error) {
	return m.ClusterStatusResult, m.ClusterStatusErr
}

func (m *Mock) GetClusterConfig(_ context.Context) (apiv1.UserFacingClusterConfig, error) {
	return m.GetClusterConfigResult, m.GetClusterConfigErr
}
func (m *Mock) SetClusterConfig(_ context.Context, request apiv1.UpdateClusterConfigRequest) error {
	m.SetClusterConfigCalledWith = request
	return m.SetClusterConfigErr
}

func (m *Mock) KubeConfig(_ context.Context, request apiv1.GetKubeConfigRequest) (string, error) {
	m.KubeConfigCalledWith = request
	return m.KubeConfigResult, m.KubeConfigErr
}

func (m *Mock) SetClusterAPIAuthToken(_ context.Context, request apiv1.SetClusterAPIAuthTokenRequest) error {
	m.SetClusterAPIAuthTokenCalledWith = request
	return m.SetClusterAPIAuthTokenErr
}

var _ k8sd.Client = &Mock{}
