---
myst:
  html_meta:
    description: "How to install Canonical Kubernetes from the k8s charm."
relatedlinks: "[Juju&#32;documentation](https://canonical.com/juju/docs)"
---

# How to install {{product}} from a charm

{{product}} is packaged as a
[charm](https://documentation.ubuntu.com/juju/3.6/reference/charm/), available
from Charmhub for all supported platforms.

## Prerequisites

This guide assumes the following:

- Juju is installed
- [Credentials](https://documentation.ubuntu.com/juju/3.6/reference/credential/) are added for a cloud
- A controller is bootstrapped
- If you are using the Juju 'localhost' cloud see
  [localhost instructions](install-lxd) instead.

If you still need to set up Juju, please take a look at the
[Juju tutorial](https://canonical.com/juju/docs/juju-cli/3.6/tutorial/), or,
for custom clouds (OpenStack, MAAS), please consult the
[Juju clouds documentation](https://canonical.com/juju/docs/juju-cli/3.6/reference/cloud/list-of-supported-clouds/).

If you cannot meet these requirements, please see the
[installation](./index) page for alternative options.

## Check available channels (optional)

It is a good idea to check the available channels before installing the charm.
Run the command:

```
juju info k8s
juju info k8s-worker
```

...which will output a list of currently available channels. See the [channels
page](../../explanation/channels) for an explanation of the different types of
channel.

## Deploy the charm

The charm can be installed with the `juju` command. Set `--channel` to the
version you picked in the previous step. Ensure you provide your control plane
node with enough resources with the `--constraints` parameter:

```{literalinclude} /_parts/install.md
:start-after: <!-- juju control constraints start -->
:end-before: <!-- juju control constraints end -->
```

## Verify the deployment

Installing the `k8s` charm sets up all the parts required to run Kubernetes.
You can watch the installation progress using `juju status`:

```
juju status --watch 1s
```

This command will output a message confirming the charm is deployed and the
cluster is bootstrapped. It is recommended to ensure that the cluster
initializes properly and is running with no issues.

Once the unit is active/idle, the cluster is installed.

## Expand the cluster

To expand your cluster beyond the current single control-plane node,
add more units with the following command:

```
juju add-unit k8s -n 2
```

This will create 2 more control-plane units clustered with the first.

Use `juju status` to watch these units approach active/idle.

For high availability, run three or more control-plane units. See the
[high availability](../../../snap/explanation/high-availability) explanation.

## Add workers

Worker units run workloads without the control plane components. To add
worker-only nodes, deploy the `k8s-worker` charm.

```{literalinclude} /_parts/install.md
:start-after: <!-- juju worker start -->
:end-before: <!-- juju worker end -->
```

The `k8s-worker` will not settle into a 'Ready' status until it has a
connection to the control plane. Integrate these new nodes with control-plane
units so they join the cluster.

```
juju integrate k8s k8s-worker:cluster
juju integrate k8s k8s-worker:containerd
juju integrate k8s k8s-worker:cos-tokens
```

Use `juju status` to watch these units approach the active/idle state.

To add more workers later, add units to the existing application:

```
juju add-unit k8s-worker -n 2
```

Because the `k8s-worker` application is already integrated with `k8s`, the new
units join the cluster without any further steps.
