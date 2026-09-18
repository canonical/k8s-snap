# etcd downgrade safety — problem, solution, gaps (visual companion)

Companion to [003-etcd-native-downgrade.md](003-etcd-native-downgrade.md), for team
discussion. Diagrams are mermaid (render on GitHub).

**TL;DR** — 1.37 ships etcd 3.7; a node downgraded to 1.36 gets an etcd 3.6 binary
that **panics at startup** against the 3.7 cluster version, and no offline tool can
fix it. We prepare the downgrade **before** snapd swaps binaries, using etcd's own
official downgrade flow, driven from the snap `pre-refresh` hook. No data loss.
Gaps: `snap revert` stays unsupported, simultaneous all-node refreshes can strand a
member (recoverable), and the fix must ship in the **source** track (1.37).

---

## 1. The problem

### 1.1 What happens today on a downgrade

```mermaid
sequenceDiagram
    autonumber
    participant Op as Operator
    participant S as snapd
    participant E as etcd on this node

    Op->>S: snap refresh --channel=1.36-classic/stable
    Note over S: pre-refresh hook has nothing to do<br/>(today's revisions ship no preparation)
    S->>E: stop-snap-services (etcd 3.7 stops)
    S->>S: link-snap — binary swapped to 3.6
    S->>E: start-snap-services (etcd 3.6 starts)
    E->>E: Recover(): local backend says cluster version = 3.7
    E--xE: PANIC — binary 3.6 below cluster version 3.7 — invalid downgrade
    loop crash loop
        S->>E: restart
        E--xE: PANIC again
    end
    Note over E: node down. cv lives in the local bbolt and<br/>neither etcdutl nor the remote cluster can rewrite it
```

### 1.2 The startup gate (why it panics)

Every etcd start reads the **local backend's** persisted cluster version and compares
it to the binary version — before raft catch-up, so a stale local value is fatal even
if the rest of the cluster has moved on:

```mermaid
flowchart TD
    A["etcd binary starts"] --> B["Recover(): read cv from LOCAL bbolt"]
    B --> C{"binary version < cv ?"}
    C -- "no" --> D["start normally"]
    C -- "yes" --> E["PANIC: invalid downgrade<br/>(MustDetectDowngrade)"]
    style E fill:#f66,stroke:#900
    style D fill:#9f9,stroke:#090
```

Key facts:

- cv moves to 3.7 once **all** members run 3.7 (raft consensus).
- cv is applied to each member's **local** backend only while that member is **running**.
- `etcdutl` can rewrite the *storage* version but has **no** cluster-version writer.
- The 3.6↔3.7 storage schema is identical — cv is the only gate that matters here.

---

## 2. The solution: prepare before the swap

snapd's refresh chain gives us exactly one window where the old etcd is still serving
and the refresh can still be aborted:

```mermaid
flowchart LR
    A["prepare-snap"] --> B["mount-snap<br/>(new revision mounted)"]
    B --> C["pre-refresh hook<br/>OLD revision, services UP"]
    C --> D["stop-snap-services"]
    D --> E["link-snap (binary swap)"]
    E --> F["start-snap-services"]
    F --> G["configure"]
    style C fill:#ff9,stroke:#960
```

We hook that window. `pre-refresh` calls a new hidden k8sd command,
`k8s x-etcd prepare-downgrade`, which automates etcd's
[official downgrade procedure](https://etcd.io/docs/v3.7/downgrades/downgrade_3_7/):

```mermaid
sequenceDiagram
    autonumber
    participant S as snapd
    participant H as pre-refresh hook
    participant K as k8s x-etcd prepare-downgrade
    participant C as etcd cluster

    S->>H: run hook (old etcd still serving)
    H->>K: prepare-downgrade
    K->>K: versions: current 3.7 → target 3.6<br/>local cv = 3.7 → real downgrade
    K->>C: health check (all members)
    K->>K: free-space check + DR snapshot (safety net only)
    K->>C: downgrade validate 3.6
    K->>C: downgrade enable 3.6
    C-->>C: raft commits cluster version 3.6<br/>applied to EVERY running member's local backend<br/>storage version migrates 3.7 → 3.6
    K->>C: wait until ALL members report storage 3.6
    K->>C: move-leader away if local member is leader
    K-->>H: exit 0 (any failure → exit 1 → refresh ABORTED, nothing changes)
    H-->>S: hook success
    S->>C: stop-snap-services
    S->>S: link-snap — binary swapped to 3.6
    S->>C: start-snap-services
    C->>C: Recover(): binary 3.6 == local cv 3.6
    Note over C: starts cleanly — all data preserved
```

Because the enable commits **while this node's etcd is still a running raft member**,
its local backend carries cv=3.6 into the swap — the panic gate passes. When the last
node in the cluster lands on 3.6, etcd auto-completes the downgrade.

### 2.1 Decision logic (handles the tricky cases)

```mermaid
flowchart TD
    Start["prepare-downgrade"] --> W{"worker node or<br/>etcd not configured?"}
    W -- "yes" --> OK["exit 0"]
    W -- "no" --> D{"target etcd version<br/>vs current"}
    D -- "target ≥ current<br/>(upgrade / same)" --> U{"stale downgrade<br/>still enabled?"}
    U -- "yes" --> U1["cancel it<br/>(ErrNoInflightDowngrade = ok)"] --> OK
    U -- "no" --> OK
    U -. "etcd unreachable" .-> U2["warn, exit 0<br/>NEVER block upgrades"]
    D -- "target < current" --> CV{"target vs LOCAL cv"}
    CV -- "target == cv" --> CR["canary rollback:<br/>cluster never moved → no-op"] --> OK
    CV -- "target == cv − 1" --> P["FULL PREPARATION<br/>health → snapshot → validate<br/>→ enable → wait storage → move-leader"]
    CV -- "target < cv − 1" --> BAD["exit 1:<br/>multi-minor unsupported"]
    CV -. "etcd unreachable" .-> BAD2["exit 1:<br/>abort refresh (safe state)"]
    P -- "success" --> OK
    P -- "failure" --> BAD2
    style P fill:#ff9,stroke:#960
    style CR fill:#9cf,stroke:#039
    style BAD fill:#f66,stroke:#900
    style BAD2 fill:#f66,stroke:#900
```

Two cases worth calling out:

- **Canary rollback** (upgrade 1 of 3 nodes, roll it back): cv never left 3.6, so no
  preparation is needed at all — the no-op path. An earlier draft of this design
  *blocked* this scenario; the matrix fixes it.
- **cv is read from the LOCAL member's `/version`**, not from any peer: the panic
  keys off the local backend, and a lagging member can disagree with the cluster
  during transitions.

### 2.2 HA rolling downgrade (the tested path)

```mermaid
sequenceDiagram
    participant A as Node A
    participant B as Node B
    participant C as Node C

    Note over A,C: all on 1.37 (etcd 3.7), cv = 3.7
    A->>A: hook: enable 3.6 (raft) → cv = 3.6 everywhere<br/>wait sv, handoff, swap, start 3.6 ✓
    B->>B: hook: enable already done (idempotent)<br/>sv already 3.6 → swap, start 3.6 ✓
    C->>C: same ✓
    Note over A,C: last member on 3.6 → etcd auto-completes<br/>DOWNGRADE ENABLED → false
```

Concurrent hooks converge: the enable is an idempotent raft write, and a second node
simply observes "already in process" and resumes at the wait step.

---

## 3. The gaps (to discuss)

### 3.1 `snap revert` cannot be fixed

```mermaid
flowchart LR
    R["snap revert"] --> NH["snapd runs NO<br/>pre/post-refresh hooks"]
    NH --> P["3.6 binary vs cv 3.7<br/>→ panic loop"]
    P --> G["wrapper guard:<br/>fail fast, print recovery"]
    G --> REC["recovery: snap refresh back to 1.37<br/>(binary matches cv → starts)<br/>then downgrade via the hooked path"]
    style P fill:#f66,stroke:#900
    style REC fill:#9f9,stroke:#090
```

`snap revert` across an etcd minor boundary is **documented as unsupported**;
`snap refresh --channel=<lower track>` is the supported downgrade path (and what CI
exercises). The wrapper guard turns an opaque panic loop into an actionable message.

### 3.2 The strandable member (simultaneous refreshes)

```mermaid
sequenceDiagram
    participant A as Node A
    participant B as Node B

    B--xB: B's etcd is down<br/>(maintenance / crash)
    A->>A: hook: enable 3.6 commits<br/>(B's local backend misses it: cv stays 3.7)
    Note over B: B later boots the 1.37 binary →<br/>raft catch-up fixes cv → fine ✓
    Note over B: B refreshed while down →<br/>hook: etcd unreachable → ABORT → protected ✓
```

Protection comes from **ordering + local state**, not new coordination state:

- a node's hook always runs *before* its own etcd stops, so its local backend always
  has cv moved before the swap;
- a node whose etcd is unreachable at hook time gets its refresh **aborted** (safe);
- worst case, the node is always revivable: re-refresh to the newer track (binary
  matches cv → starts → catch up → retry downgrade).

Residual risk: pathological simultaneous-refresh interleavings on unhealthy clusters
can still strand a member → rolling refreshes are the documented procedure.

### 3.3 The aborted upgrade (auto-revert)

The most likely *real-world* way to hit the panic — not an intentional downgrade,
but snapd undoing a **failed upgrade** to 1.37:

```mermaid
sequenceDiagram
    autonumber
    participant S as snapd
    participant H as hooks
    participant E as etcd on this node

    Note over E: node on 1.36 (etcd 3.6), cv = 3.6
    S->>H: pre-refresh hook (upgrade path: no-op)
    S->>E: stop 3.6 → link 1.37 → start 3.7 ✓
    Note over E: cv moves to 3.7 within seconds<br/>single-node: always · HA: when last node lands
    S--xS: another service fails / configure hook fails
    Note over S: UNDO path: no hooks run → no chance to prepare
    S->>E: stop 3.7 → relink 1.36 → start 3.6
    E--xE: PANIC — binary 3.6 below cv 3.7
```

Whether it bites depends on **cv timing** when the abort hits:

| Situation at abort time | cv | Reverted 3.6 binary |
|---|---|---|
| single-node cluster | 3.7 (moves seconds after 3.7 starts) | **panic — stranded** |
| HA, first/middle node upgraded | 3.6 (peers still on 3.6) | starts fine — lucky |
| HA, last node upgraded | 3.7 (its start completed the set) | **panic — stranded** |

Mitigation — **arm-before-fail**: when the abort is triggered by *our own*
`configure` hook (e.g. a k8sd health gate), that hook arms the downgrade before
exiting non-zero:

```mermaid
flowchart TD
    F["configure hook on 1.37<br/>about to fail the refresh"] --> ARM["best-effort:<br/>downgrade enable 3.6<br/>before exiting non-zero"]
    ARM --> R["snapd undo: revert to 1.36"]
    R --> OK["etcd 3.6 starts:<br/>binary == cv == 3.6 ✓"]
    style ARM fill:#ff9,stroke:#960
    style OK fill:#9f9,stroke:#090
```

Honest limit: if the trigger is **another service failing at start-snap-services**,
the chain unwinds before `configure` ever runs — arm-before-fail gets no chance.
That sub-case falls back to the wrapper guard + always-working recovery:

```mermaid
flowchart LR
    P["panic on reverted node"] --> G["wrapper guard:<br/>fail fast, print recovery"]
    G --> REC["snap refresh --channel=1.37<br/>binary matches cv → starts ✓<br/>fix root cause → optionally downgrade<br/>via the prepared path"]
    style REC fill:#9f9,stroke:#090
```

### 3.4 Gap & mitigation summary

| Gap | Severity | Mitigation |
|-----|----------|------------|
| `snap revert` across boundary | unsupported path | wrapper guard + documented recovery (refresh-based) |
| **Aborted 1.37 upgrade → auto-revert** | **likely real-world trigger** | arm-before-fail (hook-triggered aborts); wrapper guard + re-upgrade recovery (service-start aborts) |
| Simultaneous all-node refresh | residual | docs: roll nodes; recovery always possible via re-upgrade |
| Fix must ship in the **source** track | rollout constraint | backport to `release-1.37` **required**; release notes list min source revision |
| Downgrades to pre-fix revisions | permanent | documented minimum revisions |
| One etcd minor per step | upstream limit | multi-track downgrades go track by track |
| 3.7 features off between enable and last swap | upstream semantics | documented; window is minutes |
| Unhealthy cluster blocks downgrade | by design | hook aborts refresh; cluster stays on current revision |
| DR snapshot disk usage (≤3× DB) | ops | free-space pre-check; rotation |
| 10-min snapd hook timeout | constraint | 8-min internal budget; bounded waits |
| Strict confinement readability (PKI, sibling revs) | open | PoC verification item before strict ships |
| `snap change` parsing (fallback discovery) | fragility | fail-safe abort, never wrong action |

### 3.5 Explicitly not building (for now)

- **Cross-node refresh coordination** (e.g. `gate-auto-refresh` + `snapctl refresh
  --hold`, or Upgrade CR sequencing) — possible future enhancement; only gates
  *auto*-refreshes anyway.
- **A parallel state store** for the downgrade — etcd's own raft-replicated
  `DowngradeInfo` + cluster version is the single source of truth; we only add a
  per-node audit log for observability/tests.
- **Offline repair of a stranded member** — impossible without data loss (no tool
  can rewrite local cv); re-upgrade is the repair.

---

## 4. Rollout

```mermaid
flowchart TD
    M["land on main:<br/>k8sd command + hook + wrapper guard"] --> B37["backport to release-1.37<br/>REQUIRED — only hooked 1.37 revisions<br/>can downgrade themselves"]
    M --> B36["backport to release-1.36<br/>recommended — cancel path + future boundaries"]
    B37 --> RN["release notes: minimum SOURCE revision<br/>for 1.37 → 1.36 downgrades"]
    B36 --> CI["enable nightly downgrade CI<br/>across 1.36 ↔ 1.37"]
    style B37 fill:#ff9,stroke:#960
```

## 5. Questions for the team

1. Are we comfortable documenting `snap revert` as unsupported across etcd minor
   boundaries (channel downgrade as the only supported path)?
2. Rolling-only refreshes as the documented HA procedure — acceptable, or do we want
   the `gate-auto-refresh` hold mechanism investigated now?
3. DR snapshot policy: hard requirement (abort downgrade if it can't be taken) or
   best-effort with a loud warning?
4. Strict flavor: who can run the confinement PoC (PKI + sibling-revision readability
   from hook context)?
5. Backport scope: 1.37 only, or 1.36 as well in the same round?
