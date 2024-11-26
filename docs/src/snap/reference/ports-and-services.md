# Services and ports

## Network Services

There are two main types of services based on the network interface they use:

* Default Host Interface Services: These services bind to the default host interface, 
making them accessible from outside the host.
* Localhost Services: These services bind to the localhost interface,
meaning they can only be accessed from within the host.

### Services binding to the default Host interface

| Port  | Service         | Description                                                                                              |
|-------|-----------------|----------------------------------------------------------------------------------------------------------|
| 4244  | cilium-agent    | Listening address for Hubble.                                                                            |                                                                                     
| 4240  | cilium-agent    | TCP port for cluster-wide network connectivity and Cilium agent health API.                              |
| 6400  | k8sd            | Default REST API port for Canonical Kubernetes daemon.                                                   |
| 6443  | kube-apiserver  | Kubernetes API server. SSL encrypted. Clients must present a valid password from a Static Password File. |                                                                                     
| 9000  | k8s-dqlite      | SSL encrypted connection for k8s-dqlite. Client certificates required.                                   |
| 9963  | cilium-operator | Prometheus metric endpoint for the Cilium operator.                                                      |                                                                                   
| 10250 | kubelet         | Kubelet API. Anonymous authentication is disabled. X509 client certificate required.                     |                                                                                          
| 10257 | kube-controller | Kubernetes controller manager API. HTTPS with authentication and authorization.                          |                                                                                    
| 10259 | kube-scheduler  | Kubernetes scheduler API. HTTPS with authentication and authorization.                                   |                                                                                     


### Services binding to the localhost interface

| Port  | Service         | Description                                                             |
|-------|-----------------|-------------------------------------------------------------------------|
| 9234  | cilium-operator | cilium-operator  Address to serve API requests.                         |
| 9879  | cilium-agent    | TCP port for the Cilium agent health status API.                        |
| 9890  | cilium-agent    | cilium agent [gops](https://github.com/google/gops) server endpoint.    |     
| 9891  | cilium-operator | cilium-operator [gops](https://github.com/google/gops) server endpoint. |
| 10248 | kubelet         | Localhost health check endpoint.                                        |
| 10249 | kube-proxy      | Port for the metrics server.                                            |
| 10256 | kube-proxy      | Port for binding the health check server.                               |
 
## Socket Service

### Containerd

Containerd is being exposed through unix socket.

| Service      | Socket                                 |
|--------------|----------------------------------------|
| containerd 	 | unix:///run/containerd/containerd.sock |
