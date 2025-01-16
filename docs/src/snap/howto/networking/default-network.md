# How to use the default Network

{{product}} includes a high-performance, advanced network plugin
called Cilium. The network component allows cluster administrators to leverage
software-defined networking to automatically scale and secure network policies
across their cluster.

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine.
- You have a bootstrapped {{product}} cluster (see the [Getting
  Started][getting-started-guide] guide).

## Check Network status

Find out whether Network is enabled or disabled with the following command:

```
sudo k8s status
```

The default state for the cluster is `network disabled`.

## Enable Network

To enable Network, run:

```
sudo k8s enable network
```

For more information on the command, execute:

```
sudo k8s enable --help
```

## Configure Network

It is not possible to reconfigure the network on a running cluster as this will
lead to unreachable pods/services and nodes. Any configuration options the CNI
needs to be aware of (e.g. pod and service CIDR, IPv6 support) are set during
the cluster bootstrap (`k8s bootstrap` command).

### Check Network details

Let's look at the detailed status of the network as reported by Cilium.

First, find the name of the Cilium pod:

```sh
sudo k8s kubectl get pod -n kube-system -l k8s-app=cilium
```

Once you have the name of the pod, run the following command to see Cilium's
status:

```sh
sudo k8s kubectl exec -it cilium-97vcw -n kube-system -c cilium-agent \
  -- cilium status
```

You should see a wide range of metrics and configuration values for your
cluster.

## Disable Network

You can `disable` the built-in network:

``` {warning}
   If you have an active cluster, disabling Network may impact external
   access to services within your cluster.
   Ensure that you have alternative configurations in place before
   disabling Network.
```

If your underlying network is Cilium you will have to run
`sudo k8s disable gateway` before disabling network.

```
sudo k8s disable network
```

For more information on this command, run:

```
sudo k8s disable --help
```

<!-- LINKS -->

[getting-started-guide]: ../../tutorial/getting-started
