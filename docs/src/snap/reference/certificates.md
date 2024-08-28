### Cluster Certificates

#### 1. **CertificateAuthorities (CA)**

| **Common Name**         | **Usage**                            | **Path on Disk**                     | **Used For**                              |
|------------------------------|--------------------------------------|--------------------------------------|-------------------------------------------|
| `kubernetes-ca`               | General Kubernetes CA               | `/etc/kubernetes/pki/ca.crt`         | Signing all Kubernetes-related certificates |
| `kubernetes-front-proxy-ca`   | CA for front-end proxy              | `/etc/kubernetes/pki/front-proxy-ca.crt` | Signing certificates for the front-proxy |
| `client-ca`                   | CA for client certificates          | `/etc/kubernetes/pki/client-ca.crt` | Signing certificates for the client |


#### 2. **Certificates**

| **Common Name**                       | **Usage** | **Path on Disk**                                     | **Used For**                                                     | **Signed By**               |
|--------------------------------------------|-----------|------------------------------------------------------|------------------------------------------------------------------|-----------------------------|
| `kube-apiserver`                           | Server    | `/etc/kubernetes/pki/apiserver.crt`                  | Securing the API server endpoint                                 | `kubernetes-ca`             |
| `kube-apiserver-kubelet-client`            | Client    | `/etc/kubernetes/pki/apiserver-kubelet-client.crt`   | API server communication with kubelets                           | `kubernetes-ca-client`      |
| `kube-apiserver-etcd-client`               | Client    | `/etc/kubernetes/pki/apiserver-etcd-client.crt`      | API server communication with etcd                               | `kubernetes-ca-client`      |
| `front-proxy-client`                       | Client    | `/etc/kubernetes/pki/front-proxy-client.crt`         | API server communication with the front-proxy                    | `kubernetes-front-proxy-ca` |
| `kube-controller-manager`                  | Client    | `/etc/kubernetes/pki/controller-manager.crt`         | Communication between the controller manager and the API server  | `kubernetes-ca-client`      |
| `kube-scheduler`                           | Client    | `/etc/kubernetes/pki/scheduler.crt`                  | Communication between the scheduler and the API server           | `kubernetes-ca-client`      |
| `kube-proxy`                               | Client    | `/etc/kubernetes/pki/proxy.crt`                      | Communication between kube-proxy and the API server              | `kubernetes-ca-client`      |
| `system:node:$hostname`                    | Client    | `/etc/kubernetes/pki/kubelet-client.crt`             | Authentication of kubelets to the API server                     | `kubernetes-ca-client`      |
| `k8s-dqlite`             | Client    | `/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt`             | Communication between k8s-dqlite nodes and API server | `self-signed`      |
| `root@$hostname`             | Client    | `/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt`             | Communication between k8sd nodes | `self-signed`      |


#### 5. **Configurations for Kubernetes Components**

| **Config File**                    | **Usage**                              | **Path on Disk**                           | **Used For**                                 |
|------------------------------------|----------------------------------------|--------------------------------------------|----------------------------------------------|
| `admin.conf`                       | Administrator Client Config            | `/etc/kubernetes/admin.conf`               | Admin access to the cluster                  |
| `controller-manager.conf`          | Controller Manager Client Config       | `/etc/kubernetes/controller-manager.conf`  | Communication with the API server            |
| `scheduler.conf`                   | Scheduler Client Config                | `/etc/kubernetes/scheduler.conf`           | Communication with the API server            |
| `kubelet.conf`                     | Kubelet Client Config                  | `/etc/kubernetes/kubelet.conf`             | Node registration and communication with API server |
