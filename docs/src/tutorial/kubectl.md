# Basic operations with Kubernetes using kubectl

## What you will need

Before you begin, make sure you have the following:

- A bootstrapped Canonical Kubernetes cluster
(See [Getting Started](https://github.com/canonical/k8s-snap/blob/main/docs/src/tutorial/getting-started.md))

### 1. The Kubectl Command

The `kubectl` command communicates with the [Kubernetes API server](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/).

The `kubectl` command included with Canonical Kubernetes is built from the 
original upstream source into the `k8s` snap you have installed.

### 2. How To Use Kubectl

To access `kubectl`, run the following command:

```
sudo k8s kubectl <command>
```

Note: Only control plane nodes can use the `kubectl` command. Worker nodes do 
not have access to this command.

### 3. Configuration

In Canonical Kubernetes, the `kubeconfig` file that is being read to display 
the configuration when you run `kubectl config view` lives at `/snap/k8s
current/k8s/config/kubeconfig`. You can change this by setting a `KUBECONFIG` 
environment variable or passing the `--kubeconfig` flag to a command.

To find out more, you can visit 
[the official kubeconfig documentation](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)

### 4. Viewing objects

Let's review what was created in the Getting Started guide.

To see what pods were created when we enabled the `network` and `dns` 
components:

```
sudo k8s kubectl get pods -o wide -n kube-system
```

You should be seeing the `network-operator`, `networking-agent` and `coredns` pods.

```
sudo k8s kubectl get services --all-namespace
```

The `kubernetes` service in the `default` namespace is where the Kubernetes 
API server resides, and it's the endpoint with which other nodes in your 
cluster will communicate.

The `hubble-peer` service in the `kube-system` namespace is created by Canonical Kubernetes (an opinionated Kubernetes distribution) to ...

### 5. Creating and Managing Objects

Let's deploy an NGINX server using this command:

```
sudo k8s kubectl create deployment nginx --image=nginx:latest
```

To observe the NGINX pod running in the default namespace:

```
sudo k8s kubectl get pods
```

Let's now scale this deployment, which means increasing the number of pods it 
manages.

```
sudo k8s kubectl scale deployment nginx --replicas=3
```

Execute `sudo k8s kubectl get pods` again and notice that you have 3 NGINX 
pods.

Let's delete those 3 pods to demonstrate a deployment's ability to ensure the 
declared state of the cluster is maintained. 

First, open a new terminal so you can watch the changes as they happen. Run 
this command in a new terminal:

```
sudo k8s kubectl get pods --all-namespace --watch
```

Now, go back to your original terminal and run:

```
sudo k8s kubectl delete pods -l app=nginx
```

The above command deletes all pods in the cluster that are labeled with `app=nginx`.

You'll notice the original 3 pods will have a status of `Terminating` and 3 
new pods will have a status of `ContainerCreating`.

### 6. Remove Canonical Kubernetes (Optional)

To uninstall the Canonical Kubernetes snap, execute:

```
sudo snap remove k8s
```

This command removes the `k8s` snap and automatically creates a snapshot of 
all data for future restoration.

If you wish to remove the snap without saving a snapshot of its data, add 
`--purge` to the command:

```
sudo snap remove k8s --purge
```
This option ensures complete removal of the snap and its associated data.

## Next Steps

- Explore Kubernetes commands with our [Command Reference Guide](#TODO)
- Bootstrap Kubernetes with your custom configurations [Bootstrap Kubernetes](#TODO)
- Learn how to set up a multi-node environment [Setting up a Kubernetes cluster](#TODO)
- Configure storage options [Storage](#TODO)
- Master Kubernetes networking concepts: [Networking](#TODO)
- Discover how to enable and configure Ingress resources [Ingress](#TODO)

## References
[https://kubernetes.io/docs/reference/kubectl/](https://kubernetes.io/docs/reference/kubectl/)