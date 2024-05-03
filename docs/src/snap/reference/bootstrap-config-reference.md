# Bootstrap configuration file reference

A YAML file can be supplied to the `k8s bootstrap` command to configure and customize the cluster.
This Reference section provides the format of this file by listing all available options and their details.

## Format Specification

### cluster-config.network
**Type:** `object` <br>
**Required:** `No`

Configuration options for the network feature

#### cluster-config.network.enabled
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the feature should be enabled.
If omitted defaults to `true`

### cluster-config.dns
**Type:** `object` <br>
**Required:** `No`

Configuration options for the dns feature

#### cluster-config.dns.enabled
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the feature should be enabled.
If omitted defaults to `true`

#### cluster-config.dns.cluster-domain
**Type:** `string`<br>
**Required:** `No` <br>

Sets the local domain of the cluster.
If omitted defaults to `cluster.local`

#### cluster-config.dns.service-ip
**Type:** `string`<br>
**Required:** `No` <br>

Sets the IP address of the dns service.
If omitted defaults to the IP address of the Kubernetes service created by the feature.

Can be used to point to an external dns server when feature is disabled.


#### cluster-config.dns.upstream-nameservers
**Type:** `list[string]`<br>
**Required:** `No` <br>

Sets the upstream nameservers used to forward queries for out-of-cluster endpoints.
If omitted defaults to `/etc/resolv.conf` and uses the nameservers of the node.


### cluster-config.ingress
**Type:** `object` <br>
**Required:** `No`

Configuration options for the ingress feature

#### cluster-config.ingress.enabled
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the feature should be enabled.
If omitted defaults to `false`

#### cluster-config.ingress.default-tls-secret
**Type:** `string`<br>
**Required:** `No` <br>

Sets the name of the secret to be used for providing default encryption to ingresses.

Ingresses can specify another TLS secret in their resource definitions,
in which case the default secret won't be used.

#### cluster-config.ingress.enable-proxy-protocol
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the proxy protocol should be enabled for ingresses.
If omitted defaults to `false`


### cluster-config.load-balancer
**Type:** `object` <br>
**Required:** `No`

Configuration options for the load-balancer feature

#### cluster-config.load-balancer.enabled
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the feature should be enabled.
If omitted defaults to `false`

#### cluster-config.load-balancer.cidrs
**Type:** `list[string]`<br>
**Required:** `No` <br>

Sets the CIDRs used for assigning IP addresses to Kubernetes services with type `LoadBalancer`.

#### cluster-config.load-balancer.l2-mode
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if L2 mode should be enabled.
If omitted defaults to `false`

#### cluster-config.load-balancer.l2-interfaces
**Type:** `list[string]`<br>
**Required:** `No` <br>

Sets the interfaces to be used for announcing IP addresses through ARP.
If omitted all interfaces will be used.

#### cluster-config.load-balancer.bgp-mode
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if BGP mode should be enabled.
If omitted defaults to `false`

#### cluster-config.load-balancer.bgp-local-asn
**Type:** `int`<br>
**Required:** `Yes if bgp-mode is true` <br>

Sets the ASN to be used for the local virtual BGP router.

#### cluster-config.load-balancer.bgp-peer-address
**Type:** `string`<br>
**Required:** `Yes if bgp-mode is true` <br>

Sets the IP address of the BGP peer.

#### cluster-config.load-balancer.bgp-peer-asn
**Type:** `int`<br>
**Required:** `Yes if bgp-mode is true` <br>

Sets the ASN of the BGP peer.

#### cluster-config.load-balancer.bgp-peer-port
**Type:** `int`<br>
**Required:** `Yes if bgp-mode is true` <br>

Sets the port of the BGP peer.


### cluster-config.local-storage
**Type:** `object` <br>
**Required:** `No`

Configuration options for the local-storage feature

#### cluster-config.local-storage.enabled
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the feature should be enabled.
If omitted defaults to `false`

#### cluster-config.local-storage.local-path
**Type:** `string`<br>
**Required:** `No` <br>

Sets the path to be used for storing volume data.
If omitted defaults to `/var/snap/k8s/common/rawfile-storage`

#### cluster-config.local-storage.reclaim-policy
**Type:** `string`<br>
**Required:** `No` <br>
**Possible Values:** `Retain | Recycle | Delete`

Sets the reclaim policy of the storage class.
If omitted defaults to `Delete`

#### cluster-config.local-storage.default
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the storage class should be set as default.
If omitted defaults to `true`


### cluster-config.gateway
**Type:** `object` <br>
**Required:** `No`

Configuration options for the gateway feature

#### cluster-config.gateway.enabled
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the feature should be enabled.
If omitted defaults to `true`

### cluster-config.cloud-provider
**Type:** `string` <br>
**Required:** `No` <br>
**Possible Values:** `external`

Sets the cloud provider to be used by the cluster.

When this is set as `external`, node will wait for an external cloud provider to
do cloud specific setup and finish node initialization.

### control-plane-taints
**Type:** `list[string]` <br>
**Required:** `No`

List of taints to be applied to control plane nodes.

### pod-cidr
**Type:** `string` <br>
**Required:** `No`

The CIDR to be used for assigning pod addresses.
If omitted defaults to `10.1.0.0/16`

### service-cidr
**Type:** `string` <br>
**Required:** `No`

The CIDR to be used for assigning service addresses.
If omitted defaults to `10.152.183.0/24`

### disable-rbac
**Type:** `bool` <br>
**Required:** `No`

Determines if RBAC should be disabled.
If omitted defaults to `false`

### secure-port
**Type:** `int` <br>
**Required:** `No`

The port number for kube-apiserver to use.
If omitted defaults to `6443`

### k8s-dqlite-port
**Type:** `int` <br>
**Required:** `No`

The port number for k8s-dqlite to use.
If omitted defaults to `9000`

### datastore-type
**Type:** `string` <br>
**Required:** `No` <br>
**Possible Values:** `k8s-dqlite | external`

The type of datastore to be used.
If omitted defaults to `k8s-dqlite`

Can be used to point to an external datastore like etcd.

### datastore-servers
**Type:** `list[string]` <br>
**Required:** `No` <br>

The server addresses to be used when `datastore-type` is set to `external`.

### datastore-ca-crt
**Type:** `string` <br>
**Required:** `No` <br>

The CA certificate to be used when communicating with the external datastore.

### datastore-client-crt
**Type:** `string` <br>
**Required:** `No` <br>

The client certificate to be used when communicating with the external datastore.

### datastore-client-key
**Type:** `string` <br>
**Required:** `No` <br>

The client key to be used when communicating with the external datastore.

### extra-sans
**Type:** `list[string]` <br>
**Required:** `No` <br>

List of extra SANs to be added to certificates.

### ca-crt
**Type:** `string` <br>
**Required:** `No` <br>

The CA certificate to be used for Kubernetes services.
If omitted defaults to an auto generated certificate.

### ca-key
**Type:** `string` <br>
**Required:** `No` <br>

The CA key to be used for Kubernetes services.
If omitted defaults to an auto generated key.

### front-proxy-ca-crt
**Type:** `string` <br>
**Required:** `No` <br>

The CA certificate to be used for the front proxy.
If omitted defaults to an auto generated certificate.

### front-proxy-ca-key
**Type:** `string` <br>
**Required:** `No` <br>

The CA key to be used for the front proxy.
If omitted defaults to an auto generated key.

### front-proxy-client-crt
**Type:** `string` <br>
**Required:** `No` <br>

The client certificate to be used for the front proxy.
If omitted defaults to an auto generated certificate.

### front-proxy-client-key
**Type:** `string` <br>
**Required:** `No` <br>

The client key to be used for the front proxy.
If omitted defaults to an auto generated key.


### apiserver-kubelet-client-crt
**Type:** `string` <br>
**Required:** `No` <br>

The client certificate to be used by kubelet for communicating with the kube-apiserver.
If omitted defaults to an auto generated certificate.

### apiserver-kubelet-client-key
**Type:** `string` <br>
**Required:** `No` <br>

The client key to be used by kubelet for communicating with the kube-apiserver.
If omitted defaults to an auto generated key.

### service-account-key
**Type:** `string` <br>
**Required:** `No` <br>

The key to be used by the default service account.
If omitted defaults to an auto generated key.

### apiserver-crt
**Type:** `string` <br>
**Required:** `No` <br>

The certificate to be used for the kube-apiserver.
If omitted defaults to an auto generated certificate.

### apiserver-key
**Type:** `string` <br>
**Required:** `No` <br>

The key to be used for the kube-apiserver.
If omitted defaults to an auto generated key.

### kubelet-crt
**Type:** `string` <br>
**Required:** `No` <br>

The certificate to be used for the kubelet.
If omitted defaults to an auto generated certificate.

### kubelet-key
**Type:** `string` <br>
**Required:** `No` <br>

The key to be used for the kubelet.
If omitted defaults to an auto generated key.
