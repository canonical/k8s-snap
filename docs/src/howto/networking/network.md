# How to use default Network

Canonical Kubernetes includes a high-performance, advanced network plugin called Cilium. The network component allows cluster administrators to leverage
software-defined networking to automatically scale and secure network policies
across their cluster.

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine.
- You have a bootstraped Canonical Kubernetes cluster (see the [Getting Started][getting-started-guide] guide).

## Check Network status

Find out whether Network is enabled or disabled with the following command:

```bash
sudo k8s status
```

The default state for the cluster is `network disabled`.

## Enable Network

To enable Network, run:

```bash
sudo k8s enable network
```

For more information on the command, execute:

```bash
sudo k8s help enable
```

## Configure Network

Discover your configuration options by running:

```bash
sudo k8s set network –-help
```

### Check Network details

Let's look at the detailed status of the network as reported by Cilium.

First, find the nmae of the Cilium pod:

```sh
sudo k8s kubectl get pod -n kube-system -l k8s-app=cilium
```

Once you have the name of the pod, run the following command to see Cilium's status:

```sh
sudo k8s kubectl exec -it cilium-97vcw -n kube-system -c cilium-agent -- cilium status
```

You should see a wide range of metrics and configuration values for your cluster.

## Disable Network

You can `disable` the built-in network:

``` {warning} If you have custom rules in place, disabling Network may impact external access to services within your cluster.
    Ensure that you have alternative configurations in place before disabling Network.
```

```bash
sudo k8s disable network
```

For more information on this command, run:

```bash
sudo k8s help disable
```

<!-- LINKS -->

[kubectl-create-secret-tls/]: https://kubernetes.io/docs/reference/kubectl/generated/kubectl_create/kubectl_create_secret_tls/
[proxy-protocol]: https://kubernetes.io/docs/reference/networking/service-protocols/#protocol-proxy-special
[getting-started-guide]: ../../../tutorial/getting-started
[kubectl-guide]: ../../../tutorial/kubectl
