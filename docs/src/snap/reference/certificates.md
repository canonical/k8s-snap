### Kubernetes Cluster Certificates Reference Page

#### 1. **Root CAs**

| **Common Name**         | **Usage**                            | **Path on Disk**                     | **Used For**                              |
|------------------------------|--------------------------------------|--------------------------------------|-------------------------------------------|
| `kubernetes-ca`               | General Kubernetes CA               | `/etc/kubernetes/pki/ca.crt`         | Signing all Kubernetes-related certificates |
| `kubernetes-front-proxy-ca`   | CA for front-end proxy              | `/etc/kubernetes/pki/front-proxy-ca.crt` | Signing certificates for the front-proxy |
| `client-ca`                   | CA for client certificates          | `/etc/kubernetes/pki/client-ca.crt` | Signing certificates for the client |

#### 3. **Certificates Used by the API Server**

| **Common Name (CN)**            | **Usage**                     | **Path on Disk**                       | **Used For**                              |
|---------------------------------|-------------------------------|----------------------------------------|-------------------------------------------|
| `kube-apiserver`                | Server                        | `/etc/kubernetes/pki/apiserver.crt`    | Securing the API server endpoint          |
| `kube-apiserver-kubelet-client` | Client                        | `/etc/kubernetes/pki/apiserver-kubelet-client.crt` | API server communication with kubelets    |
| `kube-apiserver-etcd-client`    | Client                        | `/etc/kubernetes/pki/apiserver-etcd-client.crt` | API server communication with etcd        |
| `front-proxy-client`            | Client                        | `/etc/kubernetes/pki/front-proxy-client.crt` | API server communication with the front-proxy |

#### 4. **Certificates Used by etcd**

| **Common Name (CN)**        | **Usage**              | **Path on Disk**                         | **Used For**                               |
|-----------------------------|------------------------|------------------------------------------|--------------------------------------------|
| `kube-etcd`                 | Server, Client         | `/etc/kubernetes/pki/etcd/server.crt`    | Securing communication with etcd           |
| `kube-etcd-peer`            | Server, Client         | `/etc/kubernetes/pki/etcd/peer.crt`      | Securing communication between etcd peers  |
| `kube-etcd-healthcheck-client` | Client               | `/etc/kubernetes/pki/etcd/healthcheck-client.crt` | Health checks on etcd |

#### 5. **Certificates for Kubernetes Components**

| **Config File**                    | **Usage**                              | **Path on Disk**                           | **Used For**                                 |
|------------------------------------|----------------------------------------|--------------------------------------------|----------------------------------------------|
| `admin.conf`                       | Administrator Client Config            | `/etc/kubernetes/admin.conf`               | Admin access to the cluster                  |
| `controller-manager.conf`          | Controller Manager Client Config       | `/etc/kubernetes/controller-manager.conf`  | Communication with the API server            |
| `scheduler.conf`                   | Scheduler Client Config                | `/etc/kubernetes/scheduler.conf`           | Communication with the API server            |
| `kubelet.conf`                     | Kubelet Client Config                  | `/etc/kubernetes/kubelet.conf`             | Node registration and communication with API server |
