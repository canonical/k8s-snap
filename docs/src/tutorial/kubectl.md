# Basic operations with Kubernetes using kubectl

### 1. What you will need

Before you begin, make sure you have the following:

- A bootstrapped Canonical K8s cluster (See [insert link]...)

### 2. The Kubectl Command

This commands interacts with the Kubernetes API server (kube-apiserver) and is the most commonly used command when working with Kubernetes, so let's take some time to familiarize ourselves with it.

The `kubectl` command included with Canonical K8s is built from the original upstream source into the `k8s` snap you have installed.

### 3. How To Use Kubectl

You can access kubectl with the following command:

```
sudo k8s kubectl
```

Note: Only control plane nodes can use the `kubectl` command. Worker nodes do not have access to this command.

### 4. Configuration

In Canonical K8s, the `kubeconfig` file that is being read to display the configuration when you run `kubectl config view` lives at `/snap/k8s/current/k8s/config/kubeconfig`. You can change this by setting a `KUBECONFIG` environment variable or passing the `--kubeconfig` flag to a command.

### 5. Viewing objects

Let's review what was created in the Getting Started guide.

Let's see what pods were created when we enabled the `network` and `dns` components.

```
sudo k8s kubectl get pods -o wide -n kube-system
```

You should be seeing Cilium pods.

```
sudo k8s kubectl get svc -A
```

The `kubernetes` service in the `default` namespace is where the Kubernetes API server resides, and it's the endpoint with which other nodes in your cluster will communicate.

The `hubble-peer` service in the `kube-system` namespace is created by Canonical K8s (an opinionated K8s distribution) to ...

### 6. Creating and Managing Objects

Let's create a deployment using this command:

```
sudo k8s kubectl create deployment nginx --image=nginx:latest
```

Notice how `sudo k8s kubectl get pods` shows you one NGINX pod.

Let's now scale this deployment, which means increasing the number of pods it manages.

```
sudo k8s kubectl scale deployment nginx --replicas=3
```

Run `sudo k8s kubectl get pods` again and notice that you have 3 NGINX pods.

Let's delete those 3 pods to demonstrate a deployment's ability to ensure the declared state of the cluster is maintained.

Run `sudo k8s kubectl delete pods -l app=nginx`

If you open another terminal while the above command is executing, you'll notice the original 3 pods will have a status of `Terminating` and 3 new pods will have a status of `ContainerCreating`.

### 7. Remove Canonical Kubernetes (Optional)

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

## References
[https://kubernetes.io/docs/reference/kubectl/](https://kubernetes.io/docs/reference/kubectl/)