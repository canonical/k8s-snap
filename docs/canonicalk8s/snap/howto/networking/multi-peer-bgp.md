# How to configure multi-peer BGP (alpha)

```{note}
Multi-peer BGP configuration via annotations is an **alpha** feature
(`k8sd/v1alpha1`). The interface may change in future releases without a
deprecation period.
```

{{product}} supports configuring MetalLB with multiple BGP peers, each with its
own ASN and optional node selector. This is useful in multi-zone deployments
where each availability zone peers with a different top-of-rack router.

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine.
- You have a bootstrapped {{product}} cluster (see the [Getting
  Started][getting-started-guide] guide).
- BGP mode is enabled on the load balancer (`load-balancer.bgp-mode=true`).

## Overview

Single-peer BGP is configured with `k8s set load-balancer.bgp-*` keys.
Multi-peer BGP is configured separately via an annotation:

```
k8sd/v1alpha1/metallb/bgp-peers
```

Setting this annotation **replaces** the single-peer typed keys entirely. If
both are present, the annotation takes precedence and a warning is surfaced in
`k8s status`.

A second optional annotation controls whether MetalLB advertises all IP address
pools:

```
k8sd/v1alpha1/metallb/advertise-all-pools
```

## Configure multi-peer BGP

### Step 1 — Enable BGP mode

Enable the load balancer in BGP mode with a local ASN and your CIDRs:

```bash
sudo k8s set \
  load-balancer.enabled=true \
  load-balancer.bgp-mode=true \
  load-balancer.bgp-local-asn=65000 \
  load-balancer.cidrs=10.0.0.0/24
```

### Step 2 — Set the multi-peer annotation

Define each peer as a YAML list and set it via the `annotations` key. In this
example three availability zones (i1, i2, i3) each peer with a dedicated
top-of-rack router:

```bash
sudo k8s set 'annotations.k8sd/v1alpha1/metallb/bgp-peers=
- peerAddress: 10.116.3.164
  peerASN: 65001
  myASN: 65000
  nodeSelector:
    topology.kubernetes.io/zone: i1
- peerAddress: 10.116.3.165
  peerASN: 65002
  myASN: 65000
  nodeSelector:
    topology.kubernetes.io/zone: i2
- peerAddress: 10.116.3.166
  peerASN: 65003
  myASN: 65000
  nodeSelector:
    topology.kubernetes.io/zone: i3
'
```

Each peer entry supports the following fields:

| Field | Required | Description |
|---|---|---|
| `peerAddress` | yes | IP address of the BGP peer router |
| `peerASN` | yes | ASN of the peer router (1–4294967295) |
| `myASN` | no | Local ASN for this peer (defaults to `bgp-local-asn`) |
| `peerPort` | no | TCP port for the BGP session (default: 179) |
| `nodeSelector` | no | `matchLabels` selector; omit to select all nodes |

### Step 3 — Optionally advertise all pools

To have MetalLB advertise all IP address pools (instead of only the named pool
for this load balancer), set:

```bash
sudo k8s set 'annotations.k8sd/v1alpha1/metallb/advertise-all-pools=true'
```

### Step 4 — Verify

Check that the load balancer is healthy:

```bash
sudo k8s status
```

The status message will include `(alpha)` when the multi-peer annotation path is
active:

```
load-balancer: enabled, BGP mode (alpha)
```

Inspect the resulting BGPPeer custom resources:

```bash
sudo k8s kubectl get bgppeers -A
```

You should see one `BGPPeer` CR per entry in the annotation, named
`<cluster-name>`, `<cluster-name>-1`, `<cluster-name>-2`, etc.

## Modify or remove peers

To update the peer list, re-run `k8s set` with the new YAML value. The
annotation **replaces** the previous list atomically — there is no merge.

To revert to single-peer mode, remove the annotation and set the single-peer
typed keys:

```bash
sudo k8s set 'annotations.k8sd/v1alpha1/metallb/bgp-peers='
sudo k8s set \
  load-balancer.bgp-peer-address=10.0.0.1 \
  load-balancer.bgp-peer-asn=64512 \
  load-balancer.bgp-peer-port=179
```

## Troubleshooting

If the annotation value is invalid, {{product}} will not apply a broken
configuration. Run `k8s status` — the load-balancer line will show an error
message describing the validation failure, for example:

```
load-balancer: Failed to deploy MetalLB, the error was: invalid BGP peers: neighbor[0]: peerASN 0 out of range [1, 4294967295]
```

Correct the annotation value and the reconciler will retry automatically.

## Limitations (alpha)

- The annotation value is opaque to `k8s get` — annotations are write-only in
  the current implementation. Inspect the annotation directly with
  `k8s kubectl get node <node> -o yaml` or by re-reading your configuration
  source.
- No per-peer BFD or multi-hop support (out of scope).
- Per-peer `myASN` overrides require each node's BGP stack to present the
  correct local ASN; verify your router configuration accordingly.

## Next steps

- [Load-balancer explanation](/snap/explanation/networking.md#load-balancer)
- [How to use the default load balancer](default-loadbalancer.md)

<!-- LINKS -->
[getting-started-guide]: /snap/tutorial/getting-started
