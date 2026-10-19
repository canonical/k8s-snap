---
myst:
  html_meta:
    description: "Reference documentation for Canonical Kubernetes k8s and k8s-worker charm configuration options."
---

# Charm configuration options reference

This reference lists the configuration options for the `k8s` and `k8s-worker`
charms. They are also available on [Charmhub](https://charmhub.io/k8s/configurations).

These options can be set at
various points during the charm's lifecycle. Some options are limited to the
bootstrap process, while others can be modified at any time. Options restricted
to bootstrap are explicitly noted in their descriptions and prefixed with
`bootstrap-`.

See the {ref}`How to configure your cluster <charm-configure-cluster>` how-to guide to see how to apply these configurations.

## Configuration options for `k8s` charm

```{include} /_parts/charms/k8s-config.md
```

## Configuration options for `k8s-worker` charm

```{include} /_parts/charms/k8s-worker-config.md
```
