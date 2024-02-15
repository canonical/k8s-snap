# Getting started

## What you will need
- An Ubuntu 22.04 LTS or 20.04 LTS environment to run the commands (or
  another operating system which supports snapd - see the
  [snapd documentation](https://snapcraft.io/docs/installing-snapd))
- System Requirements: Your machine should have at least 20G disk space
  and 4G of memory

### 1. Install Canonical Kubernetes

Install the Canonical Kubernetes snap with:
```
sudo snap install --edge k8s --classic
```

### 2. Bootstrap a Kubernetes Cluster

Bootstrap a Kubernetes cluster with default configuration using:

```
sudo k8s bootstrap
```

This command initialises your cluster and configures your host system 
as a Kubernetes node.
For custom configurations, you can explore additional options using: 

```
sudo k8s bootstrap --help
```

### 3. Check cluster status

To confirm the installation was successful and your node is ready you
should run:

```
sudo k8s status
```

You should see `k8s is not ready` in the command output. This will
change once we've enabled the `network` and `dns` components.

### 4. Enable Components (DNS, Network)

With Canonical Kubernetes, you can enable and disable core components
such as DNS, gateway, ingress, network, and storage. For an overview
of components, see the [Components Overview](#TODO)

DNS resolution is fundamental for communication between pods within
the cluster and is essential for any Kubernetes deployment. To enable
DNS resolution, run:

```
sudo k8s enable dns
```

To enable network connectivity execute:

```
sudo k8s enable network
```

Run the following command to list all the pods in the `kube-system`
namespace:

```
sudo k8s kubectl get pods -n kube-system
```

You will observe three pods running:
- **coredns**: Provides DNS resolution services.
- **network-operator**: Manages the lifecycle of the networking solution.
- **network**: Facilitates network management.

Confirm that Canonical Kubernetes has transitioned to the `k8s is ready` state by running:

```
sudo k8s status --wait-ready
```

Note: To disable a component execute `sudo k8s disable <component>`

### 5. Access Kubernetes
The standard tool for deploying and managing workloads on Kuberenetes
is [kubectl](https://kubernetes.io/docs/reference/kubectl/). 
For convenience, Canonical Kubernetes bundles a version of 
kubectl for you to use with no extra setup or configuration. 
For example, to view your node you can run the command:

```
sudo k8s kubectl get nodes
```

…or to see the running services:

```
sudo k8s kubectl get services
```

### 6. Deploy an app

Kubernetes is meant for deploying apps and services. 
You can use the `kubectl`
command to do that as with any Kubernetes. 

Let's deploy a demo NGINX server:

```
sudo k8s kubectl create deployment nginx --image=nginx
```
This command launches a [pod](https://kubernetes.io/docs/concepts/workloads/pods/),
the smallest deployable unit in Kubernetes,
running the nginx application within a container.

You can check the status of your pods by running:

```
sudo k8s kubectl get pods
```

This command shows all pods in the default namespace. 
It may take a moment for the pod to be ready and running.

### 7. Remove an app
To remove the NGINX workload, execute the following command:
```
sudo k8s kubectl delete deployment nginx

```

To verify that the pod has been removed, you can check the status of pods by running:

```
sudo k8s kubectl get pods
```
### 8. Enable Components (Storage)
In scenarios where you need to preserve application data beyond the 
lifecycle of the pod, Kubernetes provides persistent volumes.

With Canonical Kubernetes, you can enable storage to configure 
your storage solutions:

```
sudo k8s enable storage
```

To verify that the storage component is enabled, execute:
```
sudo k8s status
```
You should see `storage enabled` in the command output.

Let's create a `PersistentVolumeClaim` and use it in a `Pod`. 
For example, we can deploy the following manifest:

```
sudo k8s kubectl apply -f ../../manifests/tutorial-pod-with-pvc.yaml
```
This command deploys a pod based on the YAML configuration of a 
storage writer pod and a persistent volume claim with a capacity of 1G.

To confirm that the persistent volume is up and running:

```
sudo k8s kubectl get pvc myclaim
```

You can inspect the storage-writer-pod with:

```
sudo k8s kubectl describe pod storage-writer-pod
```

### 9. Disable Components (Storage)
Begin by removing the pod along with the persistent volume claim:

```
sudo k8s kubectl delete pvc myclaim
sudo k8s kubectl delete pod storage-writer-pod
```

Next, disable the storage component:

```
sudo k8s disable storage
```

Note: To disable any component, execute  `sudo k8s disable <component>`.

### 10. Remove Canonical Kubernetes (Optional)
To uninstall the Canonical Kubernetes snap, execute:

```
sudo snap remove k8s
```

This command removes the `k8s` snap and automatically creates a snapshot of all data for future restoration.

If you wish to remove the snap without saving a snapshot of its data, add `--purge` to the command:

```
sudo snap remove k8s --purge
```
This option ensures complete removal of the snap and its associated data.

## Next Steps

- Keep mastering Canonical Kubernetes with kubectl: [How to use kubectl](#TODO)
- Explore Kubernetes commands with our [Command Reference Guide](#TODO)
- Bootstrap K8s with your custom configurations [Bootstrap K8s](#TODO)
- Learn how to set up a multi-node environment [Setting up a K8s cluster](#TODO)
- Configure storage options [Storage](#TODO)
- Master Kubernetes networking concepts: [Networking](#TODO)
- Discover how to enable and configure Ingress resources [Ingress](#TODO)
