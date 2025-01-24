package types

import (
	"github.com/canonical/k8s/pkg/utils"
)

func (c *ClusterConfig) SetDefaults() {
	// networking
	if c.Network.Enabled == nil {
		c.Network.Enabled = utils.Pointer(false)
	}
	if c.Network.GetPodCIDR() == "" {
		c.Network.PodCIDR = utils.Pointer("10.1.0.0/16")
	}
	if c.Network.GetServiceCIDR() == "" {
		c.Network.ServiceCIDR = utils.Pointer("10.152.183.0/24")
	}
	// kube-apiserver
	if c.APIServer.GetSecurePort() == 0 {
		c.APIServer.SecurePort = utils.Pointer(6443)
	}
	if c.APIServer.GetAuthorizationMode() == "" {
		c.APIServer.AuthorizationMode = utils.Pointer("Node,RBAC")
	}
	// datastore
	if c.Datastore.GetType() == "" {
		c.Datastore.Type = utils.Pointer("k8s-dqlite")
	}
	if c.Datastore.GetK8sDqlitePort() == 0 {
		c.Datastore.K8sDqlitePort = utils.Pointer(9000)
	}
	// kubelet
	if c.Kubelet.GetClusterDomain() == "" {
		c.Kubelet.ClusterDomain = utils.Pointer("cluster.local")
	}
	// dns
	if c.DNS.Enabled == nil {
		c.DNS.Enabled = utils.Pointer(false)
	}
	if len(c.DNS.GetUpstreamNameservers()) == 0 {
		c.DNS.UpstreamNameservers = utils.Pointer([]string{"/etc/resolv.conf"})
	}
	// local storage
	if c.LocalStorage.Enabled == nil {
		c.LocalStorage.Enabled = utils.Pointer(false)
	}
	if c.LocalStorage.GetLocalPath() == "" {
		c.LocalStorage.LocalPath = utils.Pointer("/var/snap/k8s/common/rawfile-storage")
	}
	if c.LocalStorage.GetReclaimPolicy() == "" {
		c.LocalStorage.ReclaimPolicy = utils.Pointer("Delete")
	}
	if c.LocalStorage.Default == nil {
		c.LocalStorage.Default = utils.Pointer(true)
	}
	// load balancer
	if c.LoadBalancer.Enabled == nil {
		c.LoadBalancer.Enabled = utils.Pointer(false)
	}
	if c.LoadBalancer.CIDRs == nil {
		c.LoadBalancer.CIDRs = utils.Pointer([]string{})
	}
	if c.LoadBalancer.L2Mode == nil {
		c.LoadBalancer.L2Mode = utils.Pointer(true)
	}
	if c.LoadBalancer.L2Interfaces == nil {
		c.LoadBalancer.L2Interfaces = utils.Pointer([]string{})
	}
	if c.LoadBalancer.BGPMode == nil {
		c.LoadBalancer.BGPMode = utils.Pointer(false)
	}
	if c.LoadBalancer.BGPLocalASN == nil {
		c.LoadBalancer.BGPLocalASN = utils.Pointer(0)
	}
	if c.LoadBalancer.BGPPeerAddress == nil {
		c.LoadBalancer.BGPPeerAddress = utils.Pointer("")
	}
	if c.LoadBalancer.BGPPeerASN == nil {
		c.LoadBalancer.BGPPeerASN = utils.Pointer(0)
	}
	if c.LoadBalancer.BGPPeerPort == nil {
		c.LoadBalancer.BGPPeerPort = utils.Pointer(0)
	}
	// ingress
	if c.Ingress.Enabled == nil {
		c.Ingress.Enabled = utils.Pointer(false)
	}
	if c.Ingress.DefaultTLSSecret == nil {
		c.Ingress.DefaultTLSSecret = utils.Pointer("")
	}
	if c.Ingress.EnableProxyProtocol == nil {
		c.Ingress.EnableProxyProtocol = utils.Pointer(false)
	}
	// gateway
	if c.Gateway.Enabled == nil {
		c.Gateway.Enabled = utils.Pointer(false)
	}
	// metrics server
	if c.MetricsServer.Enabled == nil {
		c.MetricsServer.Enabled = utils.Pointer(true)
	}
}
