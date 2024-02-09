# Getting started

## What you will need
- An Ubuntu 22.04 LTS, 20.04 LTS, 18.04 LTS or 16.04 LTS environment to run the commands (or another operating system which supports snapd - see the [snapd documentation](https://snapcraft.io/docs/installing-snapd?_ga=2.260591181.190515620.1707126165-607758451.1707126165))
- System Requirements: Your machine should have at least 20G disk space and 4G of memory

### 1. Install Canonical K8s
Canonical Kubernetes can be installed with a snap:
```
sudo snap install k8s
```

### 2. Bootstrap K8s
Initialize Kubernetes on your host system with default configurations using:
```
sudo k8s bootstrap
```
This command initializes your cluster and configures your host system as a Kubernetes node.
### 2. Check K8s Status
To confirm the installation was successful and your node is ready you may run:
```
sudo k8s status
```
You should see `k8s is ready` in the command output. You may have to wait a moment for things to settle down.
### 3. Access Kubernetes
Canonical K8s bundles its own version of `kubectl` for accessing Kubernetes. Use it to run commands to monitor and control your Kubernetes. For example, to view your node:
```
sudo k8s kubectl get nodes
```
…or to see the running services:
```
sudo k8s kubectl get services
```

### 3. Deploy an app
Kubernetes is meant for deploying apps and services. You can use the kubectl command to do that as with any Kuberenetes. 

Let's deploy a demo nginx server:
```
sudo k8s kubectl create deployment nginx --image=nginx
```
You can check the status of your pods by running:
```
sudo k8s kubectl get pods
```
This command shows all pods in the default namespace. It may take a moment for the pod to be ready and running.
### 4. Enable Components
With Canonical Kubernetes, you can enable and disable core components such as DNS, gateway, ingress, network, and storage. For an overview of components, visit:[Components Overview](https://change-me.com)

DNS resolution is fundamental for communication between pods within the cluster and is essential for any Kubernetes deployment. To enable DNS resolution, run:
```
sudo k8s enable dns
```
To enable network connectivity and configure Cilium as the networking solution, execute:
```
sudo k8s enable network
```
Run the following command to list all the pods in the `kube-system` namespace:
```
sudo k8s kubectl get pods -n kube-system
```
You will observe three pods running:
- `coredns`: Provides DNS resolution services
- `cilium-operator`: Manages the lifecycle of the Cilium networking solution.
- `cilium` Represents the Cilium networking agent.
### 5. Configure a component
???
## Next Steps
Link to further topics (networking, command ref, ingress etc)
- Explore Kubernetes commands with our [Command Reference Guide](https://change-me.com)
- Bootstrap K8s with your custom configurations [Bootstrap K8s](https://change-me.com)
- Learn how to set up a multi-node environment [Setting up a K8s cluster](https://change-me.com)
- Master Kubernetes networking concepts: [Networking](https://change-me.com)
- Discover how to enable and configure Ingress resources : [Ingress](https://change-me.com)
