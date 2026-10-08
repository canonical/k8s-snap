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
create a `BFDProfile`, and reference it from a peer's `bfdProfile` field:

```bash
cat > bgp-peers.yaml << 'EOF'
k8sd/v1alpha1/metallb/bgp-backend: frr-k8s
k8sd/v1alpha1/metallb/bgp-peers: |
  - peerAddress: 10.116.3.164
    peerASN: 65001
    myASN: 65000
    bfdProfile: fast-failover
EOF
sudo k8s set annotations="$(cat bgp-peers.yaml)"

sudo k8s kubectl apply -f - << 'EOF'
apiVersion: metallb.io/v1beta1
kind: BFDProfile
metadata:
  name: fast-failover
  namespace: metallb-system
spec:
  receiveInterval: 150
  transmitInterval: 150
  detectMultiplier: 3
EOF
```

`bgp-backend` accepts `native` (default) or `frr-k8s`. Setting `bfdProfile`
on any peer while the backend is `native` is rejected — `k8s status` shows a
degraded `load-balancer` state until either the backend is switched to
`frr-k8s` or the `bfdProfile` is removed.

`BFDProfile` resources are not managed by {{product}} — create, update, and
delete them directly with `k8s kubectl`, same as any other MetalLB CRD.

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
sudo k8s kubectl get bfdprofiles -A
sudo k8s kubectl -n metallb-system logs -l app.kubernetes.io/component=frr-k8s -c frr | grep -i bfd
```

## Troubleshooting

If the annotation value is invalid, `k8s status` shows an error, for example:

```
load-balancer: Failed to deploy MetalLB, the error was: invalid BGP peers: neighbor[0]: peerASN 0 out of range [1, 4294967295]
```

Correct the annotation and the reconciler retries automatically.

## Limitations

- The annotation value is write-only — inspect it directly with
  `k8s kubectl get node <node> -o yaml`.
- No multi-hop BGP support.
- `BFDProfile` resources must be created and managed directly with
  `k8s kubectl`; {{product}} only lets a peer reference one by name.

## Next steps

- [Load-balancer explanation](/snap/explanation/networking.md#load-balancer)
- [How to use the default load balancer](default-loadbalancer.md)

<!-- LINKS -->
[getting-started-guide]: /snap/tutorial/getting-started
