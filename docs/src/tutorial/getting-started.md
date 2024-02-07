# Getting started

## What you will need
- An Ubuntu 22.04 LTS, 20.04 LTS, 18.04 LTS or 16.04 LTS environment to run the commands (or another operating system which supports snapd - see the [snapd documentation](https://snapcraft.io/docs/installing-snapd?_ga=2.260591181.190515620.1707126165-607758451.1707126165))
- System Requirements: Your machine should have at least 20G disk space and 4G of memory

### 1. Install Canonical K8s
Canonical Kubernetes can be installed with a snap:
```
snap install k8s
```

To confirm the installation was successful you may run:
```
k8s version
```
### 2. Access Kubernetes
MicroK8s bundles its own version of `kubectl` for accessing Kubernetes. Use it to run commands to monitor and control your Kubernetes. For example, to view your node:
```
k8s kubectl get nodes
```
…or to see the running services:
```
k8s kubectl get services
```
K8s uses a namespaced kubectl command to prevent conflicts with any existing installs of kubectl. If you don’t have an existing install, it is easier to add an alias (append to ~/.bash_aliases) like this:
```
alias kubectl='k8s kubectl'
```
### 3. Deploy an app
Of course, Kubernetes is meant for deploying apps and services. You can use the kubectl command to do that as with any Kuberenetes. Try installing a demo app:
```
microk8s kubectl create deployment nginx --image=nginx
```
It may take a minute or two to install, but you can check the status:
```
microk8s kubectl get pods
```
### 4. Enable a Component

DNS resolution is fundamental for communication between pods within the cluster and is essential for any Kubernetes deployment.
```
k8s enable dns
```
Run the following command to list all the pods in the kube-system namespace:
```
kubectl get pods -n kube-system
```
### 5. Configure a component

## Next Steps
Link to further topics (networking, command ref, ingress etc)