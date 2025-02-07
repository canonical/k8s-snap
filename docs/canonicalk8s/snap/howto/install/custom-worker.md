# How to join worker nodes with custom configuration

When creating a {{ product }} cluster you may need to join a worker node with
a configuration that differs from the default. For example, the worker node
may need to use alternative certificates for security reasons or the worker
node has specific networking requirements that must be configured at node creation.
Passing extra command line arguments or a configuration file
at cluster join allows you to modify the configuration of your worker node.

## Prerequisites

This guide assumes the following:

- A working Kubernetes cluster deployed with the `k8s` snap

## Configuration options

### Command line flags

To discover the configuration options available as command line arguments, on the control node run:

```
sudo k8s join-cluster --help
```

The options are:

- `address`: microcluster address or CIDR of the node
- `file` : path to the YAML file containing your custom cluster join configuration
- `h`, `help`: help for join-cluster
- `name`: node name
- `output-format`: set the output format to plain, json or yaml
- `timeout`: the max time to wait for the command to execute

### Configuration file

More configuration options are available when a configuration file is specified. Please consult the [reference page] for all of the
available configuration options and their defaults.

## Join a custom worker with command line arguments

In this example, the worker node will be named custom-worker in the cluster, rather than defaulting to hostname of the machine.

A join token must be generated on the control node of the cluster. It must be specified the join token is for a `--worker` node. The name given for the worker node
in this example is custom-worker which is not the default hostname.

```
sudo k8s get-join-token custom-worker --worker
```

On the new worker machine, install the snap:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

Join the cluster with the token generated from the output of the `get-join-token` command and specify the `--name` we want the worker node to be called. This must match the name used in the `get-join-token` command.

```
sudo k8s join-cluster --name custom-worker <JOIN-TOKEN>
```

After a few moments, the node should have joined the cluster with a success message. Verify the node has joined the cluster with the custom name by switching to the control node and running:

```
sudo k8s kubectl get nodes
```

The output should list the `custom-worker` node in a `Ready` state.

## Join a custom worker with a configuration file

A join token must be generated on the control node of the cluster. It must be specified the join token is for a `--worker` node. The name given for the worker node
in this example is worker-machine which is the default hostname.

```
sudo k8s get-join-token worker-machine --worker
```

On the new worker machine, install the snap:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

Create a `custom_config.yaml` file that...

```
cat <<EOF > custom_config.yaml

EOF
```

Join the cluster with the token generated from the output of the `get-join-token` command and the `custom_config.yaml` file.

```
sudo k8s join-cluster --file path/to/custom_config.yaml <JOIN-TOKEN>
```

After a few moments, the node should have joined the cluster with a success message. Verify the node has joined the cluster by switching to the control node and running:

```
sudo k8s kubectl get nodes
```

The output should list the worker-machine node as in a `Ready` state.

Also verify the config file change we made has been implemented on the worker node.

```
what did we change?
```
//verify that it has the param we set above

<!-- LINKS -->

[reference page]: /snap/reference/config-files/worker-join-config.md
