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

## Command line arguments

To discover the configuration options available as command line arguments,
on the control node run:

```
sudo k8s join-cluster --help
```

In this example, the name of the new worker node joining the cluster is
specified through command line arguments.

If we do not specify the node name upon creating a worker join token on 
the control plane node the worker node will appear in the cluster with
the default hostname. In this example, we
include the name of the worker node: `custom-worker`. To generate the 
join token for a worker add the `--worker` option.

```
sudo k8s get-join-token custom-worker --worker
```

On the new worker machine, install the snap:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

Join the cluster with the token generated from the output of the
`get-join-token` command and specify the same `--name` we want the worker node
to be called. This must match the name used in the `get-join-token` command.

```
sudo k8s join-cluster --name custom-worker <JOIN-TOKEN>
```

After a few moments, the node should have joined the cluster with a success
message. Verify the node has joined the cluster with the custom name by
switching to the control node and running:

```
sudo k8s kubectl get nodes
```

The output should list the `custom-worker` node in a `Ready` state.

## Configuration file

More configuration options are available when a configuration file is specified.
Please consult the [reference page] for all of the available configuration
options and their defaults.

In this example, the configuration file provided at cluster join will set the
proxy mode of the worker machine to `ipvs`.

A join token must be generated on the control plane node of the cluster.
To generate the join token for a worker the `--worker` option is added. 
We will not specify the node name in this example.

```
sudo k8s get-join-token --worker
```

On the new worker machine, install the snap:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

Create a `custom_config.yaml` file that sets the intended custom configurations.

```
cat <<EOF > custom_config.yaml
extra-node-kube-proxy-args:
    "--proxy-mode" : "ipvs"
EOF
```

Join the cluster with the token generated from the output of the
`get-join-token` command and the `custom_config.yaml` file.

```
sudo k8s join-cluster --file path/to/custom_config.yaml <JOIN-TOKEN>
```

After a few moments, the node should have joined the cluster with a success
message. Verify the node has joined the cluster by switching to the control
node and running:

```
sudo k8s kubectl get nodes
```

The output should list the worker node as in a `Ready` state.

Also verify the proxy mode configuration has been applied to the worker node
by checking the logs of kube-proxy on the worker machine:

```
sudo journalctl -u snap.k8s.kube-proxy | grep ipvs
```

The output should show the proxy-mode is `ipvs`.

<!-- LINKS -->

[reference page]: /snap/reference/config-files/worker-join-config.md
