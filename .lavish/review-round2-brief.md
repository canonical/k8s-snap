# Triage Bot Security Remediation — Review Brief (round 2)

## What this is
An autonomous GitHub issue triage+fix bot for `canonical/k8s-snap`, living in `ci/triage_bot/`
(pure Python label-FSM) driven by `ci/cmds/triage.py` (CLI) and `.github/workflows/triage.yaml`.
It classifies issues, runs a reproduce→diagnose→verify→fix skill pipeline (LangGraph ReAct agent
with a **bash tool on a live LXD cluster**), and opens draft fix PRs.

Round 1 review (two reviewers) BLOCKED on a P0 trust model and several P1 integrity bugs.
This round asks you to verify the remediation is correct and complete, and to find anything new.

## The threat model that drove the change
The skill pipeline drives a shell-wielding LLM agent on a live cluster, seeded with
reporter-controlled issue/comment text. So: **untrusted input must never reach the shell**, and
**secrets must never sit in that shell's environment**.

## Remediation implemented (verify each)
1. **Code-authoritative trust gate.** `router.py` derives `trusted` from the triggering actor's
   `author_association` (`OWNER|MEMBER|COLLABORATOR`, frozenset `TRUSTED_ASSOCIATIONS`). Threaded onto
   `Triage`/`Retriage` actions and into `handle_triage(..., trusted=)`. Untrusted opened/reopened issues
   are classified + labelled + missing-info-checked, then **parked at `needs-triage`** without ever
   calling `rt.pipeline`. `VerifyFix` is gated in the router (a fix verdict requires a maintainer comment).
2. **Secret-stripped agent shell.** `skills.py` `_safe_env()` passes only an allowlist
   (`PATH,HOME,USER,LANG,LC_ALL,TERM,TZ`); `GH_TOKEN`/`GOOGLE_API_KEY`/any `*_TOKEN|*_KEY` never enter
   the agent's `subprocess`. Consequence: the agent can no longer push or use `gh`.
3. **Honest PR flow.** Because the shell has no creds, the trusted **orchestrator** (Python, holds token)
   pushes the branch (`github.py::push_branch`, `--force-with-lease`) then opens the PR;
   `TriageResult.pr_url` carries the real URL and the fix-pending comment only claims a PR when one exists.
4. **dry_run guard.** `github.py::create_pull_request` returns `None` under dry_run (was opening a REAL PR).
5. **Forgery-proof failure cap.** `count_failures(comments, bot_logins)` counts the hidden failure marker
   only on **bot-authored** comments (`Comment{body,author}`), so a reporter can't paste it to wedge the cap.
6. **Workflow split (defense-in-depth).** `triage.yaml`: cheap `gate` job (ubuntu-latest) computes
   `privileged` from `author_association`; trusted→self-hosted `pipeline` job (issues/contents/PR write,
   `--auto-pr`); untrusted→`ubuntu-latest` `classify` job (`contents: read`, no cluster). Code gate remains
   authoritative; workflow adds runner+permission isolation.
7. **No dead states.** Pipeline exceptions park at retriageable `triage/failed` with a failure comment
   (was: silent limbo). Duplicates park at retriageable `needs-triage` (was: terminal `not-actionable`).
8. **Cleanup.** Removed dead `PRContent`, `pr_skill_dir`, `build_command`, `Report.exists()`;
   renamed `logging.py`→`runlog.py` (stdlib shadow); manual `--issue` mode now fetches real labels and
   stamps `OWNER`.

## Verification already done
- 51 offline tests pass (incl. new: untrusted-parks-no-pipeline, forged-marker-ignored,
  pipeline-error→failed, real-PR-link, dry-run PR/push guards, router trust gating).
- black/isort/flake8/codespell clean on `ci/triage_bot` + `ci/cmds/triage.py`.
- Smoke: untrusted `opened` → classified + parked, pipeline never invoked, no branch pushed.

## Your job
Assess whether the remediation is correct, complete, and free of NEW defects. Be specific: cite
`file:line`, give severity (P0/P1/P2), and a concrete fix. Do NOT rubber-stamp; if a boundary is still
crossable or a state can still wedge, say so. Read the actual files under `ci/triage_bot/`.
