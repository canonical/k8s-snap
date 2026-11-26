<!--
Copyright 2026 Canonical, Ltd.
-->
# CI failure classification rules

The daily `ci-metrics` workflow fingerprints every failed CI job into a
*signature*, then classifies signatures with the ordered rule pack
`ci/failure_rules.yaml`. Only `product.bug` and `test.bug` signatures become
GitHub issues; everything else feeds the Mattermost digest. A wrong rule
therefore either files a bogus product bug or hides a real one, so this skill
is about adding rules that are backed by evidence, and about leaving a
failure `unknown` when the evidence is not there.

`ci/metrics/rules.py` is the authority on the rule schema and
`ci/metrics-data/README.md` on the pipeline; trust them over this file.

## Get the real data

The store lives on the `ci-metrics` branch as a SQL dump. Work from `ci/`:

```bash
git fetch origin ci-metrics
git show origin/ci-metrics:ci/metrics-data/metrics.sql > /tmp/metrics.sql
tox -e k8s-ci -- metrics rebuild --sql /tmp/metrics.sql --db /tmp/metrics.db
```

The unclassified signatures, biggest first:

```bash
sqlite3 /tmp/metrics.db "
  SELECT j.signature_id, COUNT(*) AS jobs, MIN(j.html_url)
  FROM v_job_class v JOIN jobs j USING (job_id)
  WHERE v.failure_class = 'unknown' AND j.signature_id IS NOT NULL
  GROUP BY 1 ORDER BY 2 DESC"
sqlite3 /tmp/metrics.db \
  "SELECT exemplar_excerpt FROM signatures WHERE signature_id = '<id>'"
```

`exemplar_excerpt` is exactly the text rules match against. Read the full
job log too (`gh api repos/canonical/k8s-snap/actions/jobs/<job_id>/logs`)
while GitHub still has it, usually 90 days: the excerpt is the extractor's
best guess, and the log is where you confirm it.

## Decide the class

Classify by **where the fix belongs**, not by which command printed the
error:

| Class | The fix belongs in | Examples |
|---|---|---|
| `infra.runner` | the runner machine | disk full, OOM, dpkg lock held |
| `infra.provisioning` | substrate setup (LXD, Multipass) | instance never launched |
| `external.dependency` | nothing of ours; retry | Snap Store, registries, PyPI, archive |
| `ci.config` | `.github/workflows` | missing `apt-get update`, missing artifact |
| `product.bug` | `k8s-snap`, `k8sd` or a component | service panics, node never Ready |
| `test.bug` | `tests/integration` | harness exception, timeout too short |

When the evidence does not settle the class, leave the signature `unknown`.
An unknown stays visible in the digest; a wrong class files a misleading
issue or hides a real one.

If the excerpt does not contain the cause but the log does, fix the extractor
rather than writing a rule against the symptom. A `CalledProcessError`
excerpt, for example, carries the failed command's stderr as a
`stderr: ...` line because that is the only place an unreachable Snap Store
differs from a broken refresh hook. Extractor changes go in
`ci/metrics/signature.py` with a bump of `EXTRACTOR_VERSION`.

## Write the rule

```yaml
  - id: external-snap-store          # tier-prefixed, stable, unique
    failure_class: external.dependency
    subclass: snap_store             # required; short snake_case noun
    confidence: 0.7                  # optional; below 1 when a human should confirm
    excerpt_pattern: 'unable to contact snap store'
    job:                             # optional metadata guards (regex)
      test_nodeid: '::test_feature_upgrades_inplace\b'
    examples:                        # required; real excerpts, trimmed
      - "stderr: error: unable to contact snap store"
      - excerpt: 'AssertionError: Failed to meet condition'
        job: {test_nodeid: tests/test_version_upgrades.py::test_feature_upgrades_inplace}
    notes: >-
      Why this text means this class, in a sentence or two.
```

- **Order is the semantics.** The first matching rule wins, and the pack is
  grouped by tier in the table's order, so an environment cause beats the
  product command it broke. A unit test enforces the tier order.
- **Match text, never IDs.** `signature_id` pins break on every extractor
  bump; the default pack must not use them.
- **Single-line patterns.** Patterns are case-insensitive and multiline (`^`
  and `$` match per line). The precision check tests line by line, so a
  pattern spanning lines is never checked. To scope a generic message to one
  test, add a `job.test_nodeid` guard instead. The one deliberate exception is
  a rule that must see a failed command *and* its `stderr:` line, such as
  `product-snap-install-failure`; say why in a comment.
- **No symptom rules.** `Failed to meet condition`, `exit status 1` or a bare
  `AssertionError` on their own say *that* something failed, not why. Retry
  noise in passing jobs prints the same text.
- **Examples are tests.** Loading the pack fails unless every example is
  classified by the rule that lists it, so a typo or a rule that shadows
  another breaks loading instead of moving failures between classes silently.
  Copy examples from real excerpts.
- **Bump `version`** at the top of the file whenever what the rules match
  changes. Reports use only the current version; earlier ones remain in the
  store as history.

## Prove it

From `ci/`:

```bash
tox -e lint,metrics                                   # schema, examples, tier order
tox -e k8s-ci -- metrics classify --db /tmp/metrics.db --no-dump --json
GITHUB_TOKEN=$(gh auth token) tox -e k8s-ci -- metrics check-rules \
  --db /tmp/metrics.db --repo canonical/k8s-snap --fail-on-imprecise
```

`classify` shows how many signatures each rule now wins. `check-rules`
samples successful runs of the tests a rule classifies and marks the rule
`imprecise` if its pattern hits more than 10% of them. An imprecise rule is
matching a symptom; narrow it or drop it. To see what the issue bridge would
file without posting anything, run
`metrics open-issues --db /tmp/metrics.db --dry-run --no-dump`.

## Automated proposals

Every Monday the `ci-metrics` workflow runs `metrics propose-rules`: a model
gets the biggest unclassified signatures, one at a time, with their excerpt,
job metadata, a scrubbed slice of the failed step's log, the current pack
and this guide, and replies with one rule or `"decision": "unknown"`. The
model never edits the pack. `ci/metrics/propose.py` accepts a rule only if:

- it is well formed: tier-prefixed id, no `signature_id`, a single-line regex
  of at most 200 characters without nested quantifiers;
- its examples are copied from the real evidence;
- the pack still loads with it in its tier, so no example anywhere is
  shadowed;
- it classifies the signature it was proposed for;
- it changes the class of nothing another rule already classifies;
- it hits at most 10% of successful runs of the same tests.

A rejection goes back to the model with the reason, once. Accepted rules are
inserted at the end of their tier, `version` is bumped, and a pull request
from `ci-metrics/rule-proposals` lists each rule with its evidence. Review
it like any rule change: the checks prove a rule is precise and harmless,
not that its class is right.

To run it locally against the real store (from `ci/`, after the steps in
"Get the real data"), with GitHub Models or OpenRouter credentials:

```bash
GITHUB_TOKEN=$(gh auth token) tox -e k8s-ci -- metrics propose-rules \
  --db /tmp/metrics.db --provider github --limit 5 --report /tmp/proposals.md
```

Add `--write` to update `ci/failure_rules.yaml` in place.

## Ship it

One pull request per coherent change. In the description, give for each rule
the signatures and job counts it now classifies, the `check-rules` verdict,
and one log link that shows the cause. Use Conventional Commits, for example
`feat: classify snap store outages`, and sign off every commit.
