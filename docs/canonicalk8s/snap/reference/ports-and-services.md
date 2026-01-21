# Services and ports

## Network services

### Services on all nodes

| Interface & Port | Protocol | Service        | Description                                                                                              |
| ---------------- | -------- | -------------- | -------------------------------------------------------------------------------------------------------- |
| localhost:9879   | TCP      | cilium-agent   | TCP port for the Cilium agent health status API.                                                         |
| localhost:9890   | TCP      | cilium-agent   | Cilium agent [gops](https://github.com/google/gops) server endpoint.                                     |
| localhost:10248  | TCP      | kubelet        | Localhost health check endpoint.                                                                         |
| localhost:10249  | TCP      | kube-proxy     | Port for the metrics server.                                                                             |
| localhost:10256  | TCP      | kube-proxy     | Port for binding the health check server.                                                                |
| default:4240     | TCP      | cilium-agent   | TCP port for cluster-wide network connectivity and Cilium agent health API.                              |
| default:6400     | TCP      | k8sd           | Default REST API port for Canonical Kubernetes daemon.                                                   |
| *:4244           | TCP      | cilium-agent   | Listening address for Hubble.                                                                            |
| *:6443           | TCP      | kube-apiserver | Kubernetes API server. SSL encrypted. Clients must present a valid password from a Static Password File. |
| *:10250          | TCP      | kubelet        | Kubelet API. Anonymous authentication is disabled. X509 client certificate required.                     |

### Services on control plane nodes only

| Interface & Port | Protocol | Service         | Description                                                                                                                   |
| ---------------- | -------- | --------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| localhost:9234   | TCP      | cilium-operator | cilium-operator address to serve API requests.                                                                                |
| localhost:9891   | TCP      | cilium-operator | cilium-operator [gops](https://github.com/google/gops) server endpoint.                                                       |
| default:2379     | TCP      | etcd            | SSL encrypted client connection to etcd. Client certificate required.                                                         |
| default:2380     | TCP      | etcd            | SSL encrypted peer connection to etcd. Client certificate required.                                                           |
| *:9963           | TCP      | cilium-operator | Prometheus metric endpoint for the Cilium operator.                                                                           |
| *:8472           | UDP      | cilium-agent    | Default VXLAN port used by Cilium.                                                                                            |
| *:10257          | TCP      | kube-controller | Kubernetes controller manager API. HTTPS with authentication and authorization.                                               |
| *:10259          | TCP      | kube-scheduler  | Kubernetes scheduler API. HTTPS with authentication and authorization.                                                        |

## Socket service

### Containerd

Containerd is being exposed through unix socket.

| Service    | Socket                                 |
| ---------- | -------------------------------------- |
| containerd | unix:///run/containerd/containerd.sock |
