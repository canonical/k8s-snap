---
myst:
  html_meta:
    description: "How to configure MetalLB with multiple BGP peers in Canonical Kubernetes."
---

# How to configure multi-peer BGP

```{versionadded} release-1.36
```

```{note}
Multi-peer BGP via annotations is an **alpha** feature (`k8sd/v1alpha1`).
The interface may change without a deprecation period.
```

{{product}} supports configuring MetalLB with multiple BGP peers, each with its
own ASN and optional node selector. This is useful in multi-zone deployments
where each zone peers with a different top-of-rack router.

## Prerequisites

- A bootstrapped {{product}} cluster (see [Getting Started][getting-started-guide]).

## Configure multi-peer BGP

### Enable BGP mode

First enable the load balancer:

```bash
sudo k8s enable load-balancer
```

Then configure BGP:

```bash
sudo k8s set \
  load-balancer.bgp-mode=true \
  load-balancer.bgp-local-asn=65000 \
  load-balancer.cidrs=10.0.0.0/24
```

### Set the multi-peer annotation

Define peers in a YAML file and pass the whole file via the `annotations` key:

```bash
cat > bgp-peers.yaml << 'EOF'
k8sd/v1alpha1/metallb/bgp-peers: |
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
EOF
sudo k8s set annotations="$(cat bgp-peers.yaml)"
```

Setting this annotation **replaces** the single-peer typed keys. If both are
present, the annotation takes precedence and a warning appears in `k8s status`.

Supported fields per peer entry:

| Field | Required | Description |
|---|---|---|
| `peerAddress` | yes | IP address of the BGP peer router |
| `peerASN` | yes | ASN of the peer router (1–4294967295) |
| `myASN` | no | Local ASN for this peer (defaults to `bgp-local-asn`) |
| `peerPort` | no | TCP port (default: 179) |
| `nodeSelector` | no | `matchLabels` selector; omit to select all nodes |
| `bfdProfile` | no | Name of a `BFDProfile` to attach to this peer (requires `bgp-backend: frr-k8s`, see below) |

### Optionally advertise all pools

To advertise all IP address pools instead of only the named pool, add
`k8sd/v1alpha1/metallb/advertise-all-pools: "true"` to the annotations file
and re-apply:

```bash
cat > bgp-peers.yaml << 'EOF'
k8sd/v1alpha1/metallb/bgp-peers: |
  - peerAddress: 10.116.3.164
    peerASN: 65001
    myASN: 65000
k8sd/v1alpha1/metallb/advertise-all-pools: "true"
EOF
sudo k8s set annotations="$(cat bgp-peers.yaml)"
```

Or to set only the advertise flag without changing peers:

```bash
sudo k8s set annotations="k8sd/v1alpha1/metallb/advertise-all-pools: \"true\""
```

### Enable the FRR backend for BFD

MetalLB's default BGP implementation (`native`) does not support BFD
(Bidirectional Forwarding Detection). To use BFD, switch the speaker to the
`frr-k8s` backend with the `k8sd/v1alpha1/metallb/bgp-backend` annotation,
declare a `BFDProfile` with the `k8sd/v1alpha1/metallb/bfd-profiles`
annotation, and reference it from a peer's `bfdProfile` field:

```bash
cat > bgp-peers.yaml << 'EOF'
k8sd/v1alpha1/metallb/bgp-backend: frr-k8s
k8sd/v1alpha1/metallb/bgp-peers: |
  - peerAddress: 10.116.3.164
    peerASN: 65001
    myASN: 65000
    bfdProfile: fast-failover
k8sd/v1alpha1/metallb/bfd-profiles: |
  - name: fast-failover
    receiveInterval: 150
    transmitInterval: 150
    detectMultiplier: 3
EOF
sudo k8s set annotations="$(cat bgp-peers.yaml)"
```

`bgp-backend` accepts `native` (default) or `frr-k8s`. Setting `bfdProfile`
on any peer, or setting `bfd-profiles`, while the backend is `native` is
rejected — `k8s status` shows a degraded `load-balancer` state until either
the backend is switched to `frr-k8s` or the BFD configuration is removed.

Supported fields per `bfd-profiles` entry (omitted fields use MetalLB's
defaults):

| Field | Required | Description |
|---|---|---|
| `name` | yes | Unique `BFDProfile` name (lowercase DNS subdomain), referenced by a peer's `bfdProfile` |
| `receiveInterval` | no | Minimum receive interval in ms (10–60000, default: 300) |
| `transmitInterval` | no | Minimum transmit interval in ms (10–60000, default: 300) |
| `detectMultiplier` | no | Missed packets before the session is declared down (2–255, default: 3) |
| `echoInterval` | no | Minimum echo receive interval in ms (10–60000, default: 50) |
| `echoMode` | no | Enable echo mode (default: false; not supported on multi-hop) |
| `passiveMode` | no | Wait for the peer to start the session (default: false) |
| `minimumTtl` | no | Minimum expected TTL, multi-hop only (1–254, default: 254) |

{{product}} creates the profiles in the `metallb-system` namespace while
`bgp-mode` is enabled, and keeps them in sync with the annotation: changed
fields are applied, and entries removed from it are deleted once no peer
references them. A peer may also reference a `BFDProfile` created directly
with `k8s kubectl` in `metallb-system`, as long as its name does not collide
with a profile in the annotation.

With the `frr-k8s` backend, MetalLB requires all peers that select the same
node to use the same `myASN`.

To stop using BFD, remove `bfdProfile` from the peers and remove the
`bfd-profiles` annotation in the same change, keeping the peers themselves:

```bash
cat > bgp-peers.yaml << 'EOF'
k8sd/v1alpha1/metallb/bgp-peers: |
  - peerAddress: 10.116.3.164
    peerASN: 65001
    myASN: 65000
k8sd/v1alpha1/metallb/bfd-profiles: "-"
EOF
sudo k8s set annotations="$(cat bgp-peers.yaml)"
```

### Verify

```bash
sudo k8s status
```

The status line includes `(alpha)` when the annotation path is active, and
`frr-k8s backend` when the FRR backend is enabled:

```
load-balancer: enabled, BGP mode (alpha), frr-k8s backend
```

Inspect the resulting BGPPeer resources:

```bash
sudo k8s kubectl get bgppeers -A
```

Check BFD session state (once the backend and peer are both up). The FRR
process runs in its own `frr-k8s` DaemonSet, separate from the MetalLB
speaker pods:

```bash
sudo k8s kubectl get bfdprofiles -n metallb-system
sudo k8s kubectl get bgpsessionstates -A
```

The `BGP` and `BFD` columns of `bgpsessionstates` show `Established` and `Up`
once the sessions with the peer are running (`N/A` for peers without BFD).
They are refreshed every 2 minutes, so they can lag behind.

Check that MetalLB accepted the configuration. `k8s status` only reports
whether {{product}} applied it, not whether MetalLB accepted it:

```bash
sudo k8s kubectl get configurationstates -n metallb-system
```

Every entry should show `Valid`; `Invalid` entries include an error summary.

## Troubleshooting

If the annotation value is invalid, `k8s status` shows an error, for example:

```
load-balancer: Failed to deploy MetalLB, the error was: failed to enable LoadBalancer: invalid BGP peers: neighbor[0]: peerASN 0 out of range [1, 4294967295]
```

Correct the annotation and the reconciler retries automatically. Invalid
annotations are rejected before anything is applied, so the previous
configuration keeps running until then.

MetalLB can also reject a configuration that {{product}} applied, for example
peers on the same node with different `myASN` on the `frr-k8s` backend, or a
`bfdProfile` naming a profile that does not exist. `k8s status` still reports
the load balancer as enabled; check `configurationstates` (see Verify).

### MetalLB rejects its configuration after switching to the native backend

A `BFDProfile` can be left behind in two cases:

- A peer that references it is removed in the same change as the profile,
  for example when removing that peer or turning `bgp-mode` off.
- The profile is removed from `bfd-profiles` while a peer still references
  it.

MetalLB refuses to delete a `BFDProfile` while a peer still uses it, and
{{product}} does not retry the deletion.

The leftover profile is harmless while the backend is `frr-k8s`. After
switching to the `native` backend, MetalLB rejects its whole configuration:
it stops announcing services, in BGP and L2 alike, and stops assigning IPs,
while `k8s status` still reports the load balancer as enabled.
`configurationstates` shows `Invalid` with the error:

```
bfd profiles section set
```

List the remaining profiles and delete them:

```bash
sudo k8s kubectl get bfdprofiles -n metallb-system
sudo k8s kubectl delete bfdprofile <name> -n metallb-system
```

MetalLB picks up the change immediately. To avoid this, remove the
`bfdProfile` references from the peers first (see above), then remove the
peers or switch backends.

Disabling the load balancer while BFD is configured can report a failure
mentioning `used by BGPPeer` once; the next reconcile removes MetalLB and its
remaining resources.

## Limitations

- The annotation value is write-only — inspect it directly with
  `k8s kubectl get node <node> -o yaml`.
- No multi-hop BGP support.
- Removing a peer together with the `BFDProfile` it references, or removing a
  profile that a peer still references, can leave the profile behind (see
  Troubleshooting).

## Next steps

- [Load-balancer explanation](/snap/explanation/networking.md#load-balancer)
- [How to use the default load balancer](default-loadbalancer.md)

<!-- LINKS -->
[getting-started-guide]: /snap/tutorial/getting-started
