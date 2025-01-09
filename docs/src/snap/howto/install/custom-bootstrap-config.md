# Custom bootstrap configuration

When creating a {{ product }} cluster that differs from the default
configuration you can choose to use a custom bootstrap configuration.
The CLI's interactive mode or a custom bootstrap configuration file allow you
to modify the configuration of the first node that will join your cluster. 

## Configuration options

Please consult the [reference page] for all of the
available configuration options and their defaults. 

``` {note}
Most of these configuration options are set during the initial bootstrapping
and cannot be modified afterward. Runtime changes may be unsupported and
could require deploying a new cluster. Refer to the reference page to
determine if an option allows later modifications.
```

## Interactive mode

The interactive mode allows for the selection of the built-in features, the pod
CIDR and the Service CIDR.

To bootstrap interactively, run:

```
sudo k8s bootstrap --timeout 10m --interactive
```

Here is an example custom configuration:

```
Which features would you like to enable? (network, dns, gateway, ingress, local-storage, load-balancer) [network, dns, gateway, local-storage]: network,ingress,dns
Please set the Pod CIDR: [10.1.0.0/16]: 10.1.0.0/16,fd01::/108
Please set the Service CIDR: [10.152.183.0/24]: 10.152.183.0/24,fd98::/108
```

The expected output shows your node's ip that will differ from this example:

```
Bootstrapping the cluster. This may take a few seconds, please wait.
Bootstrapped a new Kubernetes cluster with node address "192.122.3.111:6400".
The node will be 'Ready' to host workloads after the CNI is deployed successfully.
```

## Bootstrap configuration file

If your deployment requires a more fine-tuned configuration, use the bootstrap
configuration file. By default the network, dns and gateway features
are enabled, but the ingress, load-balancer and local-storage features are
disabled.


For this example, create a custom bootstrap configuration file that disables
the network feature:

```yaml
cat <<EOF > bootstrap.yaml
cluster-config:
  network:
    enabled: false
EOF
```

Then, apply the bootstrap configuration file:

```
sudo k8s bootstrap --file /path/to/bootstrap.yaml
```

To verify any changes to the built-in features run:

```
sudo k8s status
```

<!-- LINKS -->

[reference page]: /src/snap/reference/bootstrap-config-reference.md
