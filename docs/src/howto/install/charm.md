# Install Canonical Kubernetes from a charm

Canonical Kubernetes is packaged as a [charm], available from the 
snap store for all supported platforms.

## What you'll need

This guide assumes the following:

- The rest of this page assumes you already have Juju installed and have added [credentials] for a cloud and bootstrapped a controller.
- If you still need to do this, please take a look at the quickstart instructions, or, for custom clouds (OpenStack, MAAS), please consult the Juju documentation.

```{note}
If you cannot meet these requirements, please see the [Installing] page for alternative options.
```

## Check available channels (optional)

It is a good idea to check the available channels before installing the snap. Run the command:

```bash
juju info k8s
juju info k8s-worker
```

...which will output a list of currently available channels. See the [channels
page] for an explanation of the different types of channel.

## Deploying the charm

The charm can be installed with the juju command:

```bash
juju deploy k8s --channel=latest/edge
```

```{note}
The `latest/edge` channel is always under active development. This is where you will find the latest features but you may also experience instability.
```

## Bootstrap the cluster

Installing the k8s charm sets up all the parts required to run Kubernetes. One may
watch it progress using juju status

```bash
juju status --watch 1s
```

This command will output a message confirming the snap is install and the
cluster is bootstrapped. It is recommended to ensure that the cluster initialises 
properly and is running with no issues. Run the command:

Once the unit is active/idle, You'll know the cluster is installed.

## Expanding the cluster

At this point, you should have 1 control-plane node.  To expand, add more units
with the following command

```bash
juju add-unit k8s -n 2
```

This will create 2 more control-plane units clustered with the first.

Use `juju status` to watch these units approach active/idle

## Adding Workers

In many cases, one would wish for kubernetes worker only units in their cluster. 
Rather than adding more control-plane units, we'll deploy the `k8s-worker` charm.
After deploying them, integrate them with the control-plane units so they join 
the cluster.

```bash
juju deploy k8s-worker --channel=latest/edge -n 2
juju integrate k8s k8s-worker:cluster
```

Use `juju status` to watch these units approach the active/idle state. 

<!-- LINKS -->

[Installing]: ./index
[channels page]: ../../explanation/channels
[credentials]:   https://juju.is/docs/juju/credentials