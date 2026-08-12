---
myst:
  html_meta:
    description: "Reference documentation for Canonical Kubernetes charm cluster annotations."
---

# Annotations

The `k8s` charm can set the same [cluster annotations][snap-annotations] used
by the {{product}} snap, via the `cluster-annotations` config option.

```{note}
v1alpha annotations are experimental and subject to change or removal in
future {{product}} releases.
```

## Setting annotations

`cluster-annotations` accepts a space-separated list of `key=value` pairs:

```bash
juju config k8s cluster-annotations="k8sd/v1alpha1/cilium/tunnel-port=8473 k8sd/v1alpha1/csrsigning/auto-approve=true"
```

Annotations are only overwritten when `cluster-annotations` is non-empty.
Clearing the config (`juju config k8s --reset cluster-annotations`) leaves any
existing cluster annotations untouched rather than deleting them.

See the [snap annotations reference][snap-annotations] for the full list of
supported annotation keys and their values.

## Nested annotations

Some annotations, such as the [multi-peer BGP annotation][multi-peer-bgp],
require nested YAML rather than a flat `key=value` pair. `cluster-annotations`
cannot express this. Set these annotations by running the `k8s` CLI on a unit
instead:

```bash
juju exec --unit <k8s/unit#> -- sudo k8s set annotations="$(cat bgp-peers.yaml)"
```

<!-- LINKS -->

[snap-annotations]: /snap/reference/annotations.md
[multi-peer-bgp]: /snap/howto/networking/multi-peer-bgp.md
