# How to integrate Canonical Kubernetes with etcd

Integrating **etcd** with your Canonical Kubernetes deployment provides a
robust, distributed key-value store that is essential for storing critical
data needed for Kubernetes' clustering operations. This guide will walk you
through the process of deploying Canonical Kubernetes with an external etcd
cluster.

## What you will need

- A Juju controller with access to a cloud environment (see the [Juju setup]
  guide for more information).

```{warning} Once you deploy your Canonical Kubernetes cluster with a
particular datastore, you cannot switch to a different datastore
post-deployment. Planning for your datastore needs ahead of time is
crucial, particularly if you opt for an external datastore like **etcd**.
```

## Preparing the Deployment

1. **Creating the Deployment Model**:
  Begin by creating a Juju model specifically for your Canonical Kubernetes
  cluster deployment.

  ```bash
  juju add-model my-cluster
  ```
2. **Deploying Certificate Authority**:
  etcd requires a secure means of communication between its components.
  Therefore, we require a certificates authority such as [EasyRSA][easyrsa-charm]
  or [Vault][vault-charm]. Check the respective charm documentation for detailed
  instructions on how to deploy a certificates authority. In this guide, we will
  be using EasyRSA.

  ```bash
  juju deploy easyrsa
  ```

## Deploying etcd

- **Single Node Deployment**:
  - To deploy a basic etcd instance on a single node, use the command:

    ```bash
    juju deploy etcd
    ```

    This setup is straightforward but not recommended for production environments
    due to lack of high availability.

- **High Availability Setup**:
  - For environments where high availability is crucial, deploy etcd across at
    least three nodes:

    ```bash
    juju deploy etcd -n 3
    ```

    This ensures that your etcd cluster remains available even if one node fails.

## Integrating etcd with EasyRSA

Now you have to integrate etcd with your certificate authority; this will issue
the required certificates for secure communication between etcd and your
Canonical Kubernetes cluster:

```bash
juju integrate etcd easyrsa
```

## Deploying Canonical Kubernetes
Deploy the control plane units of Canonical Kubernetes with the command:

```bash
juju deploy k8s --config datastore=etcd -n 3
```
This command deploys 3 units of the Canonical Kubernetes control plane (`k8s`)
and configures them to use **etcd** as the backing datastore, ensuring high
availability.

## Integrating Canonical Kubernetes with etcd
Now that we have both the etcd datastore deployed alongside our Canonical
Kubernetes cluster, it is time to integrate our cluster with our etcd datastore.

```bash
juju integrate k8s etcd
```

This step integrates the k8s charm (Control Plane units)  with the etcd hosts,
allowing the Kubernetes cluster to utilize the etcd units as an external
datastore.

## Final Steps
**Verify the Deployment**: After completing the deployment, it's essential
to verify that all components are functioning correctly. Use the `juju status`
command to inspect the current status of your cluster.

```bash
➜  ~ juju status
Model       Controller  Cloud/Region    Version  SLA          Timestamp
my-cluster  canosphere  vsphere/Boston  3.4.0    unsupported  16:15:19-05:00

App      Version  Status  Scale  Charm    Channel      Rev  Exposed  Message
easyrsa  3.0.1    active      1  easyrsa  stable        55  no       Certificate Authority connected.
etcd     3.4.22   active      3  etcd     stable       760  no       Healthy with 3 known peers
k8s      1.29.3   active      3  k8s      latest/edge   31  no       Ready

Unit        Workload  Agent  Machine  Public address  Ports     Message
easyrsa/0*  active    idle   0        10.246.154.154            Certificate Authority connected.
etcd/0      active    idle   4        10.246.154.44   2379/tcp  Healthy with 3 known peers
etcd/1      active    idle   5        10.246.154.11   2379/tcp  Healthy with 3 known peers
etcd/2*     active    idle   6        10.246.154.42   2379/tcp  Healthy with 3 known peers
k8s/0*      active    idle   1        10.246.154.120  6443/tcp  Ready
k8s/1       active    idle   2        10.246.154.228  6443/tcp  Ready
k8s/2       active    idle   3        10.246.154.152  6443/tcp  Ready

Machine  State    Address         Inst id        Base          AZ  Message
0        started  10.246.154.154  juju-2a1cbe-0  ubuntu@22.04      poweredOn
1        started  10.246.154.120  juju-2a1cbe-1  ubuntu@22.04      poweredOn
2        started  10.246.154.228  juju-2a1cbe-2  ubuntu@22.04      poweredOn
3        started  10.246.154.152  juju-2a1cbe-3  ubuntu@22.04      poweredOn
4        started  10.246.154.44   juju-2a1cbe-4  ubuntu@22.04      poweredOn
5        started  10.246.154.11   juju-2a1cbe-5  ubuntu@22.04      poweredOn
6        started  10.246.154.42   juju-2a1cbe-6  ubuntu@22.04      poweredOn
```

<!-- LINKS -->

[easyrsa-charm]: https://charmhub.io/easyrsa 
[vault-charm]: https://charmhub.io/vault
[Juju setup]: https://juju.is/docs/juju/tutorial
