### bootstrap-certificates
**Type:** `string`<br>
**Default Value:** `self-signed`

The certificate authority to use for the cluster. This cannot be changed
after deployment. Allowed values are "self-signed" and "external". If
"external" is chosen, the charm should be integrated with an external
certificate authority charm.

### bootstrap-datastore
**Type:** `string`<br>
**Default Value:** `dqlite`

The datastore to use in Canonical Kubernetes. This cannot be changed
after deployment. Allowed values are "dqlite" and "etcd". If "etcd" is
chosen, the charm should be integrated with the etcd charm.

### bootstrap-node-taints
**Type:** `string`<br>
**Default Value:** ``

Space-separated list of taints to apply to this node at registration time.

This config is only used at bootstrap time when Kubelet first registers the
node with Kubernetes. To change node taints after deploy time, use kubectl
instead.

For more information, see the upstream Kubernetes documentation about
taints:
https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/

### bootstrap-pod-cidr
**Type:** `string`<br>
**Default Value:** `10.1.0.0/16`

Comma-separated CIDR blocks for IP addresses that can be assigned
to pods within the cluster. Can contain at most 2 blocks, one for IPv4
and one for IPv6.

After deployment it is not possible to change the size of
the IP range.

Examples:
  - "192.0.2.0/24"
  - "2001:db8::/32"
  - "192.0.2.0/24,2001:db8::/32"
  - "2001:db8::/32,192.0.2.0/24"

### bootstrap-service-cidr
**Type:** `string`<br>
**Default Value:** `10.152.183.0/24`

Comma-separated CIDR blocks for IP addresses that can be assigned
to services within the cluster. Can contain at most 2 blocks, one for IPv4
and one for IPv6.

After deployment it is not possible to change the size of
the IP range.

Examples:
  - "192.0.2.0/24"
  - "2001:db8::/32"
  - "192.0.2.0/24,2001:db8::/32"
  - "2001:db8::/32,192.0.2.0/24"

### cluster-annotations
**Type:** `string`<br>
**Default Value:** ``

Space-separated list of (key/value) pairs) that can be
used to add arbitrary metadata configuration to the Canonical
Kubernetes cluster. For more information, see the upstream Canonical
Kubernetes documentation about annotations:

https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/reference/annotations/

Example:
  e.g.: key1=value1 key2=value2

### containerd-custom-registries
**Type:** `string`<br>
**Default Value:** `[]`

Registry endpoints and credentials. Setting this config allows containerd
to pull images from registries where auth is required.

The value for this config must be a JSON array of credential objects, like this:
  e.g.: [{"url": "https://registry.example.com", "host": "my.registry:port", "username": "user", "password": "pass"}]

Credential Object Parameters:
url: REQUIRED str
  the URL to the registry, include the port if not it isn't implied from the schema.
    e.g: "url": "https://my.registry:8443"
    e.g: "url": "http://my.registry"

host: OPTIONAL str - defaults to auto-generated from the url
  could be registry host address or a name
    e.g.: myregistry.io:9000, 10.10.10.10:5432
    e.g.: myregistry.io, myregistry
  Note: It will be derived from `url` if not provided.
    e.g.: "url": "http://10.10.10.10:8000" --> "host": "10.10.10.10:8000"

username: OPTIONAL str - default ''
password: OPTIONAL str - default ''
identitytoken: OPTIONAL str - default ''
  Used by containerd for basic authentication to the registry.

ca_file: OPTIONAL str - default ''
cert_file: OPTIONAL str - default ''
key_file: OPTIONAL str - default ''
  For ssl/tls communication these should be a base64 encoded file
  e.g.:  "ca_file": "'"$(base64 -w 0 < my.custom.registry.pem)"'"

skip_verify: OPTIONAL bool - default false
  For situations where the registry has self-signed or expired certs and a quick work-around is necessary.
  e.g.: "skip_verify": true

Example config:
juju config k8s containerd_custom_registries='[{
    "url": "https://registry.example.com",
    "host": "ghcr.io",
    "ca_file": "'"$(base64 -w 0 < ~/my.custom.ca.pem)"'",
    "cert_file": "'"$(base64 -w 0 < ~/my.custom.cert.pem)"'",
    "key_file": "'"$(base64 -w 0 < ~/my.custom.key.pem)"'",
}]'

### dns-cluster-domain
**Type:** `string`<br>
**Default Value:** `cluster.local`

Sets the local domain of the cluster

### dns-enabled
**Type:** `boolean`<br>
**Default Value:** `True`

Enable/Disable the DNS feature on the cluster.

### dns-service-ip
**Type:** `string`<br>
**Default Value:** ``

Sets the IP address of the dns service. If omitted defaults to the IP address
of the Kubernetes service created by the feature.

Can be used to point to an external dns server when feature is disabled.

### dns-upstream-nameservers
**Type:** `string`<br>
**Default Value:** ``

Space-separated list of upstream nameservers used to forward queries for out-of-cluster
endpoints.

If omitted defaults to `/etc/resolv.conf` and uses the nameservers on each node.

### gateway-enabled
**Type:** `boolean`<br>
**Default Value:** `False`

Enable/Disable the gateway feature on the cluster.

### ingress-enable-proxy-protocol
**Type:** `boolean`<br>
**Default Value:** `False`

Determines if the proxy protocol should be enabled for ingresses.

### ingress-enabled
**Type:** `boolean`<br>
**Default Value:** `False`

Determines if the ingress feature should be enabled.

### kube-apiserver-extra-args
**Type:** `string`<br>
**Default Value:** ``

Space separated list of flags and key=value pairs that will be passed as arguments to
kube-apiserver.

Notes:
  Options may only be set on charm deployment

For example a value like this:
  runtime-config=batch/v2alpha1=true profiling=true
will result in kube-apiserver being run with the following options:
  --runtime-config=batch/v2alpha1=true --profiling=true

### kube-apiserver-extra-sans
**Type:** `string`<br>
**Default Value:** ``

Space separated list of extra Subject Alternative Names for the kube-apiserver
self-signed certificates.

Examples:
  - "kubernetes"
  - "kubernetes.default.svc"
  - "kubernetes.default.svc.cluster.local"

### kube-controller-manager-extra-args
**Type:** `string`<br>
**Default Value:** ``

Space separated list of flags and key=value pairs that will be passed as arguments to
kube-controller-manager.

Notes:
  Options may only be set on charm deployment
  cluster-name: cannot be overridden

For example a value like this:
  runtime-config=batch/v2alpha1=true profiling=true
will result in kube-controller-manager being run with the following options:
  --runtime-config=batch/v2alpha1=true --profiling=true

### kube-proxy-extra-args
**Type:** `string`<br>
**Default Value:** ``

Space separated list of flags and key=value pairs that will be passed as arguments to
kube-proxy.

Notes:
  Options may only be set on charm deployment

For example a value like this:
  runtime-config=batch/v2alpha1=true profiling=true
will result in kube-proxy being run with the following options:
  --runtime-config=batch/v2alpha1=true --profiling=true

### kube-scheduler-extra-args
**Type:** `string`<br>
**Default Value:** ``

Space separated list of flags and key=value pairs that will be passed as arguments to
kube-scheduler.

Notes:
  Options may only be set on charm deployment

For example a value like this:
  runtime-config=batch/v2alpha1=true profiling=true
will result in kube-scheduler being run with the following options:
  --runtime-config=batch/v2alpha1=true --profiling=true

### kubelet-extra-args
**Type:** `string`<br>
**Default Value:** ``

Space separated list of flags and key=value pairs that will be passed as arguments to
kubelet.

Notes:
  Options may only be set on charm deployment

For example a value like this:
  runtime-config=batch/v2alpha1=true profiling=true
will result in kubelet being run with the following options:
  --runtime-config=batch/v2alpha1=true --profiling=true

### load-balancer-bgp-local-asn
**Type:** `int`<br>
**Default Value:** `64512`

Local ASN for the load balancer. This is only used if load-balancer-bgp-mode
is set to true.

### load-balancer-bgp-mode
**Type:** `boolean`<br>
**Default Value:** `False`

Enable/Disable BGP mode for the load balancer. This is only used if
load-balancer-enabled is set to true.

### load-balancer-bgp-peer-address
**Type:** `string`<br>
**Default Value:** ``

Address of the BGP peer for the load balancer. This is only used if
load-balancer-bgp-mode is set to true.

### load-balancer-bgp-peer-port
**Type:** `int`<br>
**Default Value:** `179`

Port of the BGP peer for the load balancer. This is only used if
load-balancer-bgp-mode is set to true.

### load-balancer-cidrs
**Type:** `string`<br>
**Default Value:** ``

Space-separated list of CIDRs to use for the load balancer. This is
only used if load-balancer-enabled is set to true.

### load-balancer-enabled
**Type:** `boolean`<br>
**Default Value:** `False`

Enable/Disable the load balancer feature on the cluster.

### load-balancer-l2-interfaces
**Type:** `string`<br>
**Default Value:** ``

Space-separated list of interfaces to use for the load balancer. This
is only used if load-balancer-l2-mode is set to true. if unset, all
interfaces will be used.

### load-balancer-l2-mode
**Type:** `boolean`<br>
**Default Value:** `False`

Enable/Disable L2 mode for the load balancer. This is only used if
load-balancer-enabled is set to true.

### local-storage-enabled
**Type:** `boolean`<br>
**Default Value:** `True`

Enable local storage provisioning. This will create a storage class
named "local-storage" that uses the hostPath provisioner. This is
useful for development and testing purposes. It is not recommended for
production use.

### local-storage-local-path
**Type:** `string`<br>
**Default Value:** `/var/snap/k8s/common/rawfile-storage`

The path on the host where local storage will be provisioned. This
path must be writable by the kubelet. This is only used if
local-storage.enabled is set to true.

### local-storage-reclaim-policy
**Type:** `string`<br>
**Default Value:** `Delete`

The reclaim policy for local storage. This can be either "Delete" or
"Retain". If set to "Delete", the storage will be deleted when the
PersistentVolumeClaim is deleted. If set to "Retain", the storage will
be retained when the PersistentVolumeClaim is deleted.

### metrics-server-enabled
**Type:** `boolean`<br>
**Default Value:** `True`

Enable/Disable the metrics-server feature on the cluster.

### network-enabled
**Type:** `boolean`<br>
**Default Value:** `True`

Enables or disables the network feature.

### node-labels
**Type:** `string`<br>
**Default Value:** ``

Labels can be used to organize and to select subsets of nodes in the
cluster. Declare node labels in key=value format, separated by spaces.

