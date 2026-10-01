# Delivery plan — CI failure classification, fingerprinting, and issue creation

- **Status**: DRAFTING
- **Owner**: Louise Schmidtgen / [@louiseschmidtgen](https://github.com/louiseschmidtgen)
- **Companion to**: `003-ci-failure-classification-and-metrics.md` (the design
  proposal, which lands as the first PR of movement 1 below)
- **Builds on**: [#2850](https://github.com/canonical/k8s-snap/pull/2850) —
  standardised failure context and `result.json` in every e2e job

This is the delivery plan, not the design. Proposal 003 states *what* we classify
and why the taxonomy looks the way it does; this document states *how* the work
gets landed, in what order, and what has to be true before each step is enabled.

## Problem

CI in `canonical/k8s-snap` is red as a steady state and we cannot say *why*, *who
owns it*, or *whether it is new*. Part one (PR #2850, merged as `84072ee2`) fixed
evidence capture: every e2e job now emits a standardised context table plus
`result.json` carrying test, channel, revision, artifact, os, arch, substrate,
flavor, runner, attempt, sha, source, run_url, status.

Part two exists as a large proof of concept on
`agents/ci-failure-classification-metrics-design`, pinned for review at
[`53e6b6a`](https://github.com/canonical/k8s-snap/tree/53e6b6a7ef73947b5d6aa085905e63c8f36f4a19)
(~8,400 lines: proposal 003, `ci/metrics/*`, `ci/cmds/metrics.py`,
`ci/failure_rules.yaml`, `.github/workflows/ci-metrics.yaml`,
`ci/tests/metrics/*`) with data produced onto the `ci-metrics` branch. It
classifies failures into fault domains, fingerprints them into 16-hex
signatures, and rolls them up. It is unreviewed, unmerged, and stores everything
as hand-rolled JSON. Every line reference below is against that commit, not the
branch tip, which may move.

Part three — the actual gap — is that none of this reaches a human or the
autonomous triage bot (`.github/workflows/triage.yaml`,
`k8s-ai-agent-toolkit`). Nothing opens an issue, so nothing gets fixed.

## Approach

Three movements, in order, each independently valuable:

1. **Land part two in reviewable slices.** Split the PoC branch into a stack of
   small PRs, starting with the design proposal so the taxonomy is agreed before
   the code arrives.
2. **Replace the JSON store with SQLite** as the canonical store, committed to the
   `ci-metrics` branch as a deterministic *text SQL dump* (diffable, reviewable)
   plus generated markdown/JSON exports. The `.db` is a build product, never the
   thing under review.
3. **Build the signature → GitHub issue bridge.** One issue per recurring
   `product.bug` / `test.bug` signature, deduplicated by fingerprint, enriched with
   everything the triage skill needs, labelled for the bot, and left for a
   maintainer to escalate with `/triage`.

### Decisions taken

| Decision | Choice |
|---|---|
| Store | SQLite canonical; committed as text SQL dump + generated markdown/JSON |
| Who gets an issue | `product.bug` and `test.bug` signatures only, above a recurrence threshold, opened automatically |
| Triage depth on open | Label + enriched body; the expensive self-hosted pipeline stays maintainer-gated behind `/triage` |
| LLM | Stage C stays deferred and advisory: proposes rules via PR, never writes a recorded class |

### Taxonomy (axis A — fault domain, who owns it)

`infra.runner`, `infra.provisioning`, `external.dependency`, `ci.config`,
`product.bug`, `test.bug`, `unknown`. Axis B (reproducibility): `systemic`,
`config-specific`, `flaky`, `new` — derived from per-attempt job lists, which are
ground truth rather than inference.

## Movement 1 — land part two

Each item below is a PR against `main`, cherry-picked/rebased out of the PoC
branch, kept small enough to actually review.

1. **Proposal 003** — `docs/proposals/003-ci-failure-classification-and-metrics.md`,
   updated for the SQLite decision and with a new "Part 3: issue creation" section.
   Merge this first: it is the taxonomy contract everything else encodes.
2. **Scaffolding, models, GitHub client, ingest** — `ci/metrics/{models,gh,ingest}.py`,
   `ci/cmds/metrics.py` skeleton, JSON Schemas, tests. Read-only against the Actions
   API, parallel pagination, ingests every *attempt*, refuses to freeze in-flight runs.
3. **Scrubber + fingerprinting** — `ci/metrics/{scrub,signature}.py`,
   `metrics audit-secrets`, fixture tests. Hard gate: audit runs before anything is
   stored or uploaded. Join-token excision in place, not whole-line kills.
4. **Stage A metadata router** — `ci/metrics/router.py`. Deterministic, no log
   fetch; catches the runner-lost case (`failure` conclusion, no failed step, 404
   log) that a log-only classifier structurally cannot see. Router verdicts are
   never overwritten downstream.
5. **Stage B rule pack** — `ci/failure_rules.yaml` + `ci/metrics/rules.py`. Ordered,
   first-match-wins, each rule carrying a `notes` field explaining its class.
   Matches the wrapped command (`k8s bootstrap`, `k8s join-cluster`, …) because the
   harness collapses nearly everything to one `CalledProcessError` frame.
6. **Rollup + report** — `ci/metrics/{rollup,report}.py`. Rates publish their
   denominator or render `n/a`; cross-`ruleset_version` comparison refuses; integrity
   metrics (M9–M12) included so "we ran less" cannot masquerade as "we broke less".
7. **Workflow + digest** — `.github/workflows/ci-metrics.yaml`, `ci-metrics` branch
   README, Mattermost digest reshaped to lead with attribution and delta. `workflow_run`
   for latency, hourly reconciliation for completeness, one daily digest cron.

**Validation gate** (from the proposal, run on live data before the workflow is
enabled): ingest fidelity, signature collapse, runner-lost classification, weekly
attribution, ≥90% agreement on 20 manually audited failures, digest rendering,
**zero secrets** in all stored excerpts, trend report, committed volume.

## Movement 2 — SQLite store

- Canonical store: `ci/metrics-data/metrics.db`, schema-versioned. Every *fact* is
  durable; what expires is bulk raw evidence — full log excerpts and the
  per-step detail table. The step facts that decide a classification
  (`failed_step_name`, `failed_step_number`, `log_available`, and the
  runner-lost tell) are promoted onto `jobs` and are durable, so the
  interpretations that depend on them survive the artifact window. See the
  schema appendix.
- **Committed artefact is `metrics.sql`** — `sqlite3 .dump` of the durable tables
  with stable ordering — so review, `git diff`, and blame keep working and the
  repo does not accumulate binary churn. `metrics.db` is gitignored and rebuilt
  from the dump. ~2 MB/year of facts, no committed aggregates.
- Generated exports, committed for humans: `reports/latest.md` and
  `signatures/catalogue.md`.
- **No migration.** The existing `ci-metrics` JSON is discarded and regenerated
  with `metrics ingest --since 90d`; the Day-0 baseline is recomputed rather than
  carried over. `store.py` is rewritten freely rather than kept behind its current
  interface.
- Rollups are no longer committed: with durable facts every metric is derivable at
  report time under any ruleset version, which removes the frozen-aggregate and
  broken-restore failure modes entirely.

## Movement 3 — signature → issue bridge

New module `ci/metrics/issues.py` and `metrics open-issues` subcommand, run from
the daily digest cron (never from reconciliation).

**Eligibility.** The unit of eligibility is a **(signature, context)** pair, not a
signature. This matters because the same fingerprint can carry different classes
in different contexts — `8c531e6ba49b3382` is `infra.provisioning` on Multipass
and `product.bug` on LXD — so eligibility evaluated per signature would file an
infrastructure failure as a product bug. A pair is issue-worthy when all hold:
- the class **for that context** is `product.bug` or `test.bug`;
- occurrences *within that context* ≥ threshold (start: ≥3 across ≥2 distinct runs);
- not already tracked by an open or claimed issue for that same pair, and not
  closed-as-resolved within a cooldown;
- confidence is rule-derived, not `unknown`;
- the run's base is not stale (see Safety below).
Plus a per-run budget (start: max 3 new issues/day) so a bad rule cannot flood the
tracker.

**Deduplication.** Tracked in the `issues` table, keyed by signature *and
classification context* (see below) — not on `signatures`, which carries no issue
columns. On recurrence the bridge comments on the existing issue with updated
counts and the newest run link instead of opening a second one. A signature
recurring after its issue was closed reopens it with a "regressed" comment rather
than opening a duplicate.

**Idempotency — a local unique index cannot make a remote POST idempotent.** The
partial unique index stops two rows existing; it cannot stop two GitHub issues
existing. Two overlapping runs can both observe "no open issue", both POST, and
only then collide locally — or a worker can crash after the POST but before the
insert, losing the issue number entirely. So the bridge uses claim-then-post:

1. **Claim.** Insert `issue_state='claiming'` with a deterministic
   `idempotency_key` and commit, *before* any network call. The partial unique
   index covers `claiming` as well as `open`, so a competing worker loses here —
   locally, cheaply, before anything external happens.
2. **Post.** Create the issue with the key embedded in the body marker:
   `<!-- ci-signature: <id> ctx=<substrate> key=<idempotency_key> -->`.
3. **Confirm.** Record the number and flip to `open`.

Recovery is what the marker buys: a stale `claiming` row means step 2 or 3 may or
may not have happened, so the bridge **searches GitHub for the key** before
re-posting, and adopts the existing issue if it finds one. The key is in the
issue body precisely so that recovery is a search rather than a guess. A claim
that cannot be resolved after a bounded number of attempts flips to `failed` and
is surfaced in the digest rather than retried forever.

**Issue body** (so the triage skill has what `reproduce.md` needs without a human
re-gathering it): signature ID and class, first/last seen, occurrence count and
affected test count, the matrix cells hit (os/arch/channel/substrate/flavor), retry
outcome (flaky vs systemic), scrubbed excerpt, links to the runs and to the
`inspection-reports-<test>` artifacts from part one, the failing test selector and
source path, and the classifying rule with its `notes`.

**Labels.** `ci-failure`, `signature/<id>`, class label (`product-bug` /
`test-bug`), `kind/bug`, plus reproducibility (`flaky` / `systemic`). Area labels
are left to the existing classify route rather than guessed here.

**Handover to the bot — two things that must be fixed or nothing triggers:**
- `.triage-config.yaml` sets `ignored_users: ["github-actions[bot]"]`, so an issue
  opened by the default Actions identity is ignored outright. Open with `BOT_TOKEN`
  identity, or add an explicit allowance for CI-filed issues.
- `triage.yaml`'s gate routes `issues.opened` to the cheap `classify` job only; the
  self-hosted pipeline requires a maintainer `/triage` or `/reproduce`. That matches
  the chosen policy — no change needed, but the issue body should say so explicitly
  so a maintainer knows the next move.

**Closing the loop.** When an issue closes, record the resolution against the
signature and stop counting it as unowned; if the same fingerprint returns, reopen.
The digest gains an "open CI-filed issues" section with age, so nothing sits
untriaged indefinitely.

**Safety.** `--dry-run` by default in the workflow until the gate passes; every
issue body runs through the same scrubber and `audit-secrets` before posting;
issue creation needs `issues: write` and nothing else.

**A failure on a stale base is not evidence of a live bug.** Before attributing a
failure, the bridge must check the run's SHA against the merge-base with current
`main` and skip — or explicitly caveat — anything whose base predates a commit
already touching the failing area. Without this the bridge will confidently file
issues for bugs that were fixed days ago, because a signature recurring on an old
base is indistinguishable from a real recurrence by fingerprint alone. Part one
already captures `sha` and `source` per job, so the check is cheap; what it costs
is remembering to make it. This was learned the hard way while writing this
document: `linkcheck` failed on the first push of this very PR against an
already-fixed broken link, purely because the branch was cut from a `main` eleven
commits stale, and a duplicate fix PR was opened before the cause was spotted.

## Notes and risks

- **Excerpt depth is the biggest classification limiter.** The harness wraps every
  command through `util.run`, so the excerpt shows the wrapper, not the wrapped
  command's stderr. Deeper extraction is the highest-value Stage B improvement and
  should be scheduled soon after the rule pack lands — issues filed from shallow
  excerpts will be correspondingly shallow.
- **One signature can mean two things.** The generic "failed to meet condition"
  timeout is setup failure on Multipass and a real upgrade failure on LXD. Rules
  must guard on substrate; the tell is that shared-setup failures hit every test
  file at once.
- **Auto-filed issues are a credibility budget.** If the first ten are wrong,
  nobody reads the eleventh. Hence the recurrence threshold, the daily budget, and
  the ≥90% manual-agreement gate before the bridge is enabled for real.
- **Ownership routing stays out of scope.** The issue names a fault domain, not a
  person; `test-owners.yaml` needs organisational agreement first.
- Cross-repo (`k8sd`, `k8s-snap-api`) collection stays out of scope; the data model
  does not preclude it.
- Jira: ticket per movement, `Story` (`Spike` for the excerpt-depth investigation),
  ≤3 points each, under the CI-health epic. Never reference the key on GitHub.

---

# Appendix — how fingerprinting ties into classification

Read out of the PoC branch (`ci/metrics/signature.py`, `router.py`, `rules.py`,
`ci/failure_rules.yaml`) so the plan records what exists, not an idealisation.

## The separation that makes it cheap

**Fingerprint** = content-addressing of failure *text*: "is this the same failure
I've seen before?" **Classification** = attribution: "who owns it?"

The fingerprint is the key; the class is the value. We classify ~10 signatures
per run, not ~2,500 jobs.

```
 GitHub Actions run (nightly: ~2,555 jobs, ~74 failed)
        |
        v
 +--------------+   per-attempt job lists -> retry outcome = ground truth
 |  INGEST      |   (flaky vs systemic, measured not inferred)
 +------+-------+
        v
 +------------------------------------------------+
 | STAGE A -- metadata router  (NO log fetch)      |
 | failed_step_name -> fault domain                |
 +---+-----------------------------+--------------+
     | settled (~46%)              | DEFER
     | "Setup LXD"                 | "Run test_*" says nothing about why
     |   -> infra.provisioning     |
     | no failed step + 404 log    |
     |   -> infra.runner/runner_lost
     v                             v
   RECORDED                 +--------------+
   (never overwritten       |  FETCH LOG   |
    by later stages)        +------+-------+
                                   v
                     +---------------------------+
                     |  FINGERPRINT              |
                     |  extract->scrub->normalise|
                     |          ->hash           |
                     +----------+----------------+
                                v  signature_id (16 hex)
                     +---------------------------+
                     | STAGE B -- rule pack      |
                     | failure_rules.yaml        |
                     | ordered, first-match-wins |
                     +----------+----------------+
                     match -----+----- no match
                       v                  v
                   RECORDED            unknown -> M9 + digest
                                       (Stage C may propose a rule via PR)
```

## Fingerprint construction

```
RAW LOG  (hundreds of KB)
   |
   | (1) EXTRACT -- priority order, and the order IS the design
   |
   |   1. === FAILURES ===  -> walk chain to INNERMOST exception
   |                           + up to 2 frames above it
   |   2. === ERRORS ===    -> pytest puts *fixture* failures here
   |   3. short test summary info
   |   4. ##[error] marker (skipping useless "exit code 1")
   |   5. log tail
   v
"tenacity.RetryError"      <- what the tempting summary line says
"AssertionError: Service kube-proxy should be active, but it is inactive"
                           <- what the innermost exception says  (correct)
   |
   | (2) SCRUB -- before anything is stored. Join tokens excised in place,
   |              not whole-line-killed: dropping the line also drops the
   |              failing command and makes the failure unclassifiable.
   v
   | (3) NORMALISE -- collapse run-to-run variation
   |
   |   k8s-integration-4711-a3f9 -> <instance>
   |   10.1.2.3 / fe80::1        -> <ip> / <ipv6>
   |   2026-09-30T12:08:26Z      -> <ts>
   |   /home/runner/work/...     -.
   |   /home/ubuntu/actions-runner/_work/... -> <workspace>
   |   line 412                  -> line <n>
   |   3 restarts                -> <n> restarts  (counts quantify, not identify)
   |
   |   deliberately NOT blanket digit removal:
   |   "exit code 1" != "exit code 2"
   v
   | (4) HASH  sha256(normalised)[:16]
   v
signature_id = 38c23bd3a1bab957
```

Measured effect of getting extraction right: adding the `=== ERRORS ===` branch
took one weekly run from 46 signatures to 9, and unclassified from 61.7% to 1.7%.
One broken fixture had been fragmenting into dozens of single-occurrence
signatures and burying the root cause.

## What the collapse buys

```
nightly run 34172139128 -- 74 failed jobs
                          |
                          v fingerprint
  +-----------------------------------------------------+
  | c55cefd2ca3f28e7  x36  product.bug/service_not_active|  4 tests, 2 arches
  | 38c23bd3a1bab957  x18  product.bug/service_not_active|
  | 8c531e6ba49b3382  x12  product.bug/upgrade_failure   |
  | 35b12bf41b664182  x 4  product.bug/service_not_active|
  +-----------------------------------------------------+
        74 jobs -> 6 causes. Triage effort becomes proportional
        to cause count, not job count.
```

## Rule matching

Criteria are ANDed: pinned signature, excerpt regex, job metadata.

```
failure_rules.yaml   (ordered -- first match wins)
|
+- infra.runner        "no space left on device"       <- infra FIRST: a dying
+- infra.provisioning                                     host still prints
+- external.dependency "cannot install k8s"               product-shaped errors
+- ci.config
+- product.bug         signature_id: 38c23bd3a1bab957  <- PIN = this exact failure
+- product.bug         excerpt_pattern: "k8s join-cluster"  <- regex = varying family
```

Pins are the default; regexes the exception. A pin is a hash of normalised text,
so it means *this exact failure*. A loose regex placed early silently swallows
unrelated failures and makes the metrics lie. Load-time validation against
`SUBCLASSES` makes a typo fatal rather than silently non-matching, and a
shadowed-rule warning fires when a pin is unreachable behind an earlier one.

## One signature, two meanings

```
signature 8c531e6ba49b3382  "Failed to meet condition"  (generic retry timeout)
              |
      +-------+--------+
 substrate: multipass   substrate: lxd
      |                      |
 sprayed across          concentrated in
 12 test files           test_version_upgrades
      |                      |
      v                      v
 infra.provisioning     product.bug
 /multipass_setup       /upgrade_failure
```

The excerpt cannot separate these; only the substrate can. Generalisable tell:
when one signature hits every test file at once it is shared setup, not a product
regression in each of them. If the substrate cannot be parsed out of the job name,
the occurrence must fall through to `unknown` and show up in M9 rather than be
scored as a product bug on a guess.

## The harness constraint

Everything runs through `util.run`, so nearly every product failure surfaces as
the *same* `CalledProcessError` frame. The only discriminator left is the wrapped
command (`k8s x-wait-for`, `k8s join-cluster`, `k8s bootstrap`, `snap install k8s`).
Six such rules attributed 1,801 previously-unknown failures. It also bounds what
classification can honestly claim: the excerpt shows the wrapper, not the wrapped
command's stderr — which is why `excerpt-depth` is the highest-value follow-up.

## Why the signature is the primary key for movement 3

```
signature_id --+-> dedup key   <!-- ci-signature: 38c23bd3a1bab957 -->
               +-> issue labels  signature/38c23bd3a1bab957
               +-> recurrence counter -> threshold (>=3 across >=2 runs)
               +-> reopen key   same hash after close = regression
```

One stable hash avoids opening 74 issues for 6 bugs, avoids opening the same issue
again next week, and is how we know a fix actually held.

---

# Appendix — SQLite schema (movement 2)

Designed from the use case rather than from the PoC's JSON shapes. Data is
regenerated by re-ingesting from the Actions API, so nothing here is constrained
by migration.

## The insight that reshapes it

The PoC stores a 4 KB excerpt per failed job; that is the only real volume
driver. Strip the excerpt and a failure row is ~120 bytes:

```
1,105 failures/month x 12 = ~13k rows/year x 120 B ~= 1.6 MB/year
```

So **keep every fact forever and expire only the excerpts.** That deletes a whole
family of problems: no frozen rollups, no `runs_covered` guard, no refusal to
compare across rule pack versions, no falsified trend when an artifact restore
breaks. It works because of the design's own premise -- *we classify signatures,
not jobs* -- so reclassification only ever needs **one exemplar excerpt per
signature**, which is durable.

```
 DURABLE (committed, ~2 MB/yr)          EPHEMERAL (90-day artifact)
 -----------------------------          ---------------------------
 runs, jobs, signatures,                 job_excerpts (4 KB x 13k/yr = 53 MB)
 signature_classifications,              job_steps
 configs, tests, runners, issues
        |                                      |
        +-- every metric derivable ------------+ only raw evidence expires
            at any time, under any ruleset
```

## Schema

```sql
PRAGMA foreign_keys = ON;

CREATE TABLE schema_meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);

-- == dimensions: no JSON columns anywhere =============================
--
-- CONVENTION: dimension columns that participate in a UNIQUE or PRIMARY KEY
-- are NOT NULL and use '' for "absent" (and '*' for "any"). SQLite treats
-- NULLs as DISTINCT in UNIQUE constraints, so a nullable dimension silently
-- stops deduplicating: artifact builds carry no channel, and nullable
-- `channel` would mint a fresh configs row per run and inflate every
-- configs_affected count. Sentinels, not NULLs, in every key.

CREATE TABLE workflows (
    workflow_id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    file TEXT
);

-- The matrix cell, deduplicated. Makes "which configs does this signature
-- hit" a join instead of parsing a JSON array, and makes the substrate
-- guard (one signature, two meanings) a first-class query.
CREATE TABLE configs (
    config_id INTEGER PRIMARY KEY,
    os        TEXT NOT NULL DEFAULT '',
    arch      TEXT NOT NULL DEFAULT '',
    channel   TEXT NOT NULL DEFAULT '',   -- '' for artifact builds
    flavor    TEXT NOT NULL DEFAULT '',
    substrate TEXT NOT NULL DEFAULT '',   -- '' when unparseable => `unknown`
    UNIQUE (os, arch, channel, flavor, substrate)
);

CREATE TABLE tests (
    test_id INTEGER PRIMARY KEY,
    nodeid  TEXT NOT NULL UNIQUE,   -- seed row id 0, nodeid '' = no test
    file    TEXT,
    owner   TEXT            -- reserved for test-owners.yaml; NULL today
);

CREATE TABLE runners (
    runner_id   INTEGER PRIMARY KEY,
    label_key   TEXT NOT NULL UNIQUE,   -- sorted labels, canonical
    self_hosted INTEGER NOT NULL
);
CREATE TABLE runner_labels (
    runner_id INTEGER NOT NULL REFERENCES runners ON DELETE CASCADE,
    label     TEXT NOT NULL,
    PRIMARY KEY (runner_id, label)
);

-- Stage A verdicts, as a dimension rather than a parsed string. The class is
-- stored, never derived from the rule id: ids are dotted (`infra.runner.lost`)
-- and classes are dotted too, so splitting on the first '.' yields `infra`,
-- which is not a valid failure_class at all. One CHECK, one place.
CREATE TABLE router_rules (
    router_rule_id TEXT PRIMARY KEY,
    failure_class  TEXT NOT NULL,
    subclass       TEXT,
    router_version INTEGER NOT NULL,
    CHECK (failure_class IN ('infra.runner','infra.provisioning',
        'external.dependency','ci.config','product.bug','test.bug','unknown'))
);

-- == facts: durable ===================================================
CREATE TABLE runs (
    run_id  INTEGER NOT NULL,
    attempt INTEGER NOT NULL,          -- per-ATTEMPT: flake truth needs all
    workflow_id INTEGER NOT NULL REFERENCES workflows,
    event TEXT, head_branch TEXT, head_sha TEXT, pr_number INTEGER,
    created_at TEXT, started_at TEXT, completed_at TEXT NOT NULL,
    conclusion TEXT NOT NULL,          -- in-flight runs CANNOT be stored

    jobs_total     INTEGER NOT NULL,
    jobs_success   INTEGER NOT NULL,
    jobs_failure   INTEGER NOT NULL,
    jobs_skipped   INTEGER NOT NULL,
    jobs_cancelled INTEGER NOT NULL,
    runner_minutes_total  REAL NOT NULL DEFAULT 0,
    runner_minutes_failed REAL NOT NULL DEFAULT 0,

    ingest_version INTEGER NOT NULL,
    PRIMARY KEY (run_id, attempt)
);

-- M10. Test counts differ per cell: each channel collects from its own
-- release branch, so a shrinking denominator is invisible without this.
CREATE TABLE run_tests_collected (
    run_id INTEGER NOT NULL, attempt INTEGER NOT NULL,
    config_id INTEGER NOT NULL REFERENCES configs,
    count INTEGER NOT NULL,
    PRIMARY KEY (run_id, attempt, config_id),
    FOREIGN KEY (run_id, attempt) REFERENCES runs ON DELETE CASCADE
);

-- M11. A failed "Prepare Environment" means its whole matrix never ran,
-- so a naive failure rate IMPROVES when provisioning breaks badly enough.
CREATE TABLE run_blocked_jobs (
    run_id INTEGER NOT NULL, attempt INTEGER NOT NULL,
    prepare_job_name TEXT NOT NULL,
    jobs_not_run INTEGER NOT NULL,
    PRIMARY KEY (run_id, attempt, prepare_job_name),
    FOREIGN KEY (run_id, attempt) REFERENCES runs ON DELETE CASCADE
);

-- One row per FAILED job. This IS the occurrence fact table.
CREATE TABLE jobs (
    job_id  INTEGER PRIMARY KEY,       -- globally unique in GitHub
    run_id  INTEGER NOT NULL, attempt INTEGER NOT NULL,
    config_id INTEGER NOT NULL REFERENCES configs,
    test_id   INTEGER NOT NULL DEFAULT 0 REFERENCES tests,  -- 0 = no test
    runner_id INTEGER REFERENCES runners,
    job_name TEXT NOT NULL, caller_job TEXT, html_url TEXT NOT NULL,
    started_at TEXT, completed_at TEXT, duration_s REAL,

    failed_step_name   TEXT,           -- NULL + log_available=0 => runner lost
    failed_step_number INTEGER,
    log_available      INTEGER,

    -- Exactly one attribution path, or neither (deferred and unmatched).
    signature_id   TEXT REFERENCES signatures,   -- Stage B
    router_rule_id TEXT REFERENCES router_rules, -- Stage A, metadata-only
    FOREIGN KEY (run_id, attempt) REFERENCES runs ON DELETE CASCADE,
    CHECK (signature_id IS NULL OR router_rule_id IS NULL)
);
CREATE INDEX jobs_by_signature ON jobs(signature_id);
CREATE INDEX jobs_by_run       ON jobs(run_id, attempt);
CREATE INDEX jobs_by_test      ON jobs(test_id);

-- == signatures: the durable unit of work =============================
CREATE TABLE signatures (
    signature_id TEXT PRIMARY KEY,     -- sha256(normalised)[:16]
    normalised_pattern TEXT NOT NULL,
    exemplar_excerpt   TEXT NOT NULL,  -- ONE per signature, kept forever:
    exemplar_job_id    INTEGER,        -- this is what makes reclassify work
    normaliser_version INTEGER NOT NULL,   -- changing these changes IDs
    extractor_version  INTEGER NOT NULL,
    state TEXT NOT NULL DEFAULT 'active',
    quarantined_at TEXT, quarantine_expires_at TEXT,
    CHECK (state IN ('active','quarantined','fixed','wontfix'))
);

-- Append-only. Keeps the history of how a signature was classified as the
-- rule pack evolved, which is what lets us measure automated-diagnosis
-- accuracy against human triage decisions rather than assert it.
--
-- Keyed by signature AND CONTEXT, because a fingerprint is not always a
-- diagnosis: 8c531e6ba49b3382 ("failed to meet condition") is
-- infra.provisioning on Multipass and product.bug on LXD. Keying on the
-- signature alone would let the bridge file an infrastructure failure as a
-- product bug. `substrate='*'` means "any substrate" -- the common case --
-- and a concrete substrate wins over it. The sentinel is deliberate: SQLite
-- allows NULL in a PRIMARY KEY and treats NULLs as distinct, so a nullable
-- context column would admit unlimited duplicate "any" rows.
--
-- Substrate is the only discriminator the evidence has so far demanded. If a
-- second one appears, this generalises to a context_id FK into a contexts
-- table; it does not generalise by adding more nullable columns here.
CREATE TABLE signature_classifications (
    signature_id  TEXT NOT NULL REFERENCES signatures ON DELETE CASCADE,
    substrate     TEXT NOT NULL DEFAULT '*',   -- '*' = any; else exact match
    ruleset_version INTEGER NOT NULL,
    failure_class TEXT NOT NULL,
    subclass      TEXT,
    rule_id       TEXT,
    confidence    REAL,
    classified_by TEXT NOT NULL,       -- rules|human  (never 'llm')
    classified_at TEXT NOT NULL,
    PRIMARY KEY (signature_id, substrate, ruleset_version, classified_by),
    CHECK (failure_class IN ('infra.runner','infra.provisioning',
        'external.dependency','ci.config','product.bug','test.bug','unknown')),
    CHECK (classified_by IN ('rules','human'))
);

-- == ephemeral: raw evidence only (90-day artifact) ===================
CREATE TABLE job_excerpts (
    job_id INTEGER PRIMARY KEY REFERENCES jobs ON DELETE CASCADE,
    excerpt TEXT NOT NULL,             -- scrubbed, <=4 KB
    source  TEXT NOT NULL,             -- failures-block|errors-block|...
    extractor_version INTEGER NOT NULL
);
CREATE TABLE job_steps (
    job_id INTEGER NOT NULL REFERENCES jobs ON DELETE CASCADE,
    number INTEGER NOT NULL,
    name TEXT NOT NULL,
    conclusion TEXT,                   -- NULL = never ran (runner-lost tell)
    duration_s REAL,
    PRIMARY KEY (job_id, number)
);

-- Positive retry evidence. `jobs` holds only FAILED jobs, so "absent from
-- attempt N+1" is not proof of a pass -- it equally covers skipped, cancelled,
-- and jobs that never ran because their matrix was blocked (M11). Deriving
-- flakiness from absence would relabel every provisioning outage as a flake.
-- Written only for cells that failed on a previous attempt, so volume is
-- bounded by prior failures rather than by matrix size.
CREATE TABLE retry_outcomes (
    run_id    INTEGER NOT NULL,
    attempt   INTEGER NOT NULL,        -- the RETRY attempt, i.e. N+1
    config_id INTEGER NOT NULL REFERENCES configs,
    test_id   INTEGER NOT NULL DEFAULT 0 REFERENCES tests,
    conclusion TEXT NOT NULL,          -- only 'success' makes a flake
    PRIMARY KEY (run_id, attempt, config_id, test_id),
    FOREIGN KEY (run_id, attempt) REFERENCES runs ON DELETE CASCADE,
    CHECK (conclusion IN ('success','failure','skipped','cancelled','not_run'))
);

-- == the issue bridge =================================================
-- `claiming` exists because a local index cannot make a remote POST
-- idempotent. The row is written and committed BEFORE the API call, so a
-- competing worker loses the race locally and cheaply; `idempotency_key` is
-- echoed into the issue body so a crash between POST and confirm is
-- recoverable by SEARCHING GitHub for the key rather than guessing.
CREATE TABLE issues (
    claim_id     INTEGER PRIMARY KEY,
    signature_id TEXT NOT NULL REFERENCES signatures ON DELETE CASCADE,
    substrate    TEXT NOT NULL DEFAULT '*',  -- same context scope as the class
    idempotency_key TEXT NOT NULL UNIQUE,
    issue_number INTEGER,              -- NULL until the POST returns
    issue_state  TEXT NOT NULL,        -- claiming|open|closed|failed
    claimed_at TEXT NOT NULL,
    opened_at TEXT, closed_at TEXT, closed_reason TEXT,
    reopened_count INTEGER NOT NULL DEFAULT 0,
    post_attempts  INTEGER NOT NULL DEFAULT 0,
    last_comment_at TEXT,
    last_reported_occurrences INTEGER NOT NULL DEFAULT 0,
    occurrences_at_open INTEGER,
    opened_by_ruleset_version INTEGER,
    CHECK (issue_state IN ('claiming','open','closed','failed')),
    -- an issue that is live or settled must know its number
    CHECK (issue_state IN ('claiming','failed') OR issue_number IS NOT NULL)
);
-- One live issue per (signature, context). Covers `claiming` too -- that is
-- the whole point: the claim, not the POST, is what gets serialised.
CREATE UNIQUE INDEX one_live_issue_per_signature
    ON issues(signature_id, substrate)
    WHERE issue_state IN ('claiming','open');
CREATE UNIQUE INDEX one_issue_number
    ON issues(issue_number) WHERE issue_number IS NOT NULL;
```

## Views -- every metric derived, nothing frozen

```sql
-- Current class per (signature, context). Human override always wins, then
-- the newest ruleset. Specificity is resolved in v_job_class, not here.
CREATE VIEW v_signature_class AS
SELECT signature_id, substrate, failure_class, subclass, rule_id,
       confidence, classified_by
FROM (SELECT *, ROW_NUMBER() OVER (
          PARTITION BY signature_id, substrate
          ORDER BY (classified_by='human') DESC, ruleset_version DESC
      ) rn FROM signature_classifications)
WHERE rn = 1;

-- One class per failed job, whichever stage settled it.
-- The signature join is CONTEXT-AWARE: an exact substrate match outranks the
-- '*' fallback, so the Multipass/LXD split survives into every metric and
-- into issue eligibility. Router class is read from router_rules, never
-- parsed out of the rule id.
CREATE VIEW v_job_class AS
SELECT job_id, run_id, attempt, config_id, test_id, signature_id,
       failure_class, subclass, classified_by
FROM (
  SELECT j.job_id, j.run_id, j.attempt, j.config_id, j.test_id, j.signature_id,
         COALESCE(sc.failure_class, rr.failure_class, 'unknown') AS failure_class,
         COALESCE(sc.subclass, rr.subclass)                      AS subclass,
         CASE WHEN sc.signature_id IS NOT NULL THEN sc.classified_by
              WHEN rr.router_rule_id IS NOT NULL THEN 'router'
              ELSE 'none' END                                    AS classified_by,
         ROW_NUMBER() OVER (PARTITION BY j.job_id
                            ORDER BY (sc.substrate <> '*') DESC) AS rn
  FROM jobs j
  JOIN configs cfg USING (config_id)
  LEFT JOIN router_rules rr USING (router_rule_id)
  LEFT JOIN v_signature_class sc
         ON sc.signature_id = j.signature_id
        AND sc.substrate IN ('*', cfg.substrate)
) WHERE rn = 1;

-- Replaces the frozen signature_stats table entirely.
CREATE VIEW v_signature_stats AS
SELECT j.signature_id,
       COUNT(*)                    AS occurrences,
       COUNT(DISTINCT j.run_id)    AS runs_seen,
       COUNT(DISTINCT j.test_id)   AS tests_affected,
       COUNT(DISTINCT j.config_id) AS configs_affected,
       MIN(r.started_at) AS first_seen, MAX(r.started_at) AS last_seen
FROM jobs j JOIN runs r USING (run_id, attempt)
WHERE j.signature_id IS NOT NULL GROUP BY j.signature_id;

-- Per-context stats, which is the grain issue eligibility actually uses.
CREATE VIEW v_signature_context_stats AS
SELECT j.signature_id, cfg.substrate,
       COUNT(*) AS occurrences, COUNT(DISTINCT j.run_id) AS runs_seen,
       MIN(r.started_at) AS first_seen, MAX(r.started_at) AS last_seen
FROM jobs j JOIN runs r USING (run_id, attempt)
JOIN configs cfg USING (config_id)
WHERE j.signature_id IS NOT NULL GROUP BY j.signature_id, cfg.substrate;

-- Flake = failed on attempt N and PROVABLY SUCCEEDED on N+1, same cell and
-- test. Absence is not evidence: a cell that was skipped, cancelled, or never
-- ran because its matrix was blocked is absent too, and treating that as a
-- pass would relabel provisioning outages as flakes. Hence the explicit join
-- to retry_outcomes rather than a NOT EXISTS over jobs.
CREATE VIEW v_flakes AS
SELECT j.signature_id, j.run_id, j.attempt, j.config_id, j.test_id
FROM jobs j
JOIN retry_outcomes ro
  ON ro.run_id = j.run_id AND ro.attempt = j.attempt + 1
 AND ro.config_id = j.config_id AND ro.test_id = j.test_id
WHERE ro.conclusion = 'success';
```

## What changed against the PoC, and why

| PoC | Now | Why |
|---|---|---|
| `signature_stats` frozen table | `v_signature_stats` view | facts are durable, so aggregates cannot drift from them |
| `rollups` committed, `runs_covered` guard | derived at report time, **not committed** | removes the "relabel a `--since` window as a month" failure mode entirely |
| cross-`ruleset_version` comparison refused | recompute all history under the new ruleset | strictly more honest; git history of `metrics.sql` preserves what we believed |
| class on `jobs` | class on `signatures` **per context** | matches the premise (we classify signatures) without collapsing the Multipass/LXD split into one verdict |
| single classification | `signature_classifications` append-only | enables "track diagnostic accuracy against human triage" from the vision |
| 4 JSON array columns | `configs`, `tests`, `runner_labels`, `run_blocked_jobs` | the root cause of "ugly json" -- gone |
| `conclusion` nullable | `NOT NULL` | the in-flight-run bug becomes unrepresentable |
| `retried`/`retry_outcome` flags | `v_flakes` over `retry_outcomes` | a stored flag goes stale; a derivation cannot -- but it needs *positive* pass evidence, not absence |
| nullable dimension columns | `NOT NULL` + `''`/`'*'` sentinels | SQLite treats NULLs as distinct in UNIQUE, so nullable keys silently stop deduplicating |
| router class parsed from rule id | `router_rules` lookup table | splitting `infra.runner.lost` on the first `.` yields `infra`, not a valid class |
| `issues` keyed on `(signature, number)` | claim-then-post with `idempotency_key` | a local index cannot make a remote POST idempotent |

## Storage

- `metrics.sql` = `sqlite3 .dump` of the durable tables only, stable ordering.
  ~2 MB/year of facts and no aggregates. `metrics.db` gitignored, rebuilt from
  the dump.
- Generated exports, committed for humans: `reports/latest.md`,
  `signatures/catalogue.md`.
- Excerpts and full step detail stay in the 90-day Actions artifact. What is
  promoted and kept forever: one exemplar excerpt per signature (all that
  reclassification ever needs), and the step facts a classification actually
  turns on — `failed_step_name`, `failed_step_number`, `log_available` — which
  live on `jobs`, not in `job_steps`. So the runner-lost and retry
  interpretations survive the artifact window even though the table backing
  them does not.
- No migration path: data is regenerated with `metrics ingest --since 90d`.
