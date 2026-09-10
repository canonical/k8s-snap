# Proposal information

- **Index**: 003

- **Status**: **DRAFTING**

- **Name**: CI failure classification and metrics

- **Owner**: Louise Schmidtgen / [@louiseschmidtgen](https://github.com/louiseschmidtgen)

# Proposal Details

## Summary

CI in this repository is red as a steady state, and we cannot currently say
*why*. Scheduled suites have not gone green in weeks, the nightly alert posts a
several-thousand-node status tree into a channel nobody acts on, and there is
no way to tell whether a given failure is our bug, our infrastructure, or
somebody else's outage. Without attribution there is nothing to assign, and
without history there is no way to demonstrate that any fix helped.

This proposal adds a `metrics` subcommand to the existing `k8s-ci` tool that
classifies every failed CI job into a fault domain automatically, fingerprints
failures so that hundreds of jobs collapse to a handful of distinct causes, and
maintains a committed history so that improvement can be measured rather than
asserted.

Classification is deterministic and rule-based. An optional LLM stage is
proposed separately, and — importantly — it never writes a recorded
classification; it only proposes rules for humans to merge.

## Rationale

### The measured problem

Numbers below were mined from the GitHub Actions API for this repository during
design and re-confirmed during implementation against a 90-day backfill.

| Observation | Value |
|---|---|
| Nightly runs green | 0 of the last 7 |
| Weekly runs green | 0 of the last 4 |
| PR checks passing first time | ~20% over 30 runs |
| Nightly failures that are infra/setup, not tests | ~46% |
| Weekly failures that are infra/setup | ~68% |
| Share of nightly failures in the top 5 signatures | ~47% |

Two conclusions follow, and they point in the same direction.

**Most red is not a product bug.** Roughly half of nightly failures and
two-thirds of weekly failures happen before any product code runs — LXD
provisioning, snap store downloads, runners disappearing. The independent
measurement taken during implementation agrees: across the 90-day nightly
backfill, `infra.provisioning/lxd_setup` alone accounts for 40.4% of failures
and `external.dependency/snap_store` for a further 12.2%. Treating these as
test failures sends them to the wrong people, who correctly ignore them.

**A small number of causes produce most of the noise.** The top five
fingerprints cover about half of all nightly failures. In one nightly run, 72
failed jobs reduced to 6 distinct causes. Triage effort today is proportional
to *job count*; it should be proportional to *cause count*, which is two orders
of magnitude smaller.

### Why the current alerting does not help

The nightly digest reports status for every job in the matrix. It is accurate
and unusable: it says what is red, never who owns it, and because red is
normal, the alert carries no information. Anything that fires every day with
the same content stops being read. A report only changes behaviour if it names
an owner and shows a delta.

### What this enables

- Failures routed to whoever can actually fix them, by fault domain.
- Triage against ~10 distinct causes per run instead of ~100 failed jobs.
- A frozen Day-0 baseline, so "CI got better" becomes checkable.
- Detection of the failure mode where metrics improve because *less ran* —
  which is otherwise indistinguishable from real progress.

## User facing changes

None for users of the snap. This is CI tooling only.

For developers, `k8s-ci` gains a `metrics` subcommand group:

```
tox -e k8s-ci -- metrics ingest      # collect run + job metadata
tox -e k8s-ci -- metrics fetch-logs  # fetch and fingerprint failure logs
tox -e k8s-ci -- metrics classify    # apply router + rules
tox -e k8s-ci -- metrics reclassify  # replay classification offline
tox -e k8s-ci -- metrics rollup      # aggregate into periodic metrics
tox -e k8s-ci -- metrics report      # render a digest or trend report
tox -e k8s-ci -- metrics audit-secrets
```

The Mattermost digest changes shape. Before, an emoji tree of every job:

```
Nightly test results:
:red_circle: ubuntu:22.04 amd64 1.32-classic/edge
  :red_circle: tests/test_smoke.py::test_smoke
  :white_check_mark: tests/test_dns.py::test_dns
  ... (2,553 more lines)
```

After, attribution and delta first, with the tree demoted to a threaded reply:

```
#### CI failure report -- 2026-09
1105 failed job(s) of 24645 across 11 run(s) -- 4.5%  (-3.7 pp ▼)

**Who owns today's failures** (573 inspected)
- `external.dependency` -- 422 (73.6%) -- external, usually wait/retry
- `product.bug`         --  85 (14.8%) -- product team
- `infra.provisioning`  --  44  (7.7%) -- CI/infra
- `infra.runner`        --  21  (3.7%) -- CI/infra
- `unknown`             --   1  (0.2%) -- unowned, needs triage

**Top signatures**
- `8c531e6ba49b3382` x50 -- product.bug/upgrade_failure (32 tests affected)
- `c55cefd2ca3f28e7` x40 -- product.bug/service_not_active (4 tests affected)
...

**Integrity checks**
- unclassified: 0.2% of 573 inspected failure(s)
- inspection coverage: 51.9% of failures have been inspected
```

## Alternative solutions

### A hosted CI analytics product

Rejected. It would need a GitHub App with broad read access to a repository
that builds a security-sensitive artifact, adds an external dependency and cost
centre, and cannot be taught this project's specific vocabulary — LXD substrate
quirks, snap channel semantics, harness retry idioms. The classification
problem here is domain-specific; the generic product would leave us with a
prettier dashboard and the same unattributed failures.

### Classify with an LLM, per job

Rejected as the primary mechanism, on three independent grounds.

*Cost and scale.* A nightly run produces up to ~2,500 jobs. Per-job inference
is thousands of calls per night for a question that is usually answered by the
step name alone.

*Non-determinism.* If the recorded class can change between runs on identical
input, the trend line measures model variance as well as CI health. A metric
whose definition moves cannot demonstrate improvement.

*It is not needed for the common case.* Roughly half of failures are settled by
job metadata alone, with no log fetch: the failed step is `Setup LXD`, or there
is no failed step at all and the log blob 404s (a lost runner). Most of the
remainder collapse onto a handful of repeat fingerprints.

An LLM stage is still proposed, but scoped to genuinely novel signatures and
constrained to *proposing rules*, never to writing recorded data.

### Store metrics in an external database

Rejected for v1. It solves a problem we do not yet have — the committed volume
is single-digit MB per year — while adding credentials, availability
requirements and a second source of truth. Git gives us history, review and
provenance for free. If query volume ever justifies a database, the JSON
rollups are trivially loadable into one.

### Commit metrics data to `main`

Rejected, and the reason is worth recording because it is not obvious.

`main` is protected, so `GITHUB_TOKEN` cannot push to it. Routing through a
pull request is worse and, critically, **circular**: `lint_and_integration.yaml`
only ignores `docs/**`, so a commit touching `ci/` triggers the full build and
e2e matrix, and the automerge job requires every check to pass. At a ~20% PR
pass rate, the daily metrics PR would rarely merge — the system would be
blocked by exactly the redness it exists to measure.

Data therefore lands on an unprotected long-lived `ci-metrics` branch, cut from
`main` rather than orphaned so that `git log`, `git diff` and cherry-pick behave
normally and the tree deltas against `main` instead of duplicating it.

## Out of scope

- **Ownership routing.** Mapping tests and classes to teams (`test-owners.yaml`)
  needs organisational agreement, not code. The digest names a fault domain;
  turning that into a person is deferred.
- **Flake quarantine.** Auto-skipping known-flaky tests is a behaviour change
  requiring the measurement substrate to exist first, and carries a real risk of
  hiding regressions. Deferred, with metric M12 reserved to track it.
- **PR-comment bot.** Posting classification onto pull requests is deferred
  until precision is demonstrated at the validation gate.
- **Fixing the failures.** This proposal measures and attributes; it does not
  fix `test_version_downgrades_with_rollback`.
- **Cross-repository metrics.** `k8sd` and `k8s-snap-api` have their own CI.
  The data model does not preclude this; the collection does not attempt it.

## Implementation Details

### API Changes

None. This is read-only against the GitHub Actions API.

### CLI Changes

None to the `k8s` CLI. `k8s-ci` gains the `metrics` subcommand group listed
above, alongside the existing `charm`, `deps`, `docs` and `mattermost` groups.

### Database Changes

None. Data is JSON on the `ci-metrics` branch and in Actions artifacts.

### Configuration Changes

New workflow `.github/workflows/ci-metrics.yaml`, triggered by `workflow_run`
on Nightly / Weekly / Lint and integration, plus an hourly reconciliation cron
and a daily digest cron. Requires `actions: read` and `contents: write`, and
reuses the existing `MATTERMOST_WEBHOOK_URL` secret.

New file `ci/failure_rules.yaml` — the reviewable classification surface.

### Documentation Changes

`ci/metrics-data/README.md` on the `ci-metrics` branch describing the layout
and retention. Team-facing documentation of the taxonomy is folded into
`failure_rules.yaml`, where every rule carries a `notes` field explaining why
its class was chosen; that keeps the explanation next to the thing it explains.

### Testing

Unit tests under `ci/tests/metrics/` cover the job-name parser, the metadata
router, the log normaliser and fingerprinter, the secret scrubber, the rule
pack loader, rollup aggregation and report rendering. The suite is
deterministic and offline; no test makes a network call.

Three properties are pinned deliberately because they are what make the numbers
trustworthy rather than merely present:

- Rates return null, not zero, when nothing was measured — a missing
  denominator must not read as a good result.
- Failures whose logs were never fetched count as *uninspected*, not
  *unclassified* — otherwise the classification rate tracks fetch coverage and
  can be improved by fetching less.
- Rollups produced under different rule pack versions refuse to be compared —
  otherwise classifier churn is reported as CI improvement.

Beyond unit tests, a **validation gate** is run against live data before the
workflow is enabled:

| # | Check | Criterion |
|---|---|---|
| 1 | Ingest a known run | job counts match the API exactly |
| 2 | Classify a known nightly run | repeat failures collapse to one signature each |
| 3 | A job with no failed step and a 404 log | classified `infra.runner/runner_lost`, not `unknown` |
| 4 | Classify a known weekly run | snap-store and conformance failures attributed correctly |
| 5 | Manual audit of 20 random failures | ≥90% agreement with the classifier |
| 6 | Render the digest | leads with attribution; deltas correct |
| 7 | Scan all stored excerpts | **zero** secrets — hard gate |
| 8 | Trend report over the backfill | renders; integrity metrics present |
| 9 | Committed volume | tracks the ~1–3 MB/year projection |

### Considerations for backwards compatibility

No product surface is affected. Within the metrics system itself, three version
stamps are recorded on every datum: `normaliser_version`, `ruleset_version`,
`rollup_version`. Changing the normaliser changes signature IDs, which would
otherwise silently reset every trend; stamping makes that visible and makes
cross-version comparison refuse rather than mislead.

`reclassify` replays classification over stored excerpts with no API traffic,
so rule changes can be applied to the full retained history — including runs
whose GitHub logs have since expired.

### Implementation notes and guidelines

#### Taxonomy

Two axes, because "what broke" and "how it broke" answer different questions
and collapsing them loses one.

*Axis A — fault domain* (who owns it):

| Class | Meaning |
|---|---|
| `infra.runner` | Runner lost, unavailable, or died mid-job |
| `infra.provisioning` | LXD/Multipass/environment setup failed |
| `external.dependency` | Snap store, registry, GitHub API, DNS |
| `ci.config` | Our workflow or tooling is wrong |
| `product.bug` | The product genuinely misbehaved |
| `test.bug` | The test is wrong or too strict |
| `unknown` | Not yet attributable |

*Axis B — reproducibility*: `systemic` (fails everywhere), `config-specific`
(one OS/arch/channel), `flaky` (recovered on retry), `new`.

Axis B is measurable for free: GitHub exposes per-attempt job lists at
`/runs/{id}/attempts/{n}/jobs`, which gives retry outcomes as ground truth
rather than inference.

#### Classification cascade

The organising decision is to **classify signatures, not jobs**. A signature is
a fingerprint of the normalised error text. Jobs number in the thousands;
signatures number in the tens. Everything downstream is cheaper for it.

**Stage A — metadata router.** Deterministic, table-driven, no log fetches.
Maps the failed step name to a fault domain. `Setup LXD` is provisioning;
`Set up job` is the runner; `Run test_*` is deferred because a failed test step
says nothing about why.

Stage A also catches a case a log-based classifier structurally cannot see: the
job concludes `failure`, *no* step reports failure, trailing steps are `null`,
and the log API returns 404. That is a lost runner. Verified against real job
`101894558721`. Anything that only reads logs scores these `unknown` forever.

Two subtleties learned from real data:

- Step names are not always unambiguous. `Download k8s-snap` is a composite
  action with two mutually exclusive modes behind one name: channel mode calls
  `snap download` (genuinely the store), artifact mode pulls *our own* build
  output (genuinely ours). Attributing both to the store would tell the team to
  wait out a problem they own. The rule guards on whether the job carries a
  channel.
- Router verdicts are never overwritten by later stages. Metadata outranks log
  text, because log text is the thing most likely to be misleading.

**Stage B — signature and rule pack.** For deferred jobs, fetch the log, slice
to the failed step, extract the salient error, scrub secrets, normalise, and
hash to a 16-hex signature. Then match against `ci/failure_rules.yaml`, an
ordered first-match-wins list of rules pinned to signature IDs.

Extraction priority matters more than it looks. pytest reports *fixture*
failures under `=== ERRORS ===`, not `=== FAILURES ===`. An extractor that only
knows about FAILURES turns one broken fixture into dozens of unrelated
single-occurrence signatures and buries the root cause. Adding ERRORS
extraction collapsed one weekly run from 46 signatures to 9, and its
unclassified rate from 61.7% to 1.7%.

Normalisation must collapse counts as well as identifiers. `kubelet (3
restarts)` and `kubelet (1 restarts)` are the same failure and must hash the
same.

Ordering encodes precedence, because failure text is not mutually exclusive.
One real signature — a generic "failed to meet condition" retry timeout —
means *two different things*: on Multipass it is a setup failure sprayed across
a dozen unrelated test files; on LXD it is a genuine upgrade failure
concentrated in one. The excerpt cannot distinguish them; the substrate can.
The generalisable tell: **when one signature hits every test file at once, it
is shared setup, not a product regression in each of them.**

**Stage C — LLM adjudication (deferred, opt-in).** Invoked once per *novel
signature*, never per job, and cached against the signature ID forever. Its
output is a **pull request against `failure_rules.yaml`** — never a recorded
class. A human merges it; thereafter Stage B owns that signature
deterministically and the model is never consulted for it again.

This keeps the model a bounded accelerator. If it is disallowed or unavailable,
Part 1 stands alone unchanged: novel signatures land in `unknown` and surface
in the digest as "N new unowned signatures", which is the honest answer.

#### Secret handling

Excerpts are derived from job logs and land on a branch anyone can read, so the
scrubber is a hard gate, not a nicety. It is written and fixture-tested before
any excerpt is stored, removes token shapes, JWTs, `Authorization` headers,
values of `*_TOKEN` / `*_PASSWORD` variables and cloud-init bodies, and
truncates to 4 KB. `metrics audit-secrets` re-scans everything stored and runs
in CI **before** anything is committed or uploaded, so a scrubber regression
cannot publish credentials.

#### Metrics

*Headline* — M1 scheduled green rate, M2 PR first-pass rate, M3 job failure
rate (overall and by class), M4 flake rate.

*Diagnostic* — M5 top-5 signature concentration, M6 signature age, M7 unowned
signatures, M8 runner minutes spent on failed jobs.

*Integrity* — M9 unclassified rate and inspection coverage, M10 tests collected
per matrix cell, M11 jobs that never ran because an upstream job failed, M12
quarantined signatures.

The integrity tier is not optional. Every headline metric can be improved by
running less rather than by breaking less. A failed `Prepare Environment`
produces *zero* downstream jobs, so a naive failure rate **improves** when
provisioning breaks badly enough. M10 and M11 exist so that "CI got better" can
always be checked against "or did we just stop running things", and they are
surfaced in the digest whenever they move.

#### Storage and volume

| Data | Where | Retention |
|---|---|---|
| Daily/monthly rollups | committed to `ci-metrics` | forever |
| Signature catalogue | committed | forever |
| Weekly trend reports | committed | forever |
| Per-run records + excerpts | Actions artifact | 90 days |

Roughly 1–3 MB/year committed against a 14 MB repository. The 90-day excerpt
window bounds how far `reclassify` can reach; that matches GitHub's own log
retention, so nothing is really lost. Rollups older than the window are frozen
under their original `ruleset_version` rather than retroactively rewritten.

#### Collection

`workflow_run` events give low latency but are **best-effort and dropped under
load** — for a metrics system, silently losing runs means understating failure.
An hourly reconciliation pass is what makes collection complete; the event path
is only an optimisation.

The digest posts on one daily cron, never on reconciliation. Posting hourly is
precisely how the existing nightly alert became background noise.

Pagination is parallelised: page 1 yields `total_count`, from which the
remaining page count is computed and fetched concurrently, reassembled by page
number to preserve order. This took a 2,555-job nightly run from ~150 s to 19 s,
which is what makes a 90-day backfill practical at all.

#### Phasing

**Part 1 (~11 days)** — everything above except Stage C. Scaffolding, ingest,
backfill and Day-0 baseline, metadata router, signatures and scrubber, rule
pack, rollups and reporting. Deterministic and independently valuable.

**Validation gate** — the nine checks above, run against live data with the CI
owner before the workflow is enabled.

**Part 2 (~4 days)** — Stage C adjudicator, rule-PR emitter, enable and
measure. Does not begin until the gate passes.

#### Targets

Targets are set against the frozen Day-0 baseline, and deliberately lead with
attribution rather than with a green-rate number: the first honest win is
knowing who owns the red, not making it disappear.

| Horizon | Target |
|---|---|
| 30 days | <15% unclassified; every top-10 signature has an owning class |
| 60 days | infra + external share of failures halved |
| 90 days | scheduled green rate materially above zero; top-5 concentration falling |
