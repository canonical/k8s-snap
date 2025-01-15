# How to integrate {{product}} with etcd

Integrating [etcd][] with your {{product}} deployment provides a
robust, distributed key-value store that is essential for storing critical
data needed for Kubernetes' clustering operations. This guide will walk you
through the process of deploying {{product}} with an external etcd
cluster.

## Prerequisites

- A Juju controller with access to a cloud environment (see the [Juju setup]
  guide for more information).

```{warning} Once you deploy your {{product}} cluster with a
particular datastore, you cannot switch to a different datastore
post-deployment. Planning for your datastore needs ahead of time is
crucial, particularly if you opt for an external datastore like **etcd**.
```

## Preparing the Deployment

1. **Creating the Deployment Model**:
  Begin by creating a Juju model specifically for your {{product}}
  cluster deployment.

  ```bash
  juju add-model my-cluster
  ```

2. **Deploying Certificate Authority**:
  etcd requires a secure means of communication between its components.
  Therefore, we require a certificate authority such as [EasyRSA][easyrsa-charm]
  or [Vault][vault-charm]. Check the respective charm documentation for detailed
  instructions on how to deploy a certificate authority. In this guide, we will
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

    This setup is straightforward but not recommended for production
    environments due to a lack of high availability.

- **High Availability Setup**:
  - For environments where high availability is crucial, deploy etcd across at
    least three nodes:

    ```bash
    juju deploy etcd -n 3
    ```

    This ensures that your etcd cluster remains available even if one node
    fails.

## Integrating etcd with EasyRSA

Now you have to integrate etcd with your certificate authority. This will issue
the required certificates for secure communication between etcd and your
{{product}} cluster:

```bash
juju integrate etcd easyrsa
```

## Deploying {{product}}

Deploy the control plane units of {{product}} with the command:

```bash
juju deploy k8s --config bootstrap-datastore=etcd -n 3
```

This command deploys 3 units of the {{product}} control plane (`k8s`)
and configures them to use **etcd** as the backing datastore, ensuring high
availability.

```{important}
Remember to run `juju expose k8s`. This will open the required
ports to reach your cluster from outside.
```

## Integrating {{product}} with etcd

Now that we have both the etcd datastore deployed alongside our Canonical
Kubernetes cluster, it is time to integrate our cluster with our etcd datastore.

```bash
juju integrate k8s etcd
```

This step integrates the k8s charm (Control Plane units)  with the etcd hosts,
allowing the Kubernetes cluster to use the etcd units as an external
datastore.

## Final Steps

**Verify the Deployment**: After completing the deployment, it's essential
to verify that all components are functioning correctly. Use the `juju status`
command to inspect the current status of your cluster.

<!-- markdownlint-disable -->
```
Model       Controller  Cloud/Region   Version  SLA          Timestamp
my-cluster  canonicaws  aws/us-east-1  3.4.2    unsupported  16:02:18-05:00

App      Version  Status  Scale  Charm    Channel        Rev  Exposed  Message
easyrsa  3.0.1    active      1  easyrsa  latest/stable   58  no       Certificate Authority connected.
etcd     3.4.22   active      3  etcd     latest/stable  760  no       Healthy with 3 known peers
k8s      1.29.4   active      3  k8s      latest/edge     33  yes      Ready

Unit        Workload  Agent  Machine  Public address  Ports     Message
easyrsa/0*  active    idle   0        35.172.230.66             Certificate Authority connected.
etcd/0*     active    idle   1        34.204.173.161  2379/tcp  Healthy with 3 known peers
etcd/1      active    idle   2        54.225.4.183    2379/tcp  Healthy with 3 known peers
etcd/2      active    idle   3        3.208.15.61     2379/tcp  Healthy with 3 known peers
k8s/0       active    idle   4        54.89.153.117   6443/tcp  Ready
k8s/1*      active    idle   5        3.238.230.3     6443/tcp  Ready
k8s/2       active    idle   6        34.229.202.243  6443/tcp  Ready

Machine  State    Address         Inst id              Base          AZ          Message
0        started  35.172.230.66   i-0b6fc845c28864913  ubuntu@22.04  us-east-1f  running
1        started  34.204.173.161  i-05439714c88bea35f  ubuntu@22.04  us-east-1f  running
2        started  54.225.4.183    i-07ecf97ed29860334  ubuntu@22.04  us-east-1c  running
3        started  3.208.15.61     i-0be91170809d7dccc  ubuntu@22.04  us-east-1b  running
4        started  54.89.153.117   i-07906e76071b69721  ubuntu@22.04  us-east-1c  running
5        started  3.238.230.3     i-0773583e7a5fbf07e  ubuntu@22.04  us-east-1f  running
6        started  34.229.202.243  i-0f03b9833a338380c  ubuntu@22.04  us-east-1b  running
```
<!-- markdownlint-restore -->

<!-- LINKS -->
[etcd]: https://etcd.io
[easyrsa-charm]: https://charmhub.io/easyrsa
[vault-charm]: https://charmhub.io/vault
[Juju setup]: https://juju.is/docs/juju/tutorial
