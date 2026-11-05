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

### Define the peers

Define peers in a YAML file under the `k8sd/v1alpha1/metallb/bgp-peers`
annotation:

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
```

The supported fields per peer are listed in the
[annotations reference][bgp-peers-reference].

### Enable BGP mode

Enable the load balancer:

```bash
sudo k8s enable load-balancer
```

Then enable BGP mode and pass the peers file via the `annotations` key in the
same command. BGP mode requires at least one peer, so it cannot be enabled
before the peers are set:

```bash
sudo k8s set \
  load-balancer.bgp-mode=true \
  load-balancer.bgp-local-asn=65000 \
  load-balancer.cidrs=10.0.0.0/24 \
  annotations="$(cat bgp-peers.yaml)"
```

Once BGP mode is enabled, update the peers by editing `bgp-peers.yaml` and
running `sudo k8s set annotations="$(cat bgp-peers.yaml)"` again.

Each annotation passed to `k8s set` replaces its previous value as a whole, so
`bgp-peers` must always list every peer. Annotations left out of the command
keep their current value.

The annotation **replaces** the single-peer typed keys
(`load-balancer.bgp-peer-*`). If both are present, the annotation takes
precedence and a warning appears in `k8s status`.

### Optionally advertise all pools

To advertise all IP address pools instead of only the named pool, set the
`k8sd/v1alpha1/metallb/advertise-all-pools` annotation. The peers are left
unchanged:

```bash
sudo k8s set annotations="k8sd/v1alpha1/metallb/advertise-all-pools: \"true\""
```

### Enable BFD

[Bidirectional Forwarding Detection (BFD)][bfd-rfc] detects a failed link to a
peer within milliseconds, instead of waiting for the BGP hold timer to expire,
so the peer stops routing traffic to the failed node sooner. See
[MetalLB's BFD documentation][metallb-bfd] for details.

MetalLB announces service IPs through its speaker, a DaemonSet that runs a pod
on every node and holds the BGP sessions with the peers. The speaker's default
BGP implementation, [`native`][metallb-bgp], does not support BFD. To use BFD,
switch to the [`frr-k8s`][metallb-frr-k8s] backend, which runs the
[FRR][frr-bfd] routing suite in a separate [frr-k8s][frr-k8s] DaemonSet. Set
the backend with the `k8sd/v1alpha1/metallb/bgp-backend` annotation, declare a
BFD profile with the `k8sd/v1alpha1/metallb/bfd-profiles` annotation, and
reference it from a peer's `bfdProfile` field. For brevity, the examples below
show a single peer; list all of your peers:

```bash
cat > bgp-peers.yaml << 'EOF'
k8sd/v1alpha1/metallb/bgp-backend: frr-k8s
k8sd/v1alpha1/metallb/bfd-profiles: |
  - name: fast-failover
    receiveInterval: 150
    transmitInterval: 150
    detectMultiplier: 3
k8sd/v1alpha1/metallb/bgp-peers: |
  - peerAddress: 10.116.3.164
    peerASN: 65001
    myASN: 65000
    bfdProfile: fast-failover
EOF
sudo k8s set annotations="$(cat bgp-peers.yaml)"
```

The supported profile fields are listed in the
[annotations reference][bfd-profiles-reference]. A peer's `bfdProfile` must
name a profile that exists, otherwise MetalLB rejects the configuration.

BFD settings require the `frr-k8s` backend. If `bfd-profiles` or a peer's
`bfdProfile` is set while the backend is `native`, {{product}} rejects the
configuration, keeps the previous one running, and `k8s status` reports the
error, for example:

```
load-balancer: Failed to deploy MetalLB, the error was: failed to enable LoadBalancer: invalid BFD profiles: k8sd/v1alpha1/metallb/bfd-profiles requires k8sd/v1alpha1/metallb/bgp-backend annotation set to "frr-k8s"
```

{{product}} creates the profiles as `BFDProfile` resources in the
`metallb-system` namespace while `bgp-mode` is enabled, and keeps them in sync
with the annotation: changed fields are applied, and entries removed from it
are deleted once no peer references them. A peer may also reference a
`BFDProfile` created directly with `k8s kubectl` in `metallb-system`, as long
as its name does not collide with a profile in the annotation.

With the `frr-k8s` backend, MetalLB requires all peers that select the same
node to use the same `myASN`.

### Stop using BFD

Remove `bfdProfile` from the peers and remove the `bfd-profiles` annotation in
the same change, keeping the peers themselves:

```bash
cat > bgp-peers.yaml << 'EOF'
k8sd/v1alpha1/metallb/bfd-profiles: "-"
k8sd/v1alpha1/metallb/bgp-peers: |
  - peerAddress: 10.116.3.164
    peerASN: 65001
    myASN: 65000
EOF
sudo k8s set annotations="$(cat bgp-peers.yaml)"
```

Only then remove peers, disable `bgp-mode` or switch the backend back to
`native`. MetalLB refuses to delete a `BFDProfile` while a peer still uses it,
so doing these in one step leaves the profile behind (see
[Troubleshooting](#troubleshooting)).

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

With the `frr-k8s` backend, check that the `frr-k8s` pods are running, then
check the BFD profiles and session state:

```bash
sudo k8s kubectl get pods -n metallb-system -l app.kubernetes.io/component=frr-k8s
sudo k8s kubectl get bfdprofiles -n metallb-system
sudo k8s kubectl get bgpsessionstates -A
```

The `BGP` and `BFD` columns of `bgpsessionstates` show `Established` and `Up`
once the sessions with the peer are running (`N/A` for peers without BFD).
They are refreshed every 2 minutes, so they can lag behind.

`k8s status` only reports whether {{product}} applied the configuration, not
whether MetalLB accepted it. Check that with:

```bash
sudo k8s kubectl get configurationstates -n metallb-system
```

Every entry should show `Valid`; `Invalid` entries include an error summary.

## Troubleshooting

If an annotation is invalid, `k8s status` shows the error (see the example
above). Correct the annotation and the reconciler retries automatically.

If `configurationstates` shows `Invalid`, MetalLB rejected a configuration
that {{product}} applied. The error summary names the cause, such as peers on
the same node with different `myASN` on the `frr-k8s` backend. Fix the
annotation accordingly.

If it shows `Invalid` with the error `bfd profiles section set` after switching
to the `native` backend, a `BFDProfile` was left behind (see
[Stop using BFD](#stop-using-bfd)). MetalLB then stops announcing services and
assigning IPs until the profile is deleted:

```bash
sudo k8s kubectl get bfdprofiles -n metallb-system
sudo k8s kubectl delete bfdprofile <name> -n metallb-system
```

Disabling the load balancer while BFD is configured can report a failure
mentioning `used by BGPPeer` once; the next reconcile removes MetalLB and its
remaining resources.

## Limitations

- The annotation value is write-only — inspect it directly with
  `k8s kubectl get node <node> -o yaml`.
- No multi-hop BGP support.

## Next steps

- [Annotations reference][annotations-reference]
- [Load-balancer explanation](/snap/explanation/networking.md#load-balancer)
- [How to use the default load balancer](default-loadbalancer.md)

<!-- LINKS -->
[getting-started-guide]: /snap/tutorial/getting-started
[annotations-reference]: /snap/reference/annotations
[bgp-peers-reference]: /snap/reference/annotations.md#k8sdv1alpha1metallbbgp-peers
[bfd-profiles-reference]: /snap/reference/annotations.md#k8sdv1alpha1metallbbfd-profiles
[bfd-rfc]: https://datatracker.ietf.org/doc/html/rfc5880
[metallb-bfd]: https://metallb.io/configuration/#enabling-bfd-support-for-bgp-sessions
[metallb-bgp]: https://metallb.io/concepts/bgp/
[metallb-frr-k8s]: https://metallb.io/concepts/bgp/#frr-k8s-mode
[frr-k8s]: https://github.com/metallb/frr-k8s
[frr-bfd]: https://docs.frrouting.org/en/latest/bfd.html
