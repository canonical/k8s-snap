# Services and ports

## Network Services

There are two main types of services based on the network interface they use:

* Default Host Interface Services: These services bind to the default host
interface, making them accessible from outside the host.
* Localhost Services: These services bind to the localhost interface,
meaning they can only be accessed from within the host.

### Services binding to the default Host interface

| Port  | Protocol | Service         | Description                                                                                                                   |
|-------|----------|-----------------|-------------------------------------------------------------------------------------------------------------------------------|
| 2379  | TCP      | etcd            | SSL encrypted client connection to etcd. Client certificate required.                                                         |
| 4244  | TCP      | cilium-agent    | Listening address for Hubble.                                                                                                 |
| 4240  | TCP      | cilium-agent    | TCP port for cluster-wide network connectivity and Cilium agent health API.                                                   |
| 6400  | TCP      | k8sd            | Default REST API port for Canonical Kubernetes daemon.                                                                        |
| 6443  | TCP      | kube-apiserver  | Kubernetes API server. SSL encrypted. Clients must present a valid password from a Static Password File.                      |
| 8472  | UDP      | cilium-agent    | Default VXLAN port used by Cilium.                                                                                            |
| 9000  | TCP      | k8s-dqlite      | SSL encrypted connection for k8s-dqlite. Client certificates required. Only applies if datastore type is set to `k8s-dqlite`. |
| 9963  | TCP      | cilium-operator | Prometheus metric endpoint for the Cilium operator.                                                                           |
| 10250 | TCP      | kubelet         | Kubelet API. Anonymous authentication is disabled. X509 client certificate required.                                          |
| 10257 | TCP      | kube-controller | Kubernetes controller manager API. HTTPS with authentication and authorization.                                               |
| 10259 | TCP      | kube-scheduler  | Kubernetes scheduler API. HTTPS with authentication and authorization.                                                        |

### Services binding to the localhost interface

| Port  | Protocol | Service         | Description                                                             |
|-------|----------|-----------------|-------------------------------------------------------------------------|
| 2380  | TCP      | etcd            | SSL encrypted peer connection to etcd. Client certificate required.     |
| 9234  | TCP      | cilium-operator | cilium-operator  Address to serve API requests.                         |
| 9879  | TCP      | cilium-agent    | TCP port for the Cilium agent health status API.                        |
| 9890  | TCP      | cilium-agent    | cilium agent [gops](https://github.com/google/gops) server endpoint.    |
| 9891  | TCP      | cilium-operator | cilium-operator [gops](https://github.com/google/gops) server endpoint. |
| 10248 | TCP      | kubelet         | Localhost health check endpoint.                                        |
| 10249 | TCP      | kube-proxy      | Port for the metrics server.                                            |
| 10256 | TCP      | kube-proxy      | Port for binding the health check server.                               |

## Socket Service

### Containerd

Containerd is being exposed through unix socket.

| Service    | Socket                                 |
|------------|----------------------------------------|
| containerd | unix:///run/containerd/containerd.sock |
