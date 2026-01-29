# Getting started

{{product}} is a distribution of Kubernetes which includes all
the necessary tools and services needed to easily deploy and manage a cluster.
Upstream Kubernetes does not provide you with a fully functional cluster by 
default, which is why we have bundled everything you need into a snap that 
should only take a few minutes to install and deploy.

In this tutorial you will deploy a single node cluster by installing the snap 
package. You will also execute some typical cluster operations such as 
deploying an NGINX server as a sample workload and configuring cluster storage.

## Prerequisites

- System Requirements: Your machine should have at least **20G disk space**
  and **8G of memory**
- An **Ubuntu** environment to run the commands (or
  another operating system which supports snapd - see the
  [snapd documentation](https://snapcraft.io/docs/installing-snapd))
- A system with **no previous installations of containerd/docker** as this may
cause conflicts. Consider using a [Multipass virtual machine] if you would like 
an isolated working environment.

## Install {{product}}

Install the {{product}} `k8s` snap with:

```{literalinclude} ../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

This may take a few moments as the snap installs all the necessary Kubernetes
components for a fully functioning cluster such as the networking, storage, etc.

## Bootstrap a Kubernetes cluster

The bootstrap command initializes your cluster and configures your host system
as a Kubernetes node. Bootstrapping the cluster is only done once at cluster
creation.

Bootstrap a Kubernetes cluster with default configuration:

```
sudo k8s bootstrap
```

Once the bootstrap command has been successfully ran, the output should list the
node address and confirm the CNI is being deployed.

## Check cluster status

It may take a few minutes for the cluster to be ready. To confirm the
installation was successful, use `k8s status` with the `wait-ready` flag
to wait for {{product}} to bring up the cluster:


```
sudo k8s status --wait-ready
```

```{important}
This command waits a few minutes before timing out.
On a very slow network connection, or a system with very limited resources,
this default timeout might be insufficient resulting in a "Context canceled"
error. Please first ensure that your machine meets the system requirements to run a Kubernetes cluster. Then, you can either increase the timeout using the  `--timeout`
flag or re-run the command to continue waiting until the cluster is ready.
```

Congratulations, you have just deployed a single node cluster with {{product}}! 
Now let's see what you can do with it.

## Access Kubernetes

The standard tool for deploying and managing workloads on Kubernetes
is [kubectl](https://kubernetes.io/docs/reference/kubectl/).
For convenience, {{product}} bundles a version of
kubectl for you to use with no extra setup or configuration.
For example, to view your node you can run the command:

```
sudo k8s kubectl get nodes
```

…or to see the running services:

```
sudo k8s kubectl get services
```

Run the following command to list all the pods in the `kube-system`
namespace:

```
sudo k8s kubectl get pods -n kube-system
```

A [pod](https://kubernetes.io/docs/concepts/workloads/pods/) is the smallest
deployable unit in Kubernetes.  You will observe at least four pods running. 
The status of the pods may be in `ContainerCreating` while they are being 
initialized. Run the command again after a few seconds and they should report 
status as `Running`.

The functions of these pods are:

- **CoreDNS (`coredns`)**: Provides DNS resolution services.
- **Network operator (`cilium-operator`)**: Manages the life-cycle of the
networking solution.
- **Network agent (`cilium`)**: Facilitates network management.
- **Storage controller (`ck-storage-rawfile-csi-controller`)**: Manages the
life-cycle of the local storage solution.
- **Storage agent (`ck-storage-rawfile-csi-node`)** : Facilitates local storage
management.


## Deploy an app

Kubernetes is meant for deploying apps and services.
You can use the `kubectl`
command to do that as with any Kubernetes.

Let's deploy a demo web server (NGINX):

```
sudo k8s kubectl create deployment nginx --image=nginx
```

This command launches a pod running the NGINX application within a
container.

You can check the status of your pods by running:

```
sudo k8s kubectl get pods
```

This command shows all pods in the default namespace.
It may take a moment for the pod to be ready and running.

Now to check the NGINX server in the pod is working correctly, get the IP 
address of the pod by running the same command again but this time we will add 
the `-owide` argument so we get more information about the pod: 

 ```
sudo k8s kubectl get pods -owide
```

Then query the NGINX IP address using `curl`:

```
curl <POD_IP>
```

The output should confirm NGINX was successfully installed and working.

## Remove an app

To remove the NGINX workload, execute the following command:

```
sudo k8s kubectl delete deployment nginx
```

To verify that the pod has been removed, you can check the status of pods by
running:

```
sudo k8s kubectl get pods
```

## Enable local storage

As we learned earlier, {{product}} comes with everything you need to run 
and manage your cluster. Using the `enable` command you can easily turn on and 
off the bundled services. 

For example, with {{product}} you can enable local storage:

```
sudo k8s enable local-storage
```

To verify that the local storage is enabled, execute:

```
sudo k8s status
```

You should see `local-storage: enabled` in the command output.

Enabling local storage allows you to create persistent volumes which is the 
Kubernetes way to preserve application data beyond the life-cycle of a pod. 
Let's create a `PersistentVolumeClaim` and use it in a `Pod`:

```
sudo k8s kubectl apply -f https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/canonicalk8s/assets/tutorial-pod-with-pvc.yaml
```

This command deploys the YAML configuration of a pod called `storage-writer-pod`
and a persistent volume claim called `myclaim` with a capacity of 1G.

To confirm that the persistent volume is up and running:

```
sudo k8s kubectl get pvc myclaim
```

You can inspect the storage-writer-pod with:

```
sudo k8s kubectl describe pod storage-writer-pod
```

You should see `myclaim` listed under Volumes showing it has been 
assigned correctly. 

## Disable local storage

Begin by removing the pod along with the persistent volume claim:

```
sudo k8s kubectl delete pod storage-writer-pod
sudo k8s kubectl delete pvc myclaim
```

This may take a few moments as the cluster cleans up its resources.

Next, disable the local storage:

```
sudo k8s disable local-storage
```

## Remove {{product}} (Optional)

If you would like to maintain a snapshot of the `k8s` snap for future
restoration, simply run :

```
sudo snap remove k8s
```

The snapshot is a copy of the user, system and configuration data stored by
snapd for the `k8s` snap. This data can be found in `/var/snap/k8s`.

If you wish to remove the snap without saving any of its data execute:

```
sudo snap remove k8s --purge
```

The `--purge` flag ensures complete removal of the snap and its associated data.

## Next steps

- Learn more about {{product}} with kubectl in our [how to use kubectl] tutorial
- Learn how to set up a multi-node environment by [adding and removing nodes]
- Explore Kubernetes commands with our [command reference guide]

<!-- LINKS -->

[how to use kubectl]: kubectl
[command reference guide]: /snap/reference/commands
[adding and removing nodes]: add-remove-nodes
[Multipass virtual machine]: /snap/howto/install/multipass
