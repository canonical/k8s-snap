package mock

import (
	"context"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/snap"
)

// Client is a mock implementation for k8s Client.
type Client struct {
	BootstrapCalledWith    apiv1.BootstrapConfig
	BootstrapClusterMember apiv1.ClusterMember
	BootstrapErr           error
	IsBootstrappedReturn   bool
	CleanupNodeCalledWith  struct {
		Ctx      context.Context
		Snap     snap.Snap
		NodeName string
	}
	ClusterStatusReturn   apiv1.ClusterStatus
	ClusterStatusErr      error
	CreateJoinTokenReturn struct {
		Token string
		Err   error
	}
	GenerateAuthTokenReturn struct {
		Token string
		Err   error
	}
	JoinClusterCalledWith struct {
		Ctx     context.Context
		Name    string
		Address string
		Token   string
	}
	JoinClusterErr       error
	KubeConfigReturn     string
	KubeConfigErr        error
	ListComponentsReturn []apiv1.Component
	ListComponentsErr    error
	RemoveNodeCalledWith struct {
		Ctx   context.Context
		Name  string
		Force bool
	}
	RemoveNodeErr                error
	UpdateDNSComponentCalledWith struct {
		Ctx     context.Context
		Request apiv1.UpdateDNSComponentRequest
	}
	UpdateDNSComponentErr            error
	UpdateGatewayComponentCalledWith struct {
		Ctx     context.Context
		Request apiv1.UpdateGatewayComponentRequest
	}
	UpdateGatewayComponentErr        error
	UpdateIngressComponentCalledWith struct {
		Ctx     context.Context
		Request apiv1.UpdateIngressComponentRequest
	}
	UpdateIngressComponentErr             error
	UpdateLoadBalancerComponentCalledWith struct {
		Ctx     context.Context
		Request apiv1.UpdateLoadBalancerComponentRequest
	}
	UpdateLoadBalancerComponentErr   error
	UpdateNetworkComponentCalledWith struct {
		Ctx     context.Context
		Request apiv1.UpdateNetworkComponentRequest
	}
	UpdateNetworkComponentErr        error
	UpdateStorageComponentCalledWith struct {
		Ctx     context.Context
		Request apiv1.UpdateStorageComponentRequest
	}
	UpdateStorageComponentErr error
}

func (c *Client) Bootstrap(ctx context.Context, bootstrapConfig apiv1.BootstrapConfig) (apiv1.ClusterMember, error) {
	c.BootstrapCalledWith = bootstrapConfig
	return c.BootstrapClusterMember, c.BootstrapErr
}

func (c *Client) IsBootstrapped(ctx context.Context) bool {
	return c.IsBootstrappedReturn
}

func (c *Client) CleanupNode(ctx context.Context, snap snap.Snap, nodeName string) {
	c.CleanupNodeCalledWith.Ctx = ctx
	c.CleanupNodeCalledWith.Snap = snap
	c.CleanupNodeCalledWith.NodeName = nodeName
}

func (c *Client) ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error) {
	return c.ClusterStatusReturn, c.ClusterStatusErr
}

func (c *Client) CreateJoinToken(ctx context.Context, name string, worker bool) (string, error) {
	return c.CreateJoinTokenReturn.Token, c.CreateJoinTokenReturn.Err
}

func (c *Client) GenerateAuthToken(ctx context.Context, username string, groups []string) (string, error) {
	return c.GenerateAuthTokenReturn.Token, c.GenerateAuthTokenReturn.Err
}

func (c *Client) JoinCluster(ctx context.Context, name string, address string, token string) error {
	c.JoinClusterCalledWith.Ctx = ctx
	c.JoinClusterCalledWith.Name = name
	c.JoinClusterCalledWith.Address = address
	c.JoinClusterCalledWith.Token = token
	return c.JoinClusterErr
}

func (c *Client) KubeConfig(ctx context.Context) (string, error) {
	return c.KubeConfigReturn, c.KubeConfigErr
}

func (c *Client) ListComponents(ctx context.Context) ([]apiv1.Component, error) {
	return c.ListComponentsReturn, c.ListComponentsErr
}

func (c *Client) RemoveNode(ctx context.Context, name string, force bool) error {
	c.RemoveNodeCalledWith.Ctx = ctx
	c.RemoveNodeCalledWith.Name = name
	c.RemoveNodeCalledWith.Force = force
	return c.RemoveNodeErr
}

func (c *Client) UpdateDNSComponent(ctx context.Context, request apiv1.UpdateDNSComponentRequest) error {
	c.UpdateDNSComponentCalledWith.Ctx = ctx
	c.UpdateDNSComponentCalledWith.Request = request
	return c.UpdateDNSComponentErr
}

func (c *Client) UpdateGatewayComponent(ctx context.Context, request apiv1.UpdateGatewayComponentRequest) error {
	c.UpdateGatewayComponentCalledWith.Ctx = ctx
	c.UpdateGatewayComponentCalledWith.Request = request
	return c.UpdateGatewayComponentErr
}

func (c *Client) UpdateIngressComponent(ctx context.Context, request apiv1.UpdateIngressComponentRequest) error {
	c.UpdateIngressComponentCalledWith.Ctx = ctx
	c.UpdateIngressComponentCalledWith.Request = request
	return c.UpdateIngressComponentErr
}

func (c *Client) UpdateLoadBalancerComponent(ctx context.Context, request apiv1.UpdateLoadBalancerComponentRequest) error {
	c.UpdateLoadBalancerComponentCalledWith.Ctx = ctx
	c.UpdateLoadBalancerComponentCalledWith.Request = request
	return c.UpdateLoadBalancerComponentErr
}

func (c *Client) UpdateNetworkComponent(ctx context.Context, request apiv1.UpdateNetworkComponentRequest) error {
	c.UpdateNetworkComponentCalledWith.Ctx = ctx
	c.UpdateNetworkComponentCalledWith.Request = request
	return c.UpdateNetworkComponentErr
}

func (c *Client) UpdateStorageComponent(ctx context.Context, request apiv1.UpdateStorageComponentRequest) error {
	c.UpdateStorageComponentCalledWith.Ctx = ctx
	c.UpdateStorageComponentCalledWith.Request = request
	return c.UpdateStorageComponentErr
}
