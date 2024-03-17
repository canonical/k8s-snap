package types

import "github.com/canonical/k8s/pkg/utils/vals"

func (c *ClusterConfig) SetDefaults() {
	// networking
	if c.Network.GetPodCIDR() == "" {
		c.Network.PodCIDR = vals.Pointer("10.1.0.0/16")
	}
	if c.Network.GetServiceCIDR() == "" {
		c.Network.ServiceCIDR = vals.Pointer("10.152.183.0/24")
	}
	// kube-apiserver
	if c.APIServer.GetSecurePort() == 0 {
		c.APIServer.SecurePort = vals.Pointer(6443)
	}
	if c.APIServer.GetAuthorizationMode() == "" {
		c.APIServer.AuthorizationMode = vals.Pointer("Node,RBAC")
	}
	// datastore
	if c.Datastore.GetType() == "" {
		c.Datastore.Type = vals.Pointer("k8s-dqlite")
	}
	if c.Datastore.GetK8sDqlitePort() == 0 {
		c.Datastore.K8sDqlitePort = vals.Pointer(9000)
	}
	// kubelet
	if c.Kubelet.GetClusterDomain() == "" {
		c.Kubelet.ClusterDomain = vals.Pointer("cluster.local")
	}
	// features
	if c.Features.Network.Enabled == nil {
		c.Features.Network.Enabled = vals.Pointer(false)
	}
	if c.Features.DNS.Enabled == nil {
		c.Features.DNS.Enabled = vals.Pointer(false)
	}
	if len(c.Features.DNS.GetUpstreamNameservers()) == 0 {
		c.Features.DNS.UpstreamNameservers = vals.Pointer([]string{"/etc/resolv.conf"})
	}
	if c.Features.LocalStorage.Enabled == nil {
		c.Features.LocalStorage.Enabled = vals.Pointer(false)
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
	if c.Features.LoadBalancer.Enabled == nil {
		c.Features.LoadBalancer.Enabled = vals.Pointer(false)
	}
	if c.Features.LoadBalancer.L2Mode == nil {
		c.Features.LoadBalancer.L2Mode = vals.Pointer(false)
	}
	if c.Features.LoadBalancer.BGPMode == nil {
		c.Features.LoadBalancer.BGPMode = vals.Pointer(false)
	}
	if c.Features.Ingress.Enabled == nil {
		c.Features.Ingress.Enabled = vals.Pointer(false)
	}
	if c.Features.Ingress.EnableProxyProtocol == nil {
		c.Features.Ingress.EnableProxyProtocol = vals.Pointer(false)
	}
	if c.Features.Gateway.Enabled == nil {
		c.Features.Gateway.Enabled = vals.Pointer(false)
	}
	if c.Features.MetricsServer.Enabled == nil {
		c.Features.MetricsServer.Enabled = vals.Pointer(false)
	}
}
