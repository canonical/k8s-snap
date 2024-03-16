package newtypes

import (
	"fmt"
)

// MergeClusterConfig applies updates from non-empty values of the new ClusterConfig to an existing one.
// MergeClusterConfig will return an error if we try to update a config that must not be updated. once such an operation is implemented in the future, we can allow the change here.
// MergeClusterConfig will create a new ClusterConfig object to avoid mutating the existing config objects.
func MergeClusterConfig(existing ClusterConfig, new ClusterConfig) (ClusterConfig, error) {
	var (
		config ClusterConfig
		err    error
	)

	// update string fields
	for _, i := range []struct {
		name        string
		val         *string
		old         *string
		new         *string
		allowChange bool
	}{
		// certificates
		{name: "CA certificate", val: config.Certificates.CACert, old: existing.Certificates.CACert, new: new.Certificates.CACert},
		{name: "CA key", val: config.Certificates.CAKey, old: existing.Certificates.CAKey, new: new.Certificates.CAKey},
		{name: "apiserver-kubelet-client certificate", val: config.Certificates.APIServerKubeletClientCert, old: existing.Certificates.APIServerKubeletClientCert, new: new.Certificates.APIServerKubeletClientCert, allowChange: true},
		{name: "apiserver-kubelet-client key", val: config.Certificates.APIServerKubeletClientKey, old: existing.Certificates.APIServerKubeletClientKey, new: new.Certificates.APIServerKubeletClientKey, allowChange: true},
		{name: "front proxy CA certificate", val: config.Certificates.FrontProxyCACert, old: existing.Certificates.FrontProxyCACert, new: new.Certificates.FrontProxyCACert, allowChange: true},
		{name: "front proxy CA key", val: config.Certificates.FrontProxyCAKey, old: existing.Certificates.FrontProxyCAKey, new: new.Certificates.FrontProxyCAKey, allowChange: true},
		{name: "service account key", val: config.Certificates.ServiceAccountKey, old: existing.Certificates.ServiceAccountKey, new: new.Certificates.ServiceAccountKey},
		// datastore
		{name: "datastore type", val: config.Datastore.Type, old: existing.Datastore.Type, new: new.Datastore.Type},
		{name: "k8s-dqlite certificate", val: config.Datastore.K8sDqliteCert, old: existing.Datastore.K8sDqliteCert, new: new.Datastore.K8sDqliteCert},
		{name: "k8s-dqlite key", val: config.Datastore.K8sDqliteKey, old: existing.Datastore.K8sDqliteKey, new: new.Datastore.K8sDqliteKey},
		{name: "external datastore URL", val: config.Datastore.ExternalURL, old: existing.Datastore.ExternalURL, new: new.Datastore.ExternalURL, allowChange: true},
		{name: "external datastore CA certificate", val: config.Datastore.ExternalCACert, old: existing.Datastore.ExternalCACert, new: new.Datastore.ExternalCACert, allowChange: true},
		{name: "external datastore client certificate", val: config.Datastore.ExternalClientCert, old: existing.Datastore.ExternalClientCert, new: new.Datastore.ExternalClientCert, allowChange: true},
		{name: "external datastore client key", val: config.Datastore.ExternalClientKey, old: existing.Datastore.ExternalClientKey, new: new.Datastore.ExternalClientKey, allowChange: true},
		// network
		{name: "pod CIDR", val: config.Network.PodCIDR, old: existing.Network.PodCIDR, new: new.Network.PodCIDR},
		{name: "service CIDR", val: config.Network.ServiceCIDR, old: existing.Network.ServiceCIDR, new: new.Network.ServiceCIDR},
		// apiserver
		{name: "kube-apiserver authorization mode", val: config.APIServer.AuthorizationMode, old: existing.APIServer.AuthorizationMode, new: new.APIServer.AuthorizationMode, allowChange: true},
		// ingress
		{name: "ingress default TLS secret", val: config.Features.Ingress.DefaultTLSSecret, old: existing.Features.Ingress.DefaultTLSSecret, new: new.Features.Ingress.DefaultTLSSecret, allowChange: true},
		// load balancer
		{name: "load balancer BGP peer address", val: config.Features.LoadBalancer.BGPPeerAddress, old: existing.Features.LoadBalancer.BGPPeerAddress, new: new.Features.LoadBalancer.BGPPeerAddress, allowChange: true},
		// local storage
		{name: "local storage path", val: config.Features.LocalStorage.LocalPath, old: existing.Features.LocalStorage.LocalPath, new: new.Features.LocalStorage.LocalPath},
		{name: "local storage reclaim policy", val: config.Features.LocalStorage.ReclaimPolicy, old: existing.Features.LocalStorage.ReclaimPolicy, new: new.Features.LocalStorage.ReclaimPolicy, allowChange: true},
	} {
		if i.val, err = mergeField(i.old, i.new, i.allowChange); err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	// update string slice fields
	for _, i := range []struct {
		name        string
		val         *[]string
		old         *[]string
		new         *[]string
		allowChange bool
	}{
		{name: "DNS upstream nameservers", val: config.Features.DNS.UpstreamNameservers, old: existing.Features.DNS.UpstreamNameservers, new: new.Features.DNS.UpstreamNameservers, allowChange: true},
		{name: "load balancer CIDRs", val: config.Features.LoadBalancer.CIDRs, old: existing.Features.LoadBalancer.CIDRs, new: new.Features.LoadBalancer.CIDRs, allowChange: true},
		{name: "load balancer L2 interfaces", val: config.Features.LoadBalancer.L2Interfaces, old: existing.Features.DNS.UpstreamNameservers, new: new.Features.LoadBalancer.L2Interfaces, allowChange: true},
	} {
		if i.val, err = mergeSliceField(i.old, i.new, i.allowChange); err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	// update int fields
	for _, i := range []struct {
		name        string
		val         *int
		old         *int
		new         *int
		allowChange bool
	}{
		// apiserver
		{name: "kube-apiserver secure port", val: config.APIServer.SecurePort, old: existing.APIServer.SecurePort, new: new.APIServer.SecurePort},
		// datastore
		{name: "k8s-dqlite port", val: config.Datastore.K8sDqlitePort, old: existing.Datastore.K8sDqlitePort, new: new.Datastore.K8sDqlitePort},
		// load-balancer
		{name: "load balancer BGP local ASN", val: config.Features.LoadBalancer.BGPLocalASN, old: existing.Features.LoadBalancer.BGPLocalASN, new: new.Features.LoadBalancer.BGPLocalASN, allowChange: true},
		{name: "load balancer BGP peer ASN", val: config.Features.LoadBalancer.BGPPeerASN, old: existing.Features.LoadBalancer.BGPPeerASN, new: new.Features.LoadBalancer.BGPPeerASN, allowChange: true},
		{name: "load balancer BGP peer port", val: config.Features.LoadBalancer.BGPPeerPort, old: existing.Features.LoadBalancer.BGPPeerPort, new: new.Features.LoadBalancer.BGPPeerPort, allowChange: true},
	} {
		if i.val, err = mergeField(i.old, i.new, i.allowChange); err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	// update bool fields
	for _, i := range []struct {
		name        string
		val         *bool
		old         *bool
		new         *bool
		allowChange bool
	}{
		// network
		{name: "network enabled", val: config.Features.Network.Enabled, old: existing.Features.Network.Enabled, new: new.Features.Network.Enabled, allowChange: true},
		// DNS
		{name: "DNS enabled", val: config.Features.DNS.Enabled, old: existing.Features.DNS.Enabled, new: new.Features.DNS.Enabled, allowChange: true},
		// gateway
		{name: "gateway enabled", val: config.Features.Gateway.Enabled, old: existing.Features.Gateway.Enabled, new: new.Features.Gateway.Enabled, allowChange: true},
		// ingress
		{name: "ingress enabled", val: config.Features.Ingress.Enabled, old: existing.Features.Ingress.Enabled, new: new.Features.Ingress.Enabled, allowChange: true},
		{name: "ingress enable proxy protocol", val: config.Features.Ingress.EnableProxyProtocol, old: existing.Features.Ingress.EnableProxyProtocol, new: new.Features.Ingress.EnableProxyProtocol, allowChange: true},
		// load-balancer
		{name: "load balancer enabled", val: config.Features.LoadBalancer.Enabled, old: existing.Features.LoadBalancer.Enabled, new: new.Features.LoadBalancer.Enabled, allowChange: true},
		{name: "load balancer L2 mode", val: config.Features.LoadBalancer.L2Mode, old: existing.Features.LoadBalancer.L2Mode, new: new.Features.LoadBalancer.L2Mode, allowChange: true},
		{name: "load-balancer BGP mode", val: config.Features.LoadBalancer.BGPMode, old: existing.Features.LoadBalancer.BGPMode, new: new.Features.LoadBalancer.BGPMode, allowChange: true},
		// local-storage
		{name: "local storage enabled", val: config.Features.LocalStorage.Enabled, old: existing.Features.LocalStorage.Enabled, new: new.Features.LocalStorage.Enabled, allowChange: true},
		{name: "local storage set default", val: config.Features.LocalStorage.SetDefault, old: existing.Features.LocalStorage.SetDefault, new: new.Features.LocalStorage.SetDefault, allowChange: true},
		// metrics-server
		{name: "metrics server enabled", val: config.Features.MetricsServer.Enabled, old: existing.Features.MetricsServer.Enabled, new: new.Features.MetricsServer.Enabled, allowChange: true},
	} {
		if i.val, err = mergeField(i.old, i.new, i.allowChange); err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	return config, nil
}
