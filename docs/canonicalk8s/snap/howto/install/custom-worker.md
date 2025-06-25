# How to join worker nodes with a custom configuration

When creating a {{ product }} cluster you may need to join a worker node with
a configuration that differs from the default. For example, the worker node
may need to use alternative certificates for security reasons or the worker
node may have specific networking requirements that must be configured at node
creation. Passing extra command line arguments or a configuration file
at cluster join allows you to modify the configuration of your worker node.

## Prerequisites

This guide assumes the following:

- A working Kubernetes cluster deployed with the `k8s` snap

## Generate worker join token

When generating a join token for a worker node, pass the `--worker`
parameter to the `get-join-token` command. Adding the
hostname when creating a worker join token is optional and is not included here.

```
sudo k8s get-join-token --worker
```

## Install the snap

On the new worker machine, install the snap:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

## Join the cluster

### Default configuration

To join the cluster with the default configuration, on the worker node use the
token generated from the output of the `get-join-token` command and run:

```
sudo k8s join-cluster <JOIN-TOKEN>
```

### Command line arguments

To discover the configuration options available as command line arguments when
joining the cluster, on the control node run:

```
sudo k8s join-cluster --help
```

You can then run the join the cluster with the token generated from the output
of the `get-join-token` command and any arguments you may need. For example, to
set the output formatting to JSON run:

```
sudo k8s join-cluster --output-format=json <JOIN-TOKEN>
```

### Configuration file

More configuration options are available when a configuration file is specified.
Please consult the [reference page] for all of the available configuration
options and their defaults.

In this example, the configuration file provided at cluster join will set the
proxy mode of the worker machine to `ipvs`.

Create a `custom_config.yaml` file that sets the intended custom configurations.

```
cat <<EOF > custom_config.yaml
extra-node-kube-proxy-args:
    "--proxy-mode" : "ipvs"
EOF
```

On the worker node, join the cluster with the token generated from the output of
the `get-join-token` command and the `custom_config.yaml` file.

```
sudo k8s join-cluster --file path/to/custom_config.yaml <JOIN-TOKEN>
```

## Verify worker join

After a few moments, the node should have joined the cluster with a success
message. Verify the node has joined the cluster by switching to the control
node and running:

```
sudo k8s kubectl get nodes
```

The output should list the worker node in a `Ready` state.

Also verify if any custom configuration has been applied to the worker.

<!-- LINKS -->

[reference page]: /snap/reference/config-files/worker-join-config.md
