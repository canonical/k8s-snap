package types

import "github.com/canonical/k8s/pkg/utils/vals"

func (c *ClusterConfig) SetDefaults() {
	// networking
	if c.Network.Enabled == nil {
		c.Network.Enabled = vals.Pointer(false)
	}
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
	// dns
	if c.DNS.Enabled == nil {
		c.DNS.Enabled = vals.Pointer(false)
	}
	if len(c.DNS.GetUpstreamNameservers()) == 0 {
		c.DNS.UpstreamNameservers = vals.Pointer([]string{"/etc/resolv.conf"})
	}
	// local storage
	if c.LocalStorage.Enabled == nil {
		c.LocalStorage.Enabled = vals.Pointer(false)
	}
	if c.LocalStorage.GetLocalPath() == "" {
		c.LocalStorage.LocalPath = vals.Pointer("/var/snap/k8s/common/rawfile-storage")
	}
	if c.LocalStorage.GetReclaimPolicy() == "" {
		c.LocalStorage.ReclaimPolicy = vals.Pointer("Delete")
	}
	if c.LocalStorage.SetDefault == nil {
		c.LocalStorage.SetDefault = vals.Pointer(true)
	}
	// load balancer
	if c.LoadBalancer.Enabled == nil {
		c.LoadBalancer.Enabled = vals.Pointer(false)
	}
	if c.LoadBalancer.CIDRs == nil {
		c.LoadBalancer.CIDRs = vals.Pointer([]string{})
	}
	if c.LoadBalancer.L2Mode == nil {
		c.LoadBalancer.L2Mode = vals.Pointer(false)
	}
	if c.LoadBalancer.L2Interfaces == nil {
		c.LoadBalancer.L2Interfaces = vals.Pointer([]string{})
	}
	if c.LoadBalancer.BGPMode == nil {
		c.LoadBalancer.BGPMode = vals.Pointer(false)
	}
	if c.LoadBalancer.BGPLocalASN == nil {
		c.LoadBalancer.BGPLocalASN = vals.Pointer(0)
	}
	if c.LoadBalancer.BGPPeerAddress == nil {
		c.LoadBalancer.BGPPeerAddress = vals.Pointer("")
	}
	if c.LoadBalancer.BGPPeerASN == nil {
		c.LoadBalancer.BGPPeerASN = vals.Pointer(0)
	}
	if c.LoadBalancer.BGPPeerPort == nil {
		c.LoadBalancer.BGPPeerPort = vals.Pointer(0)
	}
	// ingress
	if c.Ingress.Enabled == nil {
		c.Ingress.Enabled = vals.Pointer(false)
	}
	if c.Ingress.DefaultTLSSecret == nil {
		c.Ingress.DefaultTLSSecret = vals.Pointer("")
	}
	if c.Ingress.EnableProxyProtocol == nil {
		c.Ingress.EnableProxyProtocol = vals.Pointer(false)
	}
	// gateway
	if c.Gateway.Enabled == nil {
		c.Gateway.Enabled = vals.Pointer(false)
	}
	// metrics server
	if c.MetricsServer.Enabled == nil {
		c.MetricsServer.Enabled = vals.Pointer(false)
	}
}
