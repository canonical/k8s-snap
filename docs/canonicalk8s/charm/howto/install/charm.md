# How to install {{product}} from a charm

{{product}} is packaged as a [charm], available from Charmhub for all
supported platforms.

## Prerequisites

This guide assumes the following:

- The rest of this page assumes you already have Juju installed and have added
  [credentials] for a cloud and bootstrapped a controller.
- If you still need to do this, please take a look at the quick-start
  instructions, or, for custom clouds (OpenStack, MAAS), please consult the
  [Juju documentation][juju].
- You are not using the Juju 'localhost' cloud (see [localhost
  instructions][localhost] for this).

```{note}
If you cannot meet these requirements, please see the [Installing][] page for
alternative options.
```

## Check available channels (optional)

It is a good idea to check the available channels before installing the charm.
Run the command:

```
juju info k8s
juju info k8s-worker
```

...which will output a list of currently available channels. See the [channels
page][channels] for an explanation of the different types of channel.

## Deploying the charm

The charm can be installed with the `juju` command:

```{literalinclude} /src/_parts/install.md
:start-after: <!-- juju control start -->
:end-before: <!-- juju control end -->
```

## Bootstrap the cluster

Installing the `k8s` charm sets up all the parts required to run Kubernetes.
You can watch the installation progress using juju status:

```
juju status --watch 1s
```

This command will output a message confirming the charm is deployed and the
cluster is bootstrapped. It is recommended to ensure that the cluster initialises
properly and is running with no issues.

Once the unit is active/idle, You'll know the cluster is installed.

## Expanding the cluster

At this point, you should have one control-plane node. To expand your cluster,
add more units with the following command

```
juju add-unit k8s -n 2
```

This will create 2 more control-plane units clustered with the first.

Use `juju status` to watch these units approach active/idle

## Adding Workers

In many cases, it is desirable to have additional 'worker only' units in the cluster.
Rather than adding more control-plane units, we'll deploy the `k8s-worker` charm.
After deployment, integrate these new nodes with control-plane units so they join
the cluster.


```{literalinclude} /src/_parts/install.md
:start-after: <!-- juju worker start -->
:end-before: <!-- juju worker end -->
:append: juju integrate k8s k8s-worker:cluster
```

Use `juju status` to watch these units approach the active/idle state.

<!-- LINKS -->

[Installing]:    ./index
[channels]:      ../../explanation/channels
[credentials]:   https://juju.is/docs/juju/credentials
[juju]:          https://juju.is/docs/juju/install-juju
[charm]:         https://juju.is/docs/juju/charmed-operator
[localhost]:     install-lxd
