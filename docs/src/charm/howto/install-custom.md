# Installing with custom configuration

This guide will walk you through deploying {{product}} using Juju with custom
configuration options.

## What you'll need

This guide assumes the following:
-  You have Juju installed on your system with your cloud credentials
configured and a controller bootstrapped
- A Juju model is created and selected

## Creating the configuration file

Before deploying the charm, create a YAML file with your desired configuration
options. Here's an example configuration, which for this guide we'll save as
`k8s-config.yaml`:

```yaml
k8s:
  # Specify the datastore type
  bootstrap-datastore: dqlite

  # Configure pod and service CIDR ranges
  bootstrap-pod-cidr: "192.168.0.0/16"
  bootstrap-service-cidr: "10.152.183.0/24"

  # Enable required features
  dns-enabled: true
  gateway-enabled: true
  ingress-enabled: true
  metrics-server-enabled: true

  # Configure DNS settings
  dns-cluster-domain: "cluster.local"
  dns-upstream-nameservers: "8.8.8.8 8.8.4.4"

  # Add custom node labels
  node-labels: "environment=production zone=us-east-1"

  # Configure local storage
  local-storage-enabled: true
  local-storage-reclaim-policy: "Retain"
```
You can find a full list of configuration options in the
[charm configurations] page.

```{note}
Remember that some configuration options can only be set during initial
deployment and cannot be changed afterward. Always review the
[charm configurations] documentation before deployment to ensure your settings
align with your requirements.
```

## Deploying the Charm with Custom Configuration

Deploy the `k8s` charm with your custom configuration:

```bash
juju deploy k8s --config ./k8s-config.yaml
```

## Bootstrap the Cluster

Monitor the installation progress:

```bash
juju status --watch 1s
```

Wait for the unit to reach the `active/idle` state, indicating that the
{{product}} cluster is ready.

<!-- LINKS -->
[charm configurations]: https://charmhub.io/k8s/configurations
