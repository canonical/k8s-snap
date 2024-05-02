# Bootstrap configuration file reference

A YAML file can be supplied to the `k8s bootstrap` command to configure and customize the cluster.
This Reference section provides the format of this file by listing all available options and their details.

## Format Specification

### cluster-config.network

**Type:** `object` <br>
**Required:** `No`

Configuration options for the network feature (See [network configuration])

### cluster-config.dns

**Type:** `object` <br>
**Required:** `No`

Configuration options for the dns feature (See [dns configuration])

### cluster-config.ingress

**Type:** `object` <br>
**Required:** `No`

Configuration options for the ingress feature (See [ingress configuration])

### cluster-config.load-balancer

**Type:** `object` <br>
**Required:** `No`

Configuration options for the load-balancer feature (See [load-balancer configuration])

### cluster-config.local-storage

**Type:** `object` <br>
**Required:** `No`

Configuration options for the local-storage feature (See [local-storage configuration])

### cluster-config.gateway

**Type:** `object` <br>
**Required:** `No`

Configuration options for the gateway feature (See [gateway configuration])

### cluster-config.cloud-provider

**Type:** `string` <br>
**Required:** `No` <br>
**Possible Values:** `external`

Sets the cloud provider to be used by the cluster.

When this is set as `external`, node will wait for an external cloud provider to do cloud specific setup and finish node initialization.

### control-plane-taints

**Type:** `list[string]` <br>
**Required:** `No`

List of taints to be applied to control plane nodes.

### pod-cidr

**Type:** `string` <br>
**Required:** `No`

The CIDR to be used for assigning pod addresses. If omitted defaults to `10.1.0.0/16`

### service-cidr

**Type:** `string` <br>
**Required:** `No`

The CIDR to be used for assigning service addresses. If omitted defaults to `10.152.183.0/24`

### disable-rbac

**Type:** `bool` <br>
**Required:** `No`

Determines if RBAC should be disabled. If omitted defaults to `false`

### secure-port

**Type:** `int` <br>
**Required:** `No`

The port number for kube-apiserver to use. If omitted defaults to `6443`

### k8s-dqlite-port

**Type:** `int` <br>
**Required:** `No`

The port number for k8s-dqlite to use. If omitted defaults to `9000`

### datastore-type

**Type:** `string` <br>
**Required:** `No` <br>
**Possible Values:** `k8s-dqlite | external`

The type of datastore to be used. If omitted defaults to `k8s-dqlite`

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

The CA certificate to be used for Kubernetes services. If omitted defaults to an auto generated certificate.

### ca-key

**Type:** `string` <br>
**Required:** `No` <br>

The CA key to be used for Kubernetes services. If omitted defaults to an auto generated key.

### front-proxy-ca-crt

**Type:** `string` <br>
**Required:** `No` <br>

The CA certificate to be used for the front proxy. If omitted defaults to an auto generated certificate.

### front-proxy-ca-key

**Type:** `string` <br>
**Required:** `No` <br>

The CA key to be used for the front proxy. If omitted defaults to an auto generated key.

### front-proxy-client-crt

**Type:** `string` <br>
**Required:** `No` <br>

The client certificate to be used for the front proxy. If omitted defaults to an auto generated certificate.

### front-proxy-client-key

**Type:** `string` <br>
**Required:** `No` <br>

The client key to be used for the front proxy. If omitted defaults to an auto generated key.


### apiserver-kubelet-client-crt

**Type:** `string` <br>
**Required:** `No` <br>

The client certificate to be used by kubelet for communicating with the kube-apiserver. If omitted defaults to an auto generated certificate.

### apiserver-kubelet-client-key

**Type:** `string` <br>
**Required:** `No` <br>

The client key to be used by kubelet for communicating with the kube-apiserver. If omitted defaults to an auto generated key.

### service-account-key

**Type:** `string` <br>
**Required:** `No` <br>

The key to be used by the default service account. If omitted defaults to an auto generated key.

### apiserver-crt

**Type:** `string` <br>
**Required:** `No` <br>

The certificate to be used for the kube-apiserver. If omitted defaults to an auto generated certificate.

### apiserver-key

**Type:** `string` <br>
**Required:** `No` <br>

The key to be used for the kube-apiserver. If omitted defaults to an auto generated key.

### kubelet-crt

**Type:** `string` <br>
**Required:** `No` <br>

The certificate to be used for the kubelet. If omitted defaults to an auto generated certificate.

### kubelet-key

**Type:** `string` <br>
**Required:** `No` <br>

The key to be used for the kubelet. If omitted defaults to an auto generated key.


<!--LINKS -->
[network configuration]: ./network-configuration
[dns configuration]: ./dns-configuration
[ingress configuration]: ./ingress-configuration
[load-balancer configuration]: ./load-balancer-configuration
[local-storage configuration]: ./local-storage-configuration
[gateway configuration]: ./gateway-configuration
