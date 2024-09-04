# Cluster Certificates and Configuration Directories

This reference page provides an overview of certificate authorities (CAs),
certificates and configuration directories in use by a {{ product }} cluster.

## Certificate Authorities (CAs)

This table outlines the common certificate authorities (CAs) used in a
Kubernetes environment, detailing their specific purposes, usage,
and locations on the disk.

| **Common Name**                            | **Purpose** | **File Location**  | **Primary Function**          |
|--------------------------------------------|-----------|----------------------|-------------------------------|
| `kubernetes-ca`               | General Kubernetes CA               | `/etc/kubernetes/pki/ca.crt`             | Signing all Kubernetes-related certificates |
| `kubernetes-front-proxy-ca`   | CA for front-end proxy              | `/etc/kubernetes/pki/front-proxy-ca.crt` | Signing certificates for the front-proxy    |
| `client-ca`                   | CA for client certificates          | `/etc/kubernetes/pki/client-ca.crt`      | Signing certificates for the client         |


## Certificates

This table provides an overview of the certificates currently in use,
including their roles, storage paths, and the entities responsible for
their issuance.


| **Common Name**                            | **Purpose** | **File Location**  | **Primary Function**            | **Signed By**               |
|--------------------------------------------|-----------|------------------------------------------------------|------------------------------------------------------------------|-----------------------------|
| `kube-apiserver`                           | Server    | `/etc/kubernetes/pki/apiserver.crt`                  | Securing the API server endpoint                                 | `kubernetes-ca`             |
| `apiserver-kubelet-client`            | Client    | `/etc/kubernetes/pki/apiserver-kubelet-client.crt`   | API server communication with kubelets                           | `kubernetes-ca-client`      |
| `kube-apiserver-etcd-client`               | Client    | `/etc/kubernetes/pki/apiserver-etcd-client.crt`      | API server communication with etcd                               | `kubernetes-ca-client`      |
| `front-proxy-client`                       | Client    | `/etc/kubernetes/pki/front-proxy-client.crt`         | API server communication with the front-proxy                    | `kubernetes-front-proxy-ca` |
| `system:kube-controller-manager`                  | Client    | `/etc/kubernetes/pki/controller-manager.crt`         | Communication between the controller manager and the API server  | `kubernetes-ca-client`      |
| `system:kube-scheduler`                           | Client    | `/etc/kubernetes/pki/scheduler.crt`                  | Communication between the scheduler and the API server           | `kubernetes-ca-client`      |
| `system:kube-proxy`                               | Client    | `/etc/kubernetes/pki/proxy.crt`                      | Communication between kube-proxy and the API server              | `kubernetes-ca-client`      |
| `system:node:$hostname`                    | Client    | `/etc/kubernetes/pki/kubelet-client.crt`             | Authentication of kubelets to the API server                     | `kubernetes-ca-client`      |
| `k8s-dqlite`                               | Client    | `/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt`| Communication between k8s-dqlite nodes and API server            | `self-signed`               |
| `root@$hostname`                          | Client    | `/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt`             | Communication between k8sd nodes | `self-signed`      |


## Configuration Files for Kubernetes Components

The following tables provide an overview of the configuration files used to
communicate with the cluster services.

### Control-plane node

Control-plane nodes use the following configuration files.

| **Configuration File**             | **Purpose**                            | **File Location**                          | **Primary Function**                                 |
|------------------------------------|----------------------------------------|--------------------------------------------|----------------------------------------------|
| `admin.conf`                       | Administrator Client Config            | `/etc/kubernetes/admin.conf`               | Admin access to the cluster                  |
| `controller-manager.conf`          | Controller Manager Client Config       | `/etc/kubernetes/controller-manager.conf`  | Communication with the API server            |
| `scheduler.conf`                   | Scheduler Client Config                | `/etc/kubernetes/scheduler.conf`           | Communication with the API server            |
| `kubelet.conf`                     | Kubelet Client Config                  | `/etc/kubernetes/kubelet.conf`             | Node registration and communication with API server |
| `proxy.conf`                       | Proxy Client Config                    | `/etc/kubernetes/proxy.conf`               | Communication with the API server            |

### Worker node

Worker nodes use the following configuration files.

| **Configuration File**             | **Purpose**                            | **File Location**                          | **Primary Function**                                 |
|------------------------------------|----------------------------------------|--------------------------------------------|----------------------------------------------|
| `proxy.conf`                       | Proxy Client Config                    | `/etc/kubernetes/proxy.conf`               | Communication with the API server            |
| `kubelet.conf`                     | Kubelet Client Config                  | `/etc/kubernetes/kubelet.conf`             | Node registration and communication with API server |
