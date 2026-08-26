# Triage Bot Security Remediation - Review Brief (round 3)

An autonomous GitHub issue triage+fix bot for `canonical/k8s-snap`, living in
`ci/triage_bot/` (pure Python label-FSM) driven by `ci/cmds/triage.py` (CLI) and
`.github/workflows/triage.yaml`. It classifies issues, and for a *trusted*
(maintainer-associated) event runs a reproduce->diagnose->verify->fix pipeline
whose shell agent builds the snap on a self-hosted LXD runner.

This is **round 3**. Rounds 1-2 already BLOCKED and were addressed. This round
verifies the round-2 fixes are actually airtight and looks for anything new.

## What to check
Read the actual files under `ci/triage_bot/` and `.github/workflows/triage.yaml`.
Offline tests pass (57) and lint is clean; do NOT rubber-stamp. Give each finding
a `file:line`, a severity (P0 blocker / P1 / P2 / nit), and a concrete fix.
Distinguish a real defect from a nit. End with BLOCK or APPROVE.

## Round-2 fixes to scrutinize (were these done correctly?)
1. **Shell secret leak (was P1)** `skills.py:53-115`. Agent shell now uses
   `bash -c` (not `-lc`), a secret-free `_ENV_PASSTHROUGH` allowlist, and
   `HOME` forced to a per-issue scratch dir. cwd is the checkout root via
   `repo_root()` (module-path derived, `parents[2]`). Can secrets still reach
   the reporter-driven shell? Is `repo_root()` correct in CI (checkout at repo
   root, tox runs from `ci/`)?
2. **Non-idempotent/ crashing `_open_fix_pr` (was P1)** `pipeline.py:100-128`,
   `github.py:115-147`. Now: reuse existing PR (no re-push), `push_branch`
   verifies the local branch exists (`git rev-parse --verify`) and returns False
   if not, force-push a bot-owned branch, and the whole thing is wrapped so a
   `GitHubError` degrades to no-PR. Any path that still crashes the pipeline or
   misreports a fix? Is `--force` on `triage/fix-<n>` acceptable?
3. **Dedup/missing-info re-fire on retriage (was P1)** `handlers/triage.py:32-87`,
   `handlers/retriage.py:52`. `handle_triage(recheck_gates=)` skips the two
   gates on a retriage re-entry. Can an untrusted retriage now reach the
   pipeline? (Trust gate is separate, below the gates.) Verify it can't.
4. **`closed` -> privileged self-hosted runner (was P2, DoS)**
   `.github/workflows/triage.yaml`. Gate now emits a 3-way `route`
   (pipeline/classify/cleanup); a close routes to a dedicated ubuntu `cleanup`
   job with only `contents:write` and no model key. Does any close still reach
   self-hosted? Does cleanup have enough perms (branch delete) and no more?
5. **Comment author-association inheritance (was P2)** `router.py:53-64`. A
   comment event now binds trust to `comment.author_association` only (issue
   events use the issue author's). Confirm an untrusted commenter on a
   maintainer's issue is untrusted. Any payload shape that defeats this?

## Perspective assignments
- **SecurityReviewer**: adversarial. You are a malicious issue reporter or
  commenter with NO repo association. Try to reach the self-hosted shell, leak
  `GH_TOKEN`/`GOOGLE_API_KEY`, or run the pipeline on your text. Attack the gate
  vs. bot-code trust seam, the shell env/HOME/cwd, and the workflow routing.
- **MaintainerReviewer**: correctness & maintainability. Does the FSM stay
  coherent (no wedged states, honest labels/comments), is `_open_fix_pr`
  genuinely idempotent, are the new tests meaningful (would they fail on a
  plausible regression), and is `repo_root()`/cwd/HOME sane for real CI?
