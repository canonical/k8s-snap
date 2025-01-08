# Custom Bootstrap Configuration

When creating a {{ product }} cluster that differs from the default
configuration you can choose to use a custom bootstrap configuration.
The CLI's interactive mode or a custom bootstrap config file allow you to 
modify the configuration of the first node that will join your cluster. 

## Configuration Options

Please consult the [bootstrap-configuration-reference page] for all of the
available configuration options and their defaults. These configuration options
may only be adjusted on bootstrap and not after the cluster is bootstrapped.

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

The output for this example would be:

```
Bootstrapping the cluster. This may take a few seconds, please wait.
Bootstrapped a new Kubernetes cluster with node address "192.168.3.117:6400".
The node will be 'Ready' to host workloads after the CNI is deployed successfully.
```

## Bootstrap Configuration File

If your deployment requires a more fine tuned configuration, use the bootstrap
configuration file. A good starting point can be the default
[bootstrap-config-full.yaml].

For this example, create a custom bootstrap configuration file:

```yaml
cat <<EOF > bootstrap.yaml
cluster-config:
  network:
    enabled: false
EOF
```


Then, apply the bootstrap configuration file:

```
sudo k8s bootstrap --timeout 10m --file /path/to/bootstrap.yaml
```

To verify any changes to the built-in features run:

```
sudo k8s status
```

<!-- LINKS -->

[bootstrap-configuration-reference page]: /src/snap/reference/bootstrap-config-reference.md
[bootstrap-config-full.yaml]: https://raw.githubusercontent.com/canonical/k8s-snap/refs/heads/main/src/k8s/cmd/k8s/testdata/bootstrap-config-full.yaml