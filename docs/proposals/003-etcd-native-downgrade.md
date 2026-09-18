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
- **Name**: Safe etcd downgrades via etcd's native downgrade flow

<!-- Owner name and github handle -->
- **Owner**: TBD / <!-- [@name](https://github.com/name) -->

# Proposal Details

## Summary
<!--
In a short paragraph, explain what the proposal is about and what problem
it is attempting to solve.
-->

The 1.37 release bumps the bundled etcd from v3.6.x (1.32/1.36 tracks) to
v3.7.1 ([build-scripts/components/etcd/version](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/build-scripts/components/etcd/version)).
Once a cluster has run on etcd 3.7, the persisted etcd cluster version is 3.7
and an etcd 3.6 binary **refuses to start** against it (etcd panics with
`invalid downgrade; server version is lower than determined cluster version`).
This breaks `snap refresh` channel downgrades (e.g. `1.37-classic/stable` →
`1.36-classic/stable`) and `snap revert`. This proposal drives etcd's **native
downgrade flow** (`DowngradeValidate` + `DowngradeEnable`) from the snap
`pre-refresh` hook — while the old etcd is still running and serving — so the
data dir is migrated in place and the older binary starts cleanly after the
binary swap. **No snapshot/restore, no data loss.**

## Rationale
<!--
This section COULD be as short or as long as needed. In the appropriate amount
of detail, you SHOULD explain how this proposal improves k8s-snap, what is the
problem it is trying to solve and how this makes the user experience better.

You can do this by describing user scenarios, and how this feature helps them.
You can also provide examples of how this feature may be used.
-->

### Why the downgrade breaks

etcd persists a cluster version in its backend. When all members run 3.7, the
version monitor sets the cluster version to 3.7
([server/etcdserver/version/monitor.go](https://github.com/etcd-io/etcd/blob/5e7fd0de973b807a35a18e1adbc2db077f5a3e9c/server/etcdserver/version/monitor.go)).
At startup, every member validates its own binary version against the persisted
cluster version via `MustDetectDowngrade`
([server/etcdserver/version/downgrade.go](https://github.com/etcd-io/etcd/blob/5e7fd0de973b807a35a18e1adbc2db077f5a3e9c/server/etcdserver/version/downgrade.go),
called from
[server/etcdserver/api/membership/cluster.go](https://github.com/etcd-io/etcd/blob/5e7fd0de973b807a35a18e1adbc2db077f5a3e9c/server/etcdserver/api/membership/cluster.go)):

```go
// if the cluster disables downgrade, check local version against determined cluster version.
if cv != nil && sv.LessThan(*cv) {
    lg.Panic("invalid downgrade; server version is lower than determined cluster version", ...)
}
```

So after a snap downgrade to a revision bundling etcd 3.6, etcd **panics on
every startup** — unless the cluster was explicitly prepared for downgrade.
This is the failure observed in `test_version_downgrades_with_rollback`
([tests/integration/tests/test_version_upgrades.py](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/tests/integration/tests/test_version_upgrades.py))
across the 1.36/1.37 boundary.

### The native etcd downgrade flow (no data loss)

This proposal implements the **official upstream downgrade procedure**,
[Downgrade etcd from v3.7 to v3.6](https://etcd.io/docs/v3.7/downgrades/downgrade_3_7/),
automated for the snap. The upstream procedure is:

1. Verify the cluster is healthy (`etcdctl endpoint health`).
2. Take a snapshot backup as a disaster-recovery safety net (not part of the
   downgrade itself).
3. `etcdctl downgrade validate 3.6` — succeeds iff the target is exactly one
   minor below the current cluster version
   ([server/etcdserver/version/version.go](https://github.com/etcd-io/etcd/blob/5e7fd0de973b807a35a18e1adbc2db077f5a3e9c/server/etcdserver/version/version.go)).
4. `etcdctl downgrade enable 3.6` — writes `DowngradeInfo` through raft, sets
   the cluster version to 3.6, and the version monitor migrates the storage
   version back to 3.6. Upstream requires confirming **all members' storage
   version** report 3.6 (`etcdctl endpoint status`) before stopping any
   member. The 3.6↔3.7 schema change list is empty
   ([server/storage/schema/schema.go](https://github.com/etcd-io/etcd/blob/5e7fd0de973b807a35a18e1adbc2db077f5a3e9c/server/storage/schema/schema.go),
   `schemaChanges[V3_7] = {}`), and no 3.7-annotated field exists on
   `InternalRaftRequest` (the only 3.7 proto annotation is on
   `RangeStreamResponse`, an RPC *response* that never enters the WAL), so the
   WAL minimal-version gate
   ([server/storage/wal/version.go](https://github.com/etcd-io/etcd/blob/5e7fd0de973b807a35a18e1adbc2db077f5a3e9c/server/storage/wal/version.go))
   passes in practice.
5. Restart members one at a time on the v3.6 binary with identical
   configuration, leader last (`etcdctl move-leader` away from the leader
   first). With `DowngradeInfo.Enabled` set, `MustDetectDowngrade` passes and
   the member serves normally.
6. When all members run 3.6, `CancelDowngradeIfNeeded` completes the downgrade
   automatically (`DOWNGRADE ENABLED` resets to false).

All data is preserved — this is an in-place, quorum-driven version transition,
not a restore. Mixed-version operation is supported during the transition; the
cluster speaks the protocol of the lowest member version once the downgrade is
enabled.

### Why the `pre-refresh` hook is the right interception point

Snapd refresh task order: `prepare-snap` → `mount-snap` → **`pre-refresh` hook →
`stop-snap-services`** → `link-snap` (binary swap) → `post-refresh` →
`start-snap-services` → `configure`
([overlord/snapstate/snap.go](https://github.com/canonical/snapd/blob/3a626a1619a1c46e25d031b1cf2e0da09d8205f2/overlord/snapstate/snap.go)).

The `pre-refresh` hook is the **only** hook that runs *before* the binary swap
*while the old etcd is still serving* — exactly the window in which the Downgrade
API can be driven. Two verified-but-non-obvious properties of that window shape
the implementation:

1. **The hook executes from the OLD revision.** `SetupPreRefreshHook` sets no
   revision ([canonical/snapd hooks.go](https://github.com/canonical/snapd/blob/master/overlord/hookstate/hooks.go)),
   and snapd's own spread test records `pre-refresh at revision x1` (the source
   revision) on an x1→x2 refresh. This inverts what matters for rollout: the
   mechanism must ship in the **source** track (see Backwards compatibility).
2. **The target revision is already mounted.** `mount-snap` precedes the hook in
   the chain, so the hook (running as the old revision) can read the target's
   `bom.json` from `/snap/k8s/<new-rev>/bom.json` after discovering the revision
   (see Implementation notes).

The hook delegates to a hidden k8sd command, following the established pattern
(`k8s x-snapd-config disable` in
[snap/hooks/pre-refresh](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/snap/hooks/pre-refresh)).
Crucially, the command uses **only local files and the etcd client** — never the
k8sd API socket — because k8sd cannot be relied upon in hook context (precedent:
the `post-refresh` lock-file comment in
[snap/hooks/post-refresh](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/snap/hooks/post-refresh)).

### User scenarios

1. **Operator rollback after a failed 1.37 rollout.** `snap refresh
   --channel=1.36-classic/stable k8s` on each control plane node. The
   pre-refresh hook prepares the etcd downgrade before snapd swaps binaries;
   etcd 3.6 starts cleanly and the cluster keeps all data written while on
   1.37.
2. **Nightly downgrade CI.** `test_version_downgrades_with_rollback` crosses
   the 1.36/1.37 boundary without special-casing.
3. **The aborted upgrade (auto-revert).** An operator upgrades 1.36 → 1.37;
   etcd 3.7 starts, but another service (or the `configure` hook) then fails,
   and snapd's **undo** path reverts the node to 1.36. The undo runs no hooks
   — like `snap revert` — so without mitigation the reverted node strands
   whenever the cluster version had already moved to 3.7 (**always** on
   single-node clusters; on HA, when the aborted node was the last to
   upgrade). Mitigated by the wrapper guard (precise error + recovery) and,
   when the abort is triggered by our own `configure` hook, by the
   **arm-before-fail** step (best-effort `DowngradeEnable` before exiting
   non-zero — see CLI Changes). Recovery in all cases: refresh back to 1.37
   (binary matches cv → starts), then optionally downgrade via the prepared
   path. This is the most likely real-world trigger of the panic.
4. **Canary rollback.** The most common *intentional* downgrade: one node of
   an HA cluster is upgraded to 1.37 and rolled straight back. The cluster
   version never moved (all members must run 3.7 before cv lifts), so no
   preparation is needed — the decision matrix below explicitly no-ops this
   case instead of blocking it.

## User facing changes
<!--
This section MUST describe any user-facing changes that this feature brings, if
any. If an API change is required, the affected endpoints MUST be mentioned. If
the output of any k8s command changes, the difference MUST be mentioned, with a
clear example of "before" and "after".
-->

Behavioral changes:

- **Before**: `snap refresh --channel=1.36-classic/stable` leaves `k8s.etcd`
  crash-looping (`invalid downgrade; server version is lower than determined
  cluster version`); the cluster is down and manual disaster recovery is
  required.
- **After**: the refresh succeeds; etcd starts on the older version with all
  data intact. The `snap change` output shows the pre-refresh hook preparing
  the etcd downgrade.

- **Aborted-upgrade auto-revert (new guard)**: a node that strands after a
  failed 1.37 upgrade gets a precise `k8s.etcd` startup error from the
  **wrapper guard** (`binary < cv` → exit 1 with recovery instructions:
  re-refresh to the newer track, then downgrade via the prepared path)
  instead of an opaque panic loop.

- **Aborted-upgrade mitigation (arm-before-fail)**: when the 1.37 `configure`
  hook itself is what fails the refresh, it arms the etcd downgrade
  (best-effort, bounded timeout) before exiting non-zero, so the undo-revert
  to 1.36 lands on a prepared cluster.

Documentation notes (not behavior changes): `snap revert` runs **no**
pre/post-refresh hooks (verified against snapd source: refresh hooks are
gated on `!Flags.Revert`), so revert across an etcd minor boundary stays
**unsupported**; `snap refresh --channel=<lower track>` is the supported
downgrade path — the path CI exercises.

## Alternative solutions
<!--
This section SHOULD list any possible alternative solutions that have been or
should be considered. If required, add more details about why these alternative
solutions were discarded.
-->

1. **Snapshot before upgrade, restore on downgrade** (`etcdctl snapshot save` /
   `etcdutl snapshot restore`). **Rejected**: a restore rolls the datastore
   back to snapshot time — every write between snapshot and downgrade is lost,
   which is unacceptable for a downgrade path that operators use to recover
   from a bad rollout of a *higher* layer, not to rewind their data.
2. **Patch the bundled etcd to drop/soften the `MustDetectDowngrade` panic.**
   The schema for 3.6 and 3.7 is identical, so the panic is overly conservative
   for this pair. **Rejected**: patching out an upstream safety check is
   fragile across future etcd bumps, masks genuinely unsafe downgrades
   (e.g. 3.7 → 3.5), and diverges from upstream behavior that the native
   Downgrade API already handles correctly.
3. **Detect the downgrade in the etcd service wrapper**
   ([k8s/wrappers/services/etcd](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/k8s/wrappers/services/etcd))
   and remediate before `exec`. **Rejected** as the primary mechanism: at that
   point the old binary is already in place and the cluster version is already
   3.7; enabling the downgrade requires a *running* quorum, which no longer
   exists once members panic — chicken-and-egg. (May still be added as a
   last-resort guard that fails fast with a clear message.)
4. **Do nothing; document that downgrades are unsupported.** **Rejected**:
   channel downgrades and rollbacks are a tested, advertised part of the
   release machinery (nightly downgrade tests, rollback segments).
5. **Keep etcd pinned at 3.6.x in the 1.37 track.** Previously attempted.
   **Rejected**: delays the same cliff to the next etcd bump and blocks etcd
   security/bugfix updates in 1.37.

## Out of scope
<!--
This section MUST reference any work that is out of scope for this proposal.
Out of scope items are typically unknowns that we do not yet have a clear idea
of how to solve, so we explicitly do not tackle them until we have more
information.

This section is very useful to help guide the implementation details section
below, or serve as reference for future proposals.
-->

- **`snap revert` remediation.** snapd runs no pre/post-refresh hooks on
  revert, so there is no interception point before the binary swap. Channel
  downgrade is the supported path; revert across an etcd minor boundary is
  documented as unsupported, with the wrapper guard providing the recovery
  message. (Same hard boundary as the aborted-upgrade undo path, scenario 3.)
- **Offline repair of a stranded member.** The panic gates on the member's
  *local* backend cluster version, which (a) no offline tool can rewrite
  (`etcdutl` touches storage version, not cluster version) and (b) no remote
  enable can reach after the member is stopped. A bbolt-based
  cluster-version rewriting tool could close this (HA: plus a quorum-side
  enable; single-node: self-sufficient) but is new surface for a follow-up.
  **Aborts triggered by a failing *service* at `start-snap-services`** (rather
  than by our `configure` hook) hit the same residual hole — guard + recovery
  only.
- **Restoring the DR snapshot.** The snapshot saved in the preparation step
  exists solely to satisfy the upstream disaster-recovery checklist. Restoring
  it *does* roll data back; it is a manual, last-resort operator action and
  never part of the automatic flow.
- **Downgrades spanning more than one etcd minor version** (e.g. 1.37 →
  1.32). etcd's `allowedDowngradeVersion` only permits exactly one minor step;
  the matrix hard-fails such requests with a clear message.
- **Full refresh ordering across nodes.** snapd refreshes each node
  independently; we cannot serialize "all nodes validate, then all swap".
  The design tolerates this: `enable` is cluster-wide and idempotent, and the
  hook's leader-handoff avoids election churn. Tighter integration with the
  Upgrade CR / FeatureUpgrade machinery can be considered later.
- The legacy k8s-dqlite datastore (upgrades to 1.36+ are already blocked when
  k8s-dqlite data exists, see
  [snap/hooks/post-refresh](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/snap/hooks/post-refresh)).

# Implementation Details

## API Changes
<!--
This section MUST mention any changes to the k8sd API, or any additional API
endpoints (and messages) that are required for this proposal.

Unless there is a particularly strong reason, it is preferable to add new v2/v3
APIs endpoints instead of breaking the existing APIs, such that API clients are
not affected.
-->
none — no k8sd REST API or k8s-snap-api changes. The downgrade is driven via
the etcd client already available in k8sd
([pkg/client/etcd/client.go](https://github.com/canonical/k8sd/blob/61a67e5771954b6305c2dd816fd01d8a7f0780ed/pkg/client/etcd/client.go)
embeds `clientv3.Client`, whose maintenance client exposes
`Downgrade(ctx, action, version)`).

## CLI Changes
<!--
This section MUST mention any changes to the k8s CLI, e.g. new arguments,
different outputs.
-->
Two new **hidden** commands (not part of the public CLI contract; follows the
existing `x-snapd-config`, `x-cleanup` pattern), plus one hook behavior:

1. `k8s x-etcd prepare-downgrade` — invoked by the snap `pre-refresh` hook.
   Inputs: current binary version from the **old** revision's own
   `$SNAP/bom.json`; target revision discovered from the mounted
   `/snap/k8s/<rev>/` tree (see Implementation notes); persisted cluster
   version (`cv`) read from the **local member's** `GET /version` endpoint;
   downgrade info from the maintenance `Status` RPC. Decision matrix
   (major.minor granularity):

   | # | Condition | Meaning | Action |
   |---|-----------|---------|--------|
   | 1 | worker node / no etcd args | not an etcd member | exit 0 |
   | 2 | no target revision found | same-revision refresh | exit 0 |
   | 3 | `B_tgt ≥ B_cur` (upgrade/same) | forward refresh | if downgrade still enabled → `Downgrade(CANCEL)`; `ErrNoInflightDowngrade` = success; **etcd unreachable → warn, exit 0** (never block upgrades) |
   | 4 | `B_tgt < B_cur` ∧ `B_tgt == cv` | **canary rollback** (cluster never moved) | no-op, exit 0 |
   | 5 | `B_tgt < B_cur` ∧ `DowngradeInfo.Enabled(B_tgt)` | already armed (resume after abort) | skip validate/enable; continue at storage wait |
   | 6 | `B_tgt < B_cur` ∧ `B_tgt == cv − 1` | real, supported downgrade | full preparation (below); exit 0 |
   | 7 | `B_tgt < B_cur` ∧ `B_tgt < cv − 1` | multi-minor skip | exit 1: "unsupported: spans more than one etcd minor" |
   | 8 | `B_tgt < B_cur` ∧ etcd unreachable | confirmed downgrade, cannot prepare | exit 1 (abort refresh — staying on current revision is the safe state) |

   Full preparation (row 6), bounded by an internal 8-minute budget (snapd
   hard-kills hooks at 10 minutes):
   a. health check (all members, no active alarms)
   b. free-space check ≥ ~1.2× local DB size
   c. DR snapshot to `$SNAP_COMMON/var/lib/etcd-backup/` (atomic: `.part`
      file + rename; rotate keep-3)
   d. `Downgrade(VALIDATE, B_tgt)` — `ErrInvalidDowngradeTarget` → exit 1
   e. `Downgrade(ENABLE, B_tgt)` — `ErrDowngradeInProcess` = success
   f. poll every member's `Status` until storage version == `B_tgt` (also
      transitively confirms cv moved)
   g. if the local member is raft leader and another voter exists →
      `MoveLeader` (skipped on single-node)

   On any failure → exit non-zero → snapd **aborts the refresh** before any
   service stop (the safe state). Audit: append one JSON line per run
   (versions, cv, matrix row, result) to the backup dir — the integration
   test asserts on it so the mechanism can never silently no-op.

2. `k8s x-etcd prestart-check` — invoked by the etcd service wrapper before
   `exec` (new revision runs it on every start, including after `snap
   revert`:
   - read own binary version from `$SNAP/bom.json`; query peer endpoints
      (from `args/etcd`) for cv;
   - reachable peers ∧ `binary < cv` → print exact cause + recovery
     (re-refresh to newer track, then downgrade via a hooked revision) →
     exit 1;
   - unreachable or `binary ≥ cv` → **fail open**: `exec` etcd. The guard
     must never block a legitimate start.

3. `snap/hooks/configure` behavior change (arm-before-fail): when the
   `configure` hook of a revision is about to **fail the refresh** (after
   this change, new revisions only), it first runs
   `k8s x-etcd prepare-downgrade` logic in downgrade direction as a
   best-effort parting step (bounded timeout, ≈30 s), so the snapd undo-back
   to the older revision lands on a prepared cluster. Aborts triggered by a
   failing *service* at `start-snap-services` are out of reach (configure
   never runs) — covered by the guard (mitigation 2).

## Database Changes
<!--
This section MUST mention any changes required in the k8sd database schema or
internal types.
-->
none.

## Configuration Changes
<!--
This section MUST mention any new configuration options or service arguments
that are introduced.
-->
No new configuration options or service arguments. One new on-disk artifact:

| Path | Content |
|------|---------|
| `$SNAP_COMMON/var/lib/etcd-backup/snapshot-<ts>.db` | DR snapshot taken at downgrade-prepare time (rotated, keep latest 3) |

No args changes are needed for the etcd service itself: per the upstream
checklist, v3.7 introduces no new flags relative to v3.6, so the existing
`$SNAP_COMMON/args/etcd` is valid for both binaries.

## Documentation Changes
<!--
This section MUST mention any new documentation that is required for the new
feature. Most features are expected to come with at least a How-To and an
Explanation page.

In this section, it is useful to think about any existing pages that need to
be updated (e.g. command outputs).
-->

- Update the upgrade/downgrade documentation: supported downgrade path is
  `snap refresh --channel=<lower track>`; `snap revert` across etcd minor
  versions is unsupported.
- Troubleshooting page covering both strand scenarios: (a) intentional
  `snap revert`, (b) **aborted-upgrade auto-revert** — recovery is always
  "refresh back to the newer track, then downgrade via the prepared path".
- Release notes for the tracks receiving the mechanism: downgrades across the
  etcd 3.6/3.7 boundary require the **currently installed (source) revision**
  to contain the hook (minimum revision numbers listed at release time).

## Testing
<!--
This section MUST explain how the new feature will be tested.
-->

- **Integration (existing suite)**: `test_version_downgrades_with_rollback`
  ([tests/integration/tests/test_version_upgrades.py](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/tests/integration/tests/test_version_upgrades.py))
  must pass across the 1.36 ↔ 1.37 boundary, including the rollback segments;
  assert after each segment that `k8s.etcd` is active and data written *after*
  the upgrade (i.e. while on 1.37) is still present after the downgrade —
  this is the no-data-loss assertion.
- **Anti-silent-no-op**: after every boundary crossing, assert the audit log
  contains a full-preparation entry for the expected target — a refresh that
  "worked" without the mechanism firing is a test failure.
- **Integration (new cases)**:
  - canary rollback: upgrade 1 of 3 CP nodes to 1.37, roll it straight back
    → succeeds, `endpoint status` shows no `DOWNGRADE ENABLED`;
  - abort-and-resume: abort a downgrade after the hook arms it, refresh back
    to 1.37 (assert the cancel path clears `DowngradeInfo`), downgrade again
    (assert the resume path);
  - **aborted upgrade (auto-revert)**: on a single node, start a 1.36 → 1.37
    refresh, inject a failure *in the `configure` hook* → snapd undo-reverts
    → assert etcd 3.6 starts (arm-before-fail) and `endpoint status` returns
    to normal; on a second run, inject the failure via a service at
    `start-snap-services` → assert the wrapper guard's recovery message in
    the `k8s.etcd` journal and that refreshing to 1.37 revives the node;
  - HA concurrent pre-refresh hooks converge (`ErrDowngradeInProcess`
    handled), all members start on 3.6, `DOWNGRADE ENABLED` resets to false;
  - single-node full downgrade 1.37 → 1.36;
  - etcd-down upgrade: stop `k8s.etcd` on one node, refresh 1.36 → 1.37 →
    refresh succeeds (warn-and-continue must not block upgrades);
  - simulated unhealthy etcd on a downgrade path → refresh aborts, cluster
    keeps running on the current revision.
- **k8sd unit tests**: target-revision discovery (0/1/N mounted candidates,
  collapse-on-equal-version, ambiguous → fail-safe), the decision matrix
  rows 1–8 against a fake etcd client (`ErrDowngradeInProcess`/`ErrNoInflightDowngrade`/
  `ErrInvalidDowngradeTarget` mappings, etcd-down upgrade vs downgrade
  asymmetry), single-node move-leader skip, arm-before-fail timeout behavior,
  wrapper guard fail-open semantics.

## Considerations for backwards compatibility
<!--
In this section, you MUST mention any breaking changes that are introduced by
this feature. Some examples:

- In case of deleting a database table, how do older k8sd instances handle it?
- In case of a changed API endpoint, how do existing clients handle it?
- etc
-->

- **The mechanism ships in the SOURCE revision.** The pre-refresh hook that
  runs (and the k8sd binary behind `k8s x-etcd …`) belongs to the revision
  installed *before* the refresh (verified: snapd's spread test records
  `pre-refresh at revision x1` on an x1→x2 refresh). Rollout therefore:
  1. land on `main`; **backport to `release-1.37` — required** (only hooked
     1.37 revisions can downgrade themselves to 1.36);
  2. backport to `release-1.36` — recommended (cancel path on rollback
     segments, plus future boundaries where 1.36 is the source);
  3. no backport to ≤1.32: no etcd minor boundary below 1.36 (all ship
     3.6.x) — the command no-ops there.
  Release notes must state minimum **source** revisions: downgrades from
  revisions predating the change remain broken (documented; recovery via
  re-upgrade then hooked downgrade).
- **Aborted-upgrade exposure exists from day one of 1.37** and shrinks as
  hooked revisions land: single-node clusters strand on *every* aborted
  1.36→1.37 upgrade until then (cv moves to 3.7 seconds after etcd 3.7
  starts); HA clusters only when the aborted node was the last to upgrade.
  The wrapper guard + recovery ships in the same change; arm-before-fail
  covers hook-triggered aborts.
- No API/DB changes → no mixed-version k8sd concerns; old and new k8sd
  binaries interoperate as today.
- Enabling the etcd downgrade sets the cluster version to the older minor and
  the cluster starts speaking the older protocol immediately (3.7-only
  features become unavailable even before any binary is swapped) — this is
  upstream-defined behavior and will be documented.
- **Abort/rollback ordering** follows upstream: to abort an in-progress
  downgrade, members are refreshed *back* to the newer track (the command's
  upgrade row cancels the armed `DowngradeInfo`; `ErrNoInflightDowngrade`
  mapped to success). The cluster version only returns to the newer minor
  once *all* members run the newer binary, so members on the old binary never
  panic during the abort.
- **Interruption at any step is covered by construction**: exactly one
  mutation point (the idempotent `enable` raft write) with read-only steps
  before it and wait/cleanup steps after it; post-enable states are valid for
  *both* binary versions. Retries converge via the resume row (downgrade) or
  the cancel row (upgrade). Timeout risk bounded by the internal 8-minute
  budget (snapd hard hook limit: 10 minutes).
- FIPS builds: the change only adds etcd client calls over the existing
  TLS-configured client; no new crypto surface.

## Implementation notes and guidelines
<!--
In this section, you SHOULD go into detail about how the proposal can be
implemented. If needed, link to specific parts of the code (link against
particular commits, not branches, such that any links remain valid going
forward).

This is useful as it allows the proposal owner to not be the person that
implements it.
-->

### k8s-snap: `snap/hooks/pre-refresh` and `k8s/wrappers/services/etcd`

[snap/hooks/pre-refresh](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/snap/hooks/pre-refresh)
gains one call after the existing `x-snapd-config disable`:

```bash
# Prepare etcd for a possible downgrade to an older etcd minor version.
k8s::cmd::k8s x-etcd prepare-downgrade
```

[k8s/wrappers/services/etcd](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/k8s/wrappers/services/etcd)
gains the guard between sourcing `lib.sh` and `k8s::common::execute etcd`:

```bash
# Fail fast with recovery instructions on unprepared etcd downgrades.
k8s::cmd::k8s x-etcd prestart-check
```

Both hooks must propagate the exit code (`#!/bin/bash -e` already does).

[snap/hooks/configure](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/snap/hooks/configure)
gains the arm-before-fail step: on any failure path that would exit non-zero,
run the downgrade-preparation logic best-effort with a ~30 s bound before
exiting (see CLI Changes item 3).

### k8sd: new hidden commands `k8s x-etcd prepare-downgrade` / `prestart-check`

- Command wiring follows the existing hidden commands
  ([cmd/k8s/k8s_x_snapd_config.go](https://github.com/canonical/k8sd/blob/61a67e5771954b6305c2dd816fd01d8a7f0780ed/cmd/k8s/k8s_x_snapd_config.go));
  the root command has no `PersistentPreRun` requiring the k8sd socket, and
  these commands deliberately use **local files + the etcd client only**
  (`snapctl`-style local reads are precedented by `x-snapd-config`).
- **Target-revision discovery** (layered, fail-safe):
  1. list `/snap/k8s/` entries matching `^x?[0-9]+$`, minus `$SNAP_REVISION`;
  2. zero → row 2 of the matrix; one → target;
  3. several → read all candidates' `bom.json`; if they share one etcd
     major.minor → use it (collapses nearly all `refresh.retain=2` cases);
  4. still ambiguous → parse the in-progress refresh's `link-snap` task
     summary from `snap change <id>` (classic confinement only);
  5. still ambiguous → exit 1 with remediation (e.g. `refresh.retain=2`).
- Version sources: current binary version from `$SNAP/bom.json` (the hook
  executes the **old** revision's binary; `$SNAP` is old-scoped), target from
  the discovered revision dir. bom parsing mirrors `NodeKubernetesVersion`
  ([pkg/snap/snap.go](https://github.com/canonical/k8sd/blob/61a67e5771954b6305c2dd816fd01d8a7f0780ed/pkg/snap/snap.go));
  add a helper `ComponentVersionFromBOM(path, component)` (strip `v` prefix).
- **cv must come from the LOCAL member** (`GET https://<localhost>:2379/version`,
  TLS via `EtcdPKIDir()` = `/etc/kubernetes/pki/etcd`): the startup panic
  gates on the local backend — a lagging member may disagree with the cluster
  during transitions, and local reads are what `Recover()` later consults.
- etcd connection: `snap.EtcdClient(endpoints)`
  ([pkg/snap/snap.go](https://github.com/canonical/k8sd/blob/61a67e5771954b6305c2dd816fd01d8a7f0780ed/pkg/snap/snap.go));
  endpoints **parsed from `$SNAP_COMMON/args/etcd`** (`--listen-client-urls`)
  — never hardcode 2379, operators can override ports. clientv3 v3.6.7
  exposes `Downgrade`, `Status` (carries `storageVersion` and
  `downgradeInfo`), `Snapshot`, `MemberList`, `MoveLeader`, `AlarmList`.
- Error mapping: `ErrDowngradeInProcess` / `ErrNoInflightDowngrade` =
  success; `ErrInvalidDowngradeTargetVersion` = hard failure with message.

### Key etcd facts (pinned at v3.7.1, commit `5e7fd0de`)

- Startup gate: `MustDetectDowngrade` panics when binary < cluster version,
  invoked on the member's **local** backend during `Recover()` — *before*
  raft catch-up (`server/etcdserver/api/membership/cluster.go`,
  `server/etcdserver/version/downgrade.go`). It does **not** consult
  `DowngradeInfo` directly — the enable works by moving cv down first; the
  proposal's storage-version wait accidentally but correctly covers this
  (sv only migrates after cv moves).
- Enable flow: `Downgrade(ENABLE)` → raft `DowngradeInfoSet` → cv set to
  target → `UpdateStorageVersionIfNeeded` migrates sv **on every running
  member independently** (no leader check in the monitor,
  `server/etcdserver/server.go`, `version/monitor.go`).
- A member that is **down when cv commits misses it locally forever** — the
  stranded-member trap; recovery is always re-upgrade (binary matches cv).
- 3.6↔3.7 schema is identical (`schemaChanges[V3_7] = {}`); no 3.7-annotated
  `InternalRaftRequest` fields, so the WAL gate passes.
- When all members reach the target, `CancelDowngradeIfNeeded` clears the
  downgrade flag automatically (~5 s cadence, leader-driven).
- Concurrent enable from multiple nodes is idempotent (validate→propose with
  identical `DowngradeInfo`); `ErrDowngradeInProcess` maps to success.

### Sequencing / rollout

1. Land k8sd commands + k8s-snap hook/guard + configure arm-before-fail on
   `main`.
2. Backport to `release-1.37` (**required**) and `release-1.36`
   (**recommended**); publish; release notes list minimum **source**
   revisions.
3. Re-enable/verify the nightly downgrade CI across 1.36 ↔ 1.37, including
   the anti-silent-no-op audit-log assertion.
