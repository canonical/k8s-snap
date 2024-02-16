# Install Canonical Kubernetes from a snap

Canonical Kubernetes is packaged as a [snap], available from the 
snap store for all supported platforms.

## What you'll need

This guide assumes the following:

- You are installing on Ubuntu 22.04 or later, **or** another OS which supports
  snap packages (see [snapd support])
- You have root or sudo access to the machine
- You have an internet connection
- The target machine has sufficient memory and disk space. To accommodate
  workloads, we recommend a system with at least 20G of disk space and 4G of
  memory.

```{note}
If you cannot meet these requirements, please see the [Installing] page for alternative options.
```

## Check available channels (optional)

It is a good idea to check the available channels before installing the snap. Run the command:

```bash
snap info k8s
```

...which will output a list of currently available channels. See the [channels page] for an explanation of the different types of channel.

## Install the snap

The snap can be installed with the snap command:

```bash
sudo snap install k8s --classic --channel=latest/edge
```

```{note}
In the pre-release phase, `latest/edge` is the only channel available. 
```

## Bootstrap the cluster

Installing the snap sets up all the parts required to run Kubernetes. The next step is to `bootstrap` the cluster to activate the services:

```bash
sudo k8s bootstrap
```

This command will output a message confirming the services have been started.

## Confirm the services are running

It is recommended to ensure that the cluster initialises properly and is running with no issues. Run the command:

```bash
sudo k8s status --wait-ready
```

This command will wait until the cluster indicates it is ready and then display the current status. The command will time-out if the cluster does not reach a ready state.

<!-- LINKS -->

[installing]: ./index
[channels page]: ../../explanation/channels
[snap]: https://snapcraft.io/docs
[snapd support]: https://snapcraft.io/docs/installing-snapd
