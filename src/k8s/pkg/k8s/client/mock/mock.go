package mock

import (
	"context"

	apiv1 "github.com/canonical/k8s/api/v1"
)

// Client is a mock implementation for k8s Client.
type Client struct {
	BootstrapCalledWith              apiv1.BootstrapConfig
	BootstrapClusterMember           apiv1.NodeStatus
	BootstrapErr                     error
	IsBootstrappedReturn             bool
	IsKubernetesAPIServerReadyReturn bool
	CleanupNodeCalledWith            struct {
		Ctx      context.Context
		NodeName string
	}
	ClusterStatusReturn   apiv1.ClusterStatus
	ClusterStatusErr      error
	NodeStatusReturn      apiv1.NodeStatus
	NodeStatusErr         error
	CreateJoinTokenReturn struct {
		Token string
		Err   error
	}
	GenerateAuthTokenReturn struct {
		Token string
		Err   error
	}
	RevokeAuthTokenErr    error
	JoinClusterCalledWith struct {
		Ctx     context.Context
		Name    string
		Address string
		Token   string
	}
	JoinClusterErr       error
	KubeConfigReturn     string
	KubeConfigErr        error
	RemoveNodeCalledWith struct {
		Ctx   context.Context
		Name  string
		Force bool
	}
	RemoveNodeErr              error
	GetClusterConfigCalledWith apiv1.GetClusterConfigRequest
	GetClusterConfigReturn     struct {
		Config apiv1.UserFacingClusterConfig
		Err    error
	}
	UpdateClusterConfigCalledWith apiv1.UpdateClusterConfigRequest
	UpdateClusterConfigErr        error
}

func (c *Client) Bootstrap(ctx context.Context, bootstrapConfig apiv1.BootstrapConfig) (apiv1.NodeStatus, error) {
	c.BootstrapCalledWith = bootstrapConfig
	return c.BootstrapClusterMember, c.BootstrapErr
}

func (c *Client) IsKubernetesAPIServerReady(ctx context.Context) bool {
	return c.IsKubernetesAPIServerReadyReturn
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

func (c *Client) NodeStatus(ctx context.Context) (apiv1.NodeStatus, error) {
	return c.NodeStatusReturn, c.NodeStatusErr
}

func (c *Client) CreateJoinToken(ctx context.Context, name string, worker bool) (string, error) {
	return c.CreateJoinTokenReturn.Token, c.CreateJoinTokenReturn.Err
}

func (c *Client) GenerateAuthToken(ctx context.Context, username string, groups []string) (string, error) {
	return c.GenerateAuthTokenReturn.Token, c.GenerateAuthTokenReturn.Err
}

func (c *Client) RevokeAuthToken(ctx context.Context, token string) error {
	return c.RevokeAuthTokenErr
}

func (c *Client) JoinCluster(ctx context.Context, name string, address string, token string) error {
	c.JoinClusterCalledWith.Ctx = ctx
	c.JoinClusterCalledWith.Name = name
	c.JoinClusterCalledWith.Address = address
	c.JoinClusterCalledWith.Token = token
	return c.JoinClusterErr
}

func (c *Client) KubeConfig(ctx context.Context, server string) (string, error) {
	return c.KubeConfigReturn, c.KubeConfigErr
}

func (c *Client) RemoveNode(ctx context.Context, name string, force bool) error {
	c.RemoveNodeCalledWith.Ctx = ctx
	c.RemoveNodeCalledWith.Name = name
	c.RemoveNodeCalledWith.Force = force
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
