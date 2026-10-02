<!--
To start a new proposal, create a copy of this template on this directory and
fill out the sections below.
-->

# Proposal information

<!-- Index number -->
- **Index**: 003

<!-- Status -->
- **Status**: **DRAFTING** <!-- **DRAFTING**/**ACCEPTED**/**REJECTED** -->

<!-- Short description for the feature -->
- **Name**: MetalLB FRR BGP backend and per-peer BFD

<!-- Owner name and github handle -->
- **Owner**: TBD

# Proposal Details

## Summary

MetalLB's speaker currently always runs with the `native` (GoBGP) BGP
implementation. This proposal adds an alpha annotation that switches the
speaker to the `frr-k8s` backend, and extends the existing multi-peer BGP
annotation (`k8sd/v1alpha1/metallb/bgp-peers`, introduced in release-1.36)
so that individual peers can reference a `BFDProfile`. BFD (Bidirectional
Forwarding Detection) is not supported by the `native` backend; it requires
an FRR-based one.

## Rationale

Operators running MetalLB BGP mode on links where sub-second failure
detection matters (e.g. ECMP fabrics, redundant top-of-rack routers) want BFD
alongside BGP so that peer failure is detected in milliseconds instead of
relying on the BGP hold timer (tens of seconds). MetalLB's `native` backend
(a Go implementation of BGP bundled in the speaker) does not implement BFD.
The only way to get BFD is to run the speaker against an FRR-based backend.

The vendored MetalLB chart (`metallb-0.16.1.tgz`) already bundles the
`frr-k8s` subchart and the `BFDProfile` CRD, but k8sd currently forces
`frrk8s.enabled: false` and never sets `speaker.frr.enabled: true`
([pkg/k8sd/features/metallb/loadbalancer.go, `enableLoadBalancer`][impl-main]).
There is no way for a user to opt in today.

Historically there was a `// TODO(neoaggelos): make frr enable/disable
configurable through an annotation` comment attached to the legacy
`speaker.frr.enabled` chart value. That comment was dropped from the
codebase without the feature being implemented. This proposal supersedes
that TODO: the legacy embedded `speaker.frr` mode is marked
`DEPRECATED ... will be removed in a future release` in the vendored
chart's `values.yaml` (chart version 0.16.1), in favor of `frrk8s`. We
target `frrk8s`, not the deprecated path, to avoid shipping a feature that
upstream is about to delete.

Example usage after this proposal:

```bash
sudo k8s set annotations="$(cat <<'EOF'
k8sd/v1alpha1/metallb/bgp-backend: frr-k8s
k8sd/v1alpha1/metallb/bgp-peers: |
  - peerAddress: 10.116.3.164
    peerASN: 65001
    myASN: 65000
    bfdProfile: fast-failover
EOF
)"
sudo k8s kubectl apply -f - <<'EOF'
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

## User facing changes

- New annotation `k8sd/v1alpha1/metallb/bgp-backend`, accepted values
  `native` (default, current behavior) and `frr-k8s`. Any other value is
  rejected and surfaces as a degraded `load-balancer` status, same pattern
  as the existing `bgp-peers` annotation.
- The existing `k8sd/v1alpha1/metallb/bgp-peers` annotation schema gains an
  optional `bfdProfile` field per peer entry
  ([doc table][doc-table]). Setting `bfdProfile` on any peer while
  `bgp-backend` is not `frr-k8s` is rejected with an error (BFD requires the
  FRR backend).
- `k8s status` / `load-balancer` feature status message gains a backend
  qualifier when `frr-k8s` is active, e.g.
  `enabled, BGP mode (alpha), frr-k8s backend`, mirroring the existing
  `(alpha)` suffix used for the peers annotation.
- `BFDProfile` custom resources are **not** managed by k8sd/the chart. Users
  create and manage them directly with `k8s kubectl apply`, and reference
  them by name via `bfdProfile`. This keeps scope small and avoids
  duplicating a CRD's full schema into our annotation YAML.
- No change for users who don't set `bgp-backend` — behavior defaults to
  today's `native` backend.

## Alternative solutions

- **Boolean `frr-enabled` annotation instead of a `bgp-backend` enum.**
  Rejected: a boolean reads as "enable the FRR mode", but there are two FRR
  modes in the chart (legacy `speaker.frr`, deprecated, and `frrk8s`,
  current). An enum makes the selected backend unambiguous and leaves room
  for a future `native-bfd` style addition without a breaking rename.
- **Target the legacy `speaker.frr.enabled` value (literal TODO intent).**
  Rejected: upstream marks this deprecated and slated for removal in the
  vendored chart version already in the snap. Implementing against it would
  mean reverting the toggle shortly after a chart bump.
- **Let k8sd manage `BFDProfile` resources declaratively (own annotation
  with timers etc.), instead of reference-only.** Rejected for this
  iteration: duplicates the CRD's schema in our own YAML dialect for
  marginal benefit; users can already apply `BFDProfile` objects directly
  with `k8s kubectl`. Can be revisited if there's demand for keeping BFD
  profile config inside `k8s set`/annotations instead of raw manifests.

## Out of scope

- Declarative management of `BFDProfile` CRs (see Alternative solutions).
- Cilium load balancer BGP backend (`pkg/k8sd/features/cilium/loadbalancer.go`)
  — Cilium's BGP control plane does not use MetalLB/FRR and is unaffected.
- Removing/migrating the legacy `speaker.frr` value — it stays hardcoded to
  `false`; this proposal only wires up `frrk8s`.
- Multi-hop BGP, which is a separate, currently-unimplemented limitation
  called out in the existing [multi-peer BGP how-to][multi-peer-doc].

# Implementation Details

## API Changes

New annotation keys in `k8s-snap-api`
(`api/annotations/metallb/metallb.go`, alongside the existing
`AnnotationBGPPeers` / `AnnotationAdvertiseAllPools` added in
[k8s-snap-api@c23486a][api-main]):

```go
// AnnotationBGPBackend selects the BGP implementation used by the MetalLB
// speaker. Supported values: "native" (default) and "frr-k8s" (required
// for BFD). Any other value is rejected.
AnnotationBGPBackend = "k8sd/v1alpha1/metallb/bgp-backend"
```

The `bgp-peers` YAML schema (parsed in `k8sd`, not part of the Go API
package itself) gains an optional `bfdProfile` string field per peer; no
new Go API type is required since peers are already free-form YAML.

No gRPC/REST k8sd API changes — annotations are passed through the existing
generic `types.Annotations` map already plumbed into
`ApplyLoadBalancer(ctx, snap, loadbalancer, network, annotations)`.

## CLI Changes

None. Configuration continues to go through `k8s set annotations=...`.

## Database Changes

None.

## Configuration Changes

- `pkg/k8sd/features/metallb/loadbalancer.go`:
  - Add `backendFromAnnotations(annotations types.Annotations) (frrk8sEnabled bool, active bool, err error)`,
    mirroring `neighborsFromAnnotations`. Validates the value is `""`,
    `"native"`, or `"frr-k8s"`.
  - Replace the hardcoded `"frrk8s": map[string]any{"enabled": false}` in
    `enableLoadBalancer`'s `metalLBValues` with the parsed value. Leave
    `speaker.frr.enabled` hardcoded `false` (legacy path untouched).
  - Add `bfdProfile string` to the `bgpNeighbor` struct and to the
    `peerYAML` struct parsed from the `bgp-peers` annotation.
  - In `enableLoadBalancer`, after resolving `neighbors`, validate: if any
    neighbor has a non-empty `bfdProfile` and the backend is not
    `frr-k8s`, return an error (`"bfdProfile requires bgp-backend: frr-k8s"`).
  - `buildLoadBalancerValues`: emit `"bfdProfile": n.bfdProfile` in the
    neighbor map when non-empty (same optional-field-omission pattern
    already used for `myASN` / `nodeSelector`).
  - `ApplyLoadBalancer`: extend the BGP status message to append
    `, frr-k8s backend` when active, alongside the existing `(alpha)`
    suffix logic for the peers annotation.
- `k8s/manifests/charts/ck-loadbalancer/templates/metallb/bgp-policy.yaml`
  (in this repo, `k8s-snap`): render `spec.bfdProfile` on the `BGPPeer`
  when `$n.bfdProfile` is set.
- `k8s/manifests/charts/ck-loadbalancer/values.schema.json`: add
  `bfdProfile: {"type": "string"}` to `bgp.neighbors.items.properties`.

Reference for current state (pre-change):
[k8sd@2430e1b, pkg/k8sd/features/metallb/loadbalancer.go][impl-main].

## Documentation Changes

- Extend `docs/canonicalk8s/snap/howto/networking/multi-peer-bgp.md`:
  - Document `bgp-backend` annotation and its two values.
  - Add `bfdProfile` to the peer field table.
  - Add a worked example creating a `BFDProfile` and referencing it.
  - Remove "No per-peer BFD or multi-hop support" from Limitations (keep
    the multi-hop limitation, drop the BFD half).
- No change needed to `default-loadbalancer.md` beyond the existing
  cross-link to the multi-peer-bgp how-to.

## Testing

- `k8sd` unit tests in `pkg/k8sd/features/metallb/loadbalancer_test.go`:
  - `backendFromAnnotations`: valid values (`""`, `native`, `frr-k8s`),
    invalid value error.
  - `buildLoadBalancerValues`: neighbor map includes/omits `bfdProfile`.
  - `ApplyLoadBalancer` end-to-end (mirroring existing
    `TestApplyLoadBalancerWithAnnotations` cases): backend annotation flips
    `frrk8s.enabled` in the first `helm.Apply` call's values; `bfdProfile`
    without `frr-k8s` backend returns an error and a degraded status;
    `bfdProfile` with `frr-k8s` backend renders on the neighbor map.
- No new integration test is strictly required for the toggle itself (no
  real FRR/BFD peer available in CI), but
  `tests/integration/tests/test_loadbalancer.py` should get a smoke case
  that sets `bgp-backend: frr-k8s` and asserts the `frr-k8s` speaker
  container/pod comes up healthy (not full BFD session validation).

## Considerations for backwards compatibility

- Default behavior (no `bgp-backend` annotation) is unchanged: `native`
  backend, `frrk8s.enabled: false`, identical to today.
- Existing `bgp-peers` entries without `bfdProfile` are unaffected; the
  field is purely additive and optional.
- Older `k8sd` talking to a newer `k8s-snap-api` (or vice versa) simply
  won't recognize the new annotation key/field and falls back to current
  `native` behavior — no breaking change to the annotation map's wire
  format (`map[string]string`).

## Implementation notes and guidelines

- Must land in three repos in order: `k8s-snap-api` (new annotation
  constant) → `k8sd` (consumes it, add a temporary `replace` directive in
  `k8sd/go.mod` pointing at a local `k8s-snap-api` checkout during
  development, per this repo's root `AGENTS.md`) → `k8s-snap` (chart
  template + schema + docs). Remove the `replace` directive before merging
  the `k8sd` PR.
- Current code pointers (pre-change):
  - [`k8sd@2430e1b`, `pkg/k8sd/features/metallb/loadbalancer.go`][impl-main]
  - [`k8s-snap-api@c23486a`, `api/annotations/metallb/metallb.go`][api-main]
  - `k8s-snap@2947e29`, `k8s/manifests/charts/ck-loadbalancer/templates/metallb/bgp-policy.yaml`
  - Vendored chart reference: `k8s/manifests/charts/metallb-0.16.1.tgz`,
    `metallb/values.yaml` lines ~330-379 (`speaker.frr` deprecation notice,
    `frrk8s` block), `metallb/charts/crds/templates/crds.yaml` (`BFDProfile`,
    `BGPPeer.spec.bfdProfile` schema).
- `frrk8s.enabled` and `speaker.frr.enabled` are mutually exclusive in the
  upstream chart template (`metallb/templates/speaker.yaml` `fail`s if both
  are true) — the implementation must never set both, which is naturally
  satisfied by always hardcoding `speaker.frr.enabled: false`.

<!-- LINKS -->
[impl-main]: https://github.com/canonical/k8sd/blob/2430e1b8ae84ebef5f21b859a779a44f8acdc292/pkg/k8sd/features/metallb/loadbalancer.go
[api-main]: https://github.com/canonical/k8s-snap-api/blob/c23486a68644d6a2a381f9c731d26e1b2c5bf513/api/annotations/metallb/metallb.go
[doc-table]: /snap/howto/networking/multi-peer-bgp.md
[multi-peer-doc]: /snap/howto/networking/multi-peer-bgp.md
