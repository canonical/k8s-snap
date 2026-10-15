# CI metrics data

This directory is the canonical-facts half of the CI failure classification
pipeline described in
[`docs/proposals/003-ci-failure-classification-and-metrics.md`](https://github.com/canonical/k8s-snap/blob/main/docs/proposals/003-ci-failure-classification-and-metrics.md)
and
[`docs/dev/ci-failure-classification-delivery-plan.md`](https://github.com/canonical/k8s-snap/blob/main/docs/dev/ci-failure-classification-delivery-plan.md).
The pipeline itself lives under `ci/metrics/` and `ci/cmds/metrics.py`, reachable
via `python ci/k8s-ci.py metrics --help`.

## The `ci-metrics` branch

The committed state below is read from and written to a dedicated `ci-metrics`
branch, not `main`. `.github/workflows/ci-metrics.yaml` checks out `main` for
code, then layers this directory's state from `ci-metrics` on top before
running — so the pipeline's code changes with `main` (via normal PR review),
but its durable facts accumulate on `ci-metrics` independently of any single
code PR. The workflow pushes back to `ci-metrics` by compare-and-swap, never
`--force`; a rejected push means a concurrent run won and this run retries.

- **`metrics.sql`** — a deterministic `sqlite3 .dump` of the durable fact
  tables (runs, jobs, signatures, classifications, issue bridge state).
  `metrics.db` is rebuilt from this file at the start of every run
  (`metrics rebuild`) and is gitignored — the dump, not the binary database,
  is the thing under review and `git diff`/blame.
- **`reports/latest.md`** and **`signatures/catalogue.md`** — generated
  Markdown exports for humans, regenerated every run (`metrics report`) and
  force-added to the commit despite being gitignored on `main` (the ignore
  rule exists so local `tox -e metrics` runs don't dirty a `main` checkout;
  it does not apply to the facts this branch is meant to carry).

Nothing here is edited by hand. To inspect or rebuild locally:

```sh
git fetch origin ci-metrics
git show origin/ci-metrics:ci/metrics-data/metrics.sql > ci/metrics-data/metrics.sql
python ci/k8s-ci.py metrics rebuild
python ci/k8s-ci.py metrics rollup | less
```

## What runs when

`.github/workflows/ci-metrics.yaml` has four triggers:

| Trigger | What it does |
|---|---|
| `workflow_run` completion of Nightly/Weekly/Lint and integration | Ingests just the run that completed, classifies, regenerates reports, dry-runs (or posts, if `post_issues` was set) the issue bridge. Lowest latency. |
| Hourly `schedule` | Re-ingests a short backfill window for the two scheduled workflows only, to self-heal any `workflow_run` delivery GitHub dropped. Ingest is an upsert keyed on `(run_id, attempt)`, so overlap with the per-run trigger is harmless. |
| Daily `schedule` | Everything the hourly run does, plus a Mattermost digest (`metrics digest`) leading with attribution and delta, per the delivery plan. |
| `workflow_dispatch` | Manual trigger. `post_issues: true` posts real issues instead of dry-running; an optional `backfill_since` input (e.g. `90d`) re-ingests a historical window, for the one-time Day-0 baseline the delivery plan describes. |

The issue bridge itself (`metrics open-issues`) stays `--dry-run` by default
on every trigger; only `workflow_dispatch` with `post_issues: true` can create
or update real GitHub issues, and even then it is gated on the recurrence
threshold and daily budget described in the delivery plan.
