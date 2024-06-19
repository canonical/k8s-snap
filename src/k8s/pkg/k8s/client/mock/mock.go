package mock

import (
	"context"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
)

// Client is a mock implementation for k8s Client.
type Client struct {
	BootstrapCalledWith struct {
		Ctx     context.Context
		Request apiv1.PostClusterBootstrapRequest
	}
	BootstrapClusterMember apiv1.NodeStatus
	BootstrapErr           error
	IsBootstrappedReturn   bool
	CleanupNodeCalledWith  struct {
		Ctx      context.Context
		NodeName string
	}
	ClusterStatusReturn    apiv1.ClusterStatus
	ClusterStatusErr       error
	NodeStatusReturn       apiv1.NodeStatus
	NodeStatusErr          error
	GetJoinTokenCalledWith apiv1.GetJoinTokenRequest
	GetJoinTokenReturn     struct {
		Token string
		Err   error
	}
	GenerateAuthTokenCalledWith apiv1.GenerateKubernetesAuthTokenRequest
	GenerateAuthTokenReturn     struct {
		Token string
		Err   error
	}
	RevokeAuthTokenCalledWith  apiv1.RevokeKubernetesAuthTokenRequest
	RevokeAuthTokenErr         error
	JoinClusterCalledWith      apiv1.JoinClusterRequest
	JoinClusterErr             error
	KubeConfigReturn           string
	KubeConfigErr              error
	RemoveNodeCalledWith       apiv1.RemoveNodeRequest
	RemoveNodeErr              error
	GetClusterConfigCalledWith apiv1.GetClusterConfigRequest
	GetClusterConfigReturn     struct {
		Config apiv1.UserFacingClusterConfig
		Err    error
	}
	UpdateClusterConfigCalledWith    apiv1.UpdateClusterConfigRequest
	UpdateClusterConfigErr           error
	SetClusterAPIAuthTokenCalledWith apiv1.SetClusterAPIAuthTokenRequest
	SetClusterAPIAuthTokenErr        error
}

func (c *Client) Bootstrap(ctx context.Context, request apiv1.PostClusterBootstrapRequest) (apiv1.NodeStatus, error) {
	c.BootstrapCalledWith.Ctx = ctx
	c.BootstrapCalledWith.Request = request
	return c.BootstrapClusterMember, c.BootstrapErr
}

func (c *Client) IsBootstrapped(ctx context.Context) bool {
	return c.IsBootstrappedReturn
}

func (c *Client) CleanupNode(ctx context.Context, nodeName string) {
	c.CleanupNodeCalledWith.Ctx = ctx
	c.CleanupNodeCalledWith.NodeName = nodeName
}

func (c *Client) ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error) {
	return c.ClusterStatusReturn, c.ClusterStatusErr
}

func (c *Client) LocalNodeStatus(ctx context.Context) (apiv1.NodeStatus, error) {
	return c.NodeStatusReturn, c.NodeStatusErr
}

func (c *Client) GetJoinToken(ctx context.Context, request apiv1.GetJoinTokenRequest) (string, error) {
	c.GetJoinTokenCalledWith = request
	return c.GetJoinTokenReturn.Token, c.GetJoinTokenReturn.Err
}

func (c *Client) GenerateAuthToken(ctx context.Context, request apiv1.GenerateKubernetesAuthTokenRequest) (string, error) {
	c.GenerateAuthTokenCalledWith = request
	return c.GenerateAuthTokenReturn.Token, c.GenerateAuthTokenReturn.Err
}

func (c *Client) RevokeAuthToken(ctx context.Context, request apiv1.RevokeKubernetesAuthTokenRequest) error {
	c.RevokeAuthTokenCalledWith = request
	return c.RevokeAuthTokenErr
}

func (c *Client) JoinCluster(ctx context.Context, request apiv1.JoinClusterRequest) error {
	c.JoinClusterCalledWith = request
	return c.JoinClusterErr
}

func (c *Client) KubeConfig(ctx context.Context, request apiv1.GetKubeConfigRequest) (string, error) {
	return c.KubeConfigReturn, c.KubeConfigErr
}

func (c *Client) RemoveNode(ctx context.Context, request apiv1.RemoveNodeRequest) error {
	c.RemoveNodeCalledWith = request
	return c.RemoveNodeErr
}

func (c *Client) UpdateClusterConfig(ctx context.Context, request apiv1.UpdateClusterConfigRequest) error {
	c.UpdateClusterConfigCalledWith = request
	return c.UpdateClusterConfigErr
}

func (c *Client) GetClusterConfig(ctx context.Context, request apiv1.GetClusterConfigRequest) (apiv1.UserFacingClusterConfig, error) {
	c.GetClusterConfigCalledWith = request
	return c.GetClusterConfigReturn.Config, c.GetClusterConfigReturn.Err
}

func (c *Client) SetClusterAPIAuthToken(ctx context.Context, request apiv1.SetClusterAPIAuthTokenRequest) error {
	c.SetClusterAPIAuthTokenCalledWith = request
	return c.SetClusterAPIAuthTokenErr
}

var _ client.Client = &Client{}
