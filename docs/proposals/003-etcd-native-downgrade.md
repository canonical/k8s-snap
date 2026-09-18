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

Snapd refresh task order: `prepare-snap` → **`pre-refresh` hook (runs from the
*new* revision while the *old* services are still running)** →
`stop-snap-services` → `link-snap` (binary swap) → `post-refresh` →
`start-snap-services` → `configure`.

The `pre-refresh` hook is therefore the **only** snapd hook that runs *before*
the binary swap *while the old etcd is still serving* — exactly the window in
which the Downgrade API can be driven. It already exists and already delegates
to a hidden k8sd command (`k8s x-snapd-config disable`,
[snap/hooks/pre-refresh](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/snap/hooks/pre-refresh)),
so the pattern is established.

Because hooks ship with the *target* revision, this covers channel downgrades
to any revision that contains the hook. The integration-test downgrade path
uses `snap refresh` to channels, i.e. exactly this flow.

### User scenarios

1. **Operator rollback after a failed 1.37 rollout.** `snap refresh
   --channel=1.36-classic/stable k8s` on each control plane node. The
   pre-refresh hook enables the etcd downgrade before snapd swaps binaries;
   etcd 3.6 starts cleanly and the cluster keeps all data written while on
   1.37.
2. **Nightly downgrade CI.** `test_version_downgrades_with_rollback` crosses
   the 1.36/1.37 boundary without special-casing.
3. **Upgrade abort symmetry.** If a downgrade is enabled but the refresh is
   aborted, the next *upgrade* refresh cancels the stale `DowngradeInfo` (see
   Implementation notes), returning the cluster to its normal state.

## User facing changes
<!--
This section MUST describe any user-facing changes that this feature brings, if
any. If an API change is required, the affected endpoints MUST be mentioned. If
the output of any k8s command changes, the difference MUST be mentioned, with a
clear example of "before" and "after".
-->

Behavioral change on snap downgrade across an etcd minor-version boundary:

- **Before**: `snap refresh --channel=1.36-classic/stable` leaves `k8s.etcd`
  crash-looping (`invalid downgrade; server version is lower than determined
  cluster version`); the cluster is down and manual disaster recovery is
  required.
- **After**: the refresh succeeds; etcd starts on the older version with all
  data intact. The `snap change` output shows the pre-refresh hook preparing
  the etcd downgrade.

Documentation note (not a behavior change): `snap revert` runs **no**
pre/post-refresh hooks (snapd skips refresh hooks for reverts), so it cannot
be intercepted. `snap revert` across an etcd minor-version boundary is
documented as unsupported; `snap refresh --channel=<lower track>` is the
supported downgrade path. This is already the path exercised by CI.

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
  documented as unsupported. (A future proposal could add a wrapper-level
  guard that fails fast with recovery instructions.)
- **Restoring the DR snapshot.** The snapshot saved in step 4b exists solely
  to satisfy the upstream disaster-recovery checklist (a destroyed cluster
  mid-downgrade). Restoring it *does* roll data back to snapshot time; it is
  a manual, last-resort operator action and never part of the automatic flow.
- **Downgrades spanning more than one etcd minor version** (e.g. 1.37 →
  1.32). etcd's `allowedDowngradeVersion` only permits exactly one minor step;
  multi-step downgrades must go track by track, each with its own enable step.
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
One new **hidden** command (not part of the public CLI contract; follows the
existing `x-snapd-config`, `x-cleanup` pattern):
- `k8s x-etcd prepare-downgrade` — invoked by the snap `pre-refresh` hook.
  Behavior:
  1. No-op on worker nodes and on nodes where etcd is not configured.
  2. Compare the etcd version of the *currently installed* revision
     (`/snap/k8s/current/bom.json`, `components.etcd.version`) with the *target*
     revision's (`$SNAP/bom.json`); parse both as semver.
  3. If target ≥ current → this is an upgrade or same-version refresh: if a
     stale `DowngradeInfo` is enabled, cancel it (`Downgrade(CANCEL)`);
     exit 0.
  4. If target < current → etcd downgrade, following the upstream checklist
     ([downgrade_3_7](https://etcd.io/docs/v3.7/downgrades/downgrade_3_7/)):
     a. Verify cluster health (`endpoint health`); unhealthy → exit non-zero.
     b. Save a snapshot to `$SNAP_COMMON/var/lib/etcd-backup/` (upstream DR
        checklist item; never restored on the happy path — see Out of scope).
     c. `Downgrade(VALIDATE, <target major.minor>)`.
     d. `Downgrade(ENABLE, <target major.minor>)`. Treat
        `ErrDowngradeInProcess` as success (another node already enabled it).
     e. **Wait until every member's storage version reports the target**
        (`endpoint status`, bounded timeout) — upstream requires this before
        any member is stopped.
     f. If the local member is the raft leader, `move-leader` to another
        member so snapd's service stop does not force an election storm.
  5. On any failure → exit non-zero. A failing `pre-refresh` hook **aborts the
     refresh**, which is the desired safe behavior: the downgrade is blocked
     instead of killing etcd.

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
- Release notes for the tracks receiving the hook (1.36 patch, 1.37, main):
  downgrades across the etcd 3.6/3.7 boundary require the target revision to
  contain this change (minimum revision numbers to be listed at release time).
- Troubleshooting page: what to do if a downgrade was attempted against a
  target revision without the hook (recovery via re-refresh to the newer
  track, then downgrade again to a patched revision).

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
- **Integration (new cases)**:
  - aborted refresh after enable: abort the refresh post-`pre-refresh`, then
    refresh back to the newer track; assert the hook cancels the stale
    `DowngradeInfo` and the cluster returns to normal,
  - HA (3 control plane nodes): concurrent pre-refresh hooks converge
    (`ErrDowngradeInProcess` handled), all members start on 3.6, and
    `endpoint status` shows `DOWNGRADE ENABLED` reset to false once complete,
  - simulated refresh to a revision whose hook detects an unhealthy etcd →
    refresh aborts and the cluster keeps running on the current revision.
- **k8sd unit tests**: version comparison (upgrade/same/downgrade matrix),
  `ErrDowngradeInProcess` handling, cancel-on-upgrade path, failure → non-zero
  exit.

## Considerations for backwards compatibility
<!--
In this section, you MUST mention any breaking changes that are introduced by
this feature. Some examples:

- In case of deleting a database table, how do older k8sd instances handle it?
- In case of a changed API endpoint, how do existing clients handle it?
- etc
-->

- The hook ships with the **target** revision. Downgrades to revisions that
  predate this change remain broken (documented; operators must downgrade to a
  patched revision or newer within the track). The hook therefore needs to be
  **backported to all still-supported release branches** (at minimum
  `release-1.36`, so 1.37 → 1.36 works; consider `release-1.32` so 1.36 → 1.32
  keeps working the same way).
- No API/DB changes → no mixed-version k8sd concerns; old and new k8sd
  binaries interoperate as today.
- Enabling the etcd downgrade sets the cluster version to the older minor and
  the cluster starts speaking the older protocol immediately (3.7-only
  features become unavailable even before any binary is swapped) — this is
  upstream-defined behavior and will be documented.
- **Abort/rollback ordering** follows upstream: to abort an in-progress
  downgrade, members still on the old binary must first be refreshed *back*
  to the newer track (each such refresh's hook sees an upgrade and cancels
  the stale `DowngradeInfo` — step 3 of the command). The cluster version
  only returns to the newer minor once *all* members run the newer binary, so
  members on the old binary never panic during the abort. Once fully
  re-upgraded, the operator can retry the downgrade.
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

### k8s-snap: `snap/hooks/pre-refresh`

[snap/hooks/pre-refresh](https://github.com/canonical/k8s-snap/blob/4c8a5437d887316d82309829eee723d40a18ab22/snap/hooks/pre-refresh)
gains one call after the existing `x-snapd-config disable`:

```bash
# Prepare etcd for a possible downgrade to an older etcd minor version.
k8s::cmd::k8s x-etcd prepare-downgrade
```

The hook must propagate the exit code (already `#!/bin/bash -e`) so a failed
prepare aborts the refresh. Worker nodes self-skip inside the command.

### k8sd: new hidden command `k8s x-etcd prepare-downgrade`

- Command wiring follows the existing hidden commands
  ([cmd/k8s/k8s_x_snapd_config.go](https://github.com/canonical/k8sd/blob/61a67e5771954b6305c2dd816fd01d8a7f0780ed/cmd/k8s/k8s_x_snapd_config.go)).
- Version sources:
  - current/old revision: `/snap/k8s/current/bom.json` (the `current` symlink
    still points at the old revision during `pre-refresh`; it is switched at
    `link-snap`),
  - target revision: `$SNAP/bom.json` (the hook runs from the new revision).
  - bom parsing mirrors `NodeKubernetesVersion`
    ([pkg/snap/snap.go](https://github.com/canonical/k8sd/blob/61a67e5771954b6305c2dd816fd01d8a7f0780ed/pkg/snap/snap.go));
    add a small helper to read `components.etcd.version` from an explicit path.
- etcd connection: reuse `snap.EtcdClient(endpoints)`
  ([pkg/snap/snap.go](https://github.com/canonical/k8sd/blob/61a67e5771954b6305c2dd816fd01d8a7f0780ed/pkg/snap/snap.go));
  endpoint `https://127.0.0.1:2379` (etcd listens on localhost per
  [pkg/k8sd/setup/etcd.go](https://github.com/canonical/k8sd/blob/61a67e5771954b6305c2dd816fd01d8a7f0780ed/pkg/k8sd/setup/etcd.go)),
  certs from `EtcdPKIDir()` (`/etc/kubernetes/pki/etcd`).
- Downgrade calls: `client.Downgrade(ctx, clientv3.DowngradeValidate, target)`
  then `client.Downgrade(ctx, clientv3.DowngradeEnable, target)` with
  `target = fmt.Sprintf("%d.%d", major, minor)` of the *target* etcd version
  (etcd expects the one-minor-below cluster version; see
  `allowedDowngradeVersion` in
  [server/etcdserver/version/downgrade.go](https://github.com/etcd-io/etcd/blob/5e7fd0de973b807a35a18e1adbc2db077f5a3e9c/server/etcdserver/version/downgrade.go)).
- Storage-version wait: poll `endpoint status` (maintenance `Status` RPC)
  until every member reports storage version == target minor, bounded timeout
  (~60 s; upstream notes the migration "usually happens very fast").
- Leader handoff: `endpoint status` to find the leader ID; if it is the local
  member, `MoveLeader` to another non-learner member.
- Map etcd's `rpctypes.ErrDowngradeInProcess` / `ErrInvalidDowngradeTarget`
  to clear messages; `ErrDowngradeInProcess` on enable = success.
- Skip logic: `snaputil.IsWorker(a.snap)` and absence of
  `$SNAP_COMMON/args/etcd` (etcd not configured) → exit 0.

### Sequencing / rollout

1. Land the k8sd hidden command + k8s-snap hook on `main`.
2. Backport to `release-1.37` and `release-1.36` (at minimum); publish new
   revisions to those tracks. Only then do downgrades across the boundary
   become safe.
3. Re-enable/verify the nightly downgrade CI across 1.36 ↔ 1.37.

### Key etcd facts (pinned at v3.7.1, commit `5e7fd0de`)

- Startup gate: `MustDetectDowngrade` panics when binary < persisted cluster
  version and `DowngradeInfo.Enabled` is false
  (`server/etcdserver/version/downgrade.go`,
  `server/etcdserver/api/membership/cluster.go`).
- Enable flow: `Downgrade(ENABLE)` → raft `DowngradeInfoSet` → cluster version
  set to target → `UpdateStorageVersionIfNeeded` migrates storage version
  (`server/etcdserver/version/monitor.go`).
- 3.6↔3.7 schema is identical (`schemaChanges[V3_7] = {}`,
  `server/storage/schema/schema.go`); no 3.7-annotated `InternalRaftRequest`
  fields, so the WAL gate passes.
- When all members reach the target, `CancelDowngradeIfNeeded` clears the
  downgrade flag automatically.
