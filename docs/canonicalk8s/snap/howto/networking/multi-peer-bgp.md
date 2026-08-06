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

### Verify

```bash
sudo k8s status
```

The status line includes `(alpha)` when the annotation path is active:

```
load-balancer: enabled, BGP mode (alpha)
```

Inspect the resulting BGPPeer resources:

```bash
sudo k8s kubectl get bgppeers -A
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
- No per-peer BFD or multi-hop support.

## Next steps

- [Load-balancer explanation](/snap/explanation/networking.md#load-balancer)
- [How to use the default load balancer](default-loadbalancer.md)

<!-- LINKS -->
[getting-started-guide]: /snap/tutorial/getting-started
