package newtypes

import "github.com/canonical/k8s/pkg/utils/vals"

func (c *ClusterConfig) SetDefaults() {
	if c.Network.GetPodCIDR() == "" {
		c.Network.PodCIDR = vals.Pointer("10.1.0.0/16")
	}
	if c.Network.GetServiceCIDR() == "" {
		c.Network.ServiceCIDR = vals.Pointer("10.152.183.0/24")
	}
	if c.APIServer.GetSecurePort() == 0 {
		c.APIServer.SecurePort = vals.Pointer(6443)
	}
	if c.APIServer.GetAuthorizationMode() == "" {
		c.APIServer.AuthorizationMode = vals.Pointer("Node,RBAC")
	}
	if c.Datastore.GetK8sDqlitePort() == 0 {
		c.Datastore.K8sDqlitePort = vals.Pointer(9000)
	}
	if len(c.Features.DNS.GetUpstreamNameservers()) == 0 {
		c.Features.DNS.UpstreamNameservers = vals.Pointer([]string{"/etc/resolv.conf"})
	}
	if c.Features.LocalStorage.GetLocalPath() == "" {
		c.Features.LocalStorage.LocalPath = vals.Pointer("/var/snap/k8s/common/rawfile-storage")
	}
	if c.Features.LocalStorage.GetReclaimPolicy() == "" {
		c.Features.LocalStorage.ReclaimPolicy = vals.Pointer("Delete")
	}
	if c.Features.LocalStorage.SetDefault == nil {
		c.Features.LocalStorage.SetDefault = vals.Pointer(true)
	}
	if c.Features.LoadBalancer.L2Mode == nil {
		c.Features.LoadBalancer.L2Mode = vals.Pointer(true)
	}
}
