#
# Copyright 2026 Canonical, Ltd.
#
"""The reproduce -> verify -> reproducer -> fix skill pipeline."""

from __future__ import annotations

import logging
import subprocess
from pathlib import Path
from typing import Optional

from .github import GitHubError
from .report import Report
from .schema import (
    DiagnoseResult,
    FixResult,
    ReproduceResult,
    ReproducerResult,
    TriageResult,
    VerifyResult,
)
from .skills import (
    DEFAULT_CLUSTER_PREFIX,
    commit_sha,
    ensure_worktree,
    remove_worktree,
    render_report_json,
    repo_root,
    run_skill,
)

log = logging.getLogger("triage_bot")

# One line per stage, not per tool call: this is the narrative a human
# watching the log wants ("reproduce -> verify -> reproducer -> ..."), while
# the per-tool-call trace inside each stage logs at DEBUG (see runlog.py).
_SUMMARY = {
    "reproduce": lambda r: (
        f"skipped ({r.skipped_reason})"
        if r.skipped
        else f"reproducible={r.reproducible}"
    ),
    "verify": lambda r: f"verdict={r.verdict} confidence={r.confidence}",
    "reproducer": lambda r: (
        f"test={r.test_path or '(none)'} fails_before_fix={r.fails_before_fix}"
    ),
    "diagnose": lambda r: f"confidence={r.confidence}",
    "fix": lambda r: f"fixed={r.fixed}",
}


def _branch(issue) -> str:
    return f"triage/fix-{issue.number}"


def run_pipeline(rt, issue) -> TriageResult:
    """Run the full skill pipeline for an issue and aggregate the result."""
    log.info("[pipeline] issue #%s: starting", issue.number)
    workdir = rt.workdir(issue.number)
    # Every skill runs here, never in the primary checkout, and the branch is
    # already checked out so no skill needs a `git checkout -b`.
    checkout = ensure_worktree(rt.worktree(issue.number), _branch(issue))
    report = Report(workdir).start(issue.number, issue.title, issue.body)
    skill_dir = rt.ctx.triage_skill_dir
    model = rt.ctx.triage_model
    # Scopes this run's hack/cluster-up.sh calls (agent + cleanup) so a
    # concurrent run on another issue can never collide with or destroy
    # this one's cluster.
    cluster_prefix = f"{DEFAULT_CLUSTER_PREFIX}-{issue.number}"
    # Independent ground truth for "did a skill actually commit what it
    # claims", captured before any skill runs -- see the checks below.
    base_sha = commit_sha(checkout)

    def _run(step, model_cls, instructions):
        log.info("[%s] running", step)
        result = run_skill(
            skill_dir=skill_dir,
            step=step,
            instructions=instructions,
            result_model=model_cls,
            workdir=workdir,
            cwd=checkout,
            model_spec=model,
            extra_context=report.read(),
            run_id=f"issue-{issue.number}-{step}",
            jsonl_path=rt.ctx.jsonl_path,
            cluster_prefix=cluster_prefix,
        )
        report.append(step, render_report_json(result))
        log.info("[%s] %s", step, _SUMMARY[step](result))
        return result

    try:
        reproduce: ReproduceResult = _run(
            "reproduce",
            ReproduceResult,
            f"Reproduce issue #{issue.number}: {issue.title}",
        )
        if reproduce.skipped:
            log.info("[pipeline] issue #%s: stopping (skipped)", issue.number)
            return TriageResult(
                completed_stage="reproduce",
                skipped=True,
                skipped_reason=reproduce.skipped_reason,
            )
        if not reproduce.reproducible:
            log.info("[pipeline] issue #%s: stopping (not reproducible)", issue.number)
            return TriageResult(completed_stage="reproduce", reproducible=False)

        verify: VerifyResult = _run(
            "verify",
            VerifyResult,
            "Decide whether the reproduced behaviour is a real bug or intended.",
        )
        if verify.verdict == "intended-behavior":
            log.info(
                "[pipeline] issue #%s: stopping (intended behaviour)", issue.number
            )
            return TriageResult(
                completed_stage="verify",
                reproducible=True,
                verdict="intended-behavior",
            )

        reproducer: ReproducerResult = _run(
            "reproducer",
            ReproducerResult,
            "Write an end-to-end test that captures this bug and prove it fails now.",
        )
        reproducer_sha = commit_sha(checkout)
        # Trust, but verify: a skill can honestly believe it succeeded and
        # still not have actually committed (see reproducer.md's own note on
        # this). A commit either landed on the branch or it didn't --
        # checking that costs one `git rev-parse` and catches the mismatch
        # regardless of why the self-report drifted from reality.
        if reproducer.fails_before_fix and reproducer_sha == base_sha:
            log.warning(
                "[pipeline] issue #%s: reproducer reported fails_before_fix "
                "but no commit landed on %s -- treating as no reproducer",
                issue.number,
                _branch(issue),
            )
            reproducer = ReproducerResult(
                test_path=reproducer.test_path, fails_before_fix=False
            )
        # A test that does not fail today cannot demonstrate a fix tomorrow, so
        # stop rather than edit code against an unproven premise.
        if not (reproducer.test_path and reproducer.fails_before_fix):
            log.info(
                "[pipeline] issue #%s: stopping (no failing test produced)",
                issue.number,
            )
            return TriageResult(
                completed_stage="reproducer",
                reproducible=True,
                verdict=verify.verdict,
            )

        # Diagnosis and the fix come last, anchored to an executable failure.
        diagnose: DiagnoseResult = _run(
            "diagnose",
            DiagnoseResult,
            "Diagnose the root cause behind the failing test.",
        )
        fix: FixResult = _run(
            "fix",
            FixResult,
            "Make the failing end-to-end test pass with a minimal fix.",
        )
        # Same ground-truth check as the reproducer above: fixed or
        # verification_blocked both claim a commit exists beyond the
        # reproducer's own; if the branch didn't move, neither claim holds.
        if (fix.fixed or fix.verification_blocked) and commit_sha(
            checkout
        ) == reproducer_sha:
            log.warning(
                "[pipeline] issue #%s: fix reported fixed=%s "
                "verification_blocked=%s but no additional commit landed on "
                "%s -- treating as no fix",
                issue.number,
                fix.fixed,
                fix.verification_blocked,
                _branch(issue),
            )
            fix = FixResult()

        # Opened even when the fix failed: a proven-red test is worth landing.
        pr_url: Optional[str] = None
        if rt.ctx.auto_pr:
            pr_url = _open_pr(rt, issue, fix, reproducer)

        log.info(
            "[pipeline] issue #%s: done (fixed=%s, verification_blocked=%s, pr=%s)",
            issue.number,
            fix.fixed,
            fix.verification_blocked,
            pr_url or "none",
        )
        return TriageResult(
            completed_stage="fix",
            reproducible=True,
            verdict=verify.verdict,
            diagnosis_confidence=diagnose.confidence,
            fixed=fix.fixed,
            commit_message=fix.commit_message,
            pr_url=pr_url,
            test_path=reproducer.test_path,
            verification_blocked=fix.verification_blocked,
            blocked_reason=fix.blocked_reason,
        )
    finally:
        _cleanup(checkout, cluster_prefix)


def _cleanup(checkout: Path, cluster_prefix: str) -> None:
    """Destroy triage cluster and remove the temporary worktree."""
    log.info("[cleanup] destroying cluster %s", cluster_prefix)
    primary_lines = subprocess.run(
        ["git", "-C", str(checkout), "worktree", "list", "--porcelain"],
        capture_output=True,
        text=True,
    ).stdout.splitlines()
    primary_path = next(
        (
            ln.removeprefix("worktree ")
            for ln in primary_lines
            if ln.startswith("worktree ")
        ),
        str(repo_root()),
    )
    script = Path(primary_path) / "hack" / "cluster-up.sh"
    if not script.exists():
        log.warning("[cleanup] cluster-up.sh not found at %s, skipping", script)
        remove_worktree(checkout)
        return
    try:
        result = subprocess.run(
            ["bash", str(script), "--prefix", cluster_prefix, "--destroy"],
            capture_output=True,
            text=True,
            timeout=600,
        )
        if result.returncode == 0:
            log.info("[cleanup] done")
        else:
            log.warning(
                "[cleanup] cluster-up.sh --prefix %s --destroy exited %s: %s",
                cluster_prefix,
                result.returncode,
                result.stderr.strip()[-500:],
            )
    except Exception as exc:
        log.warning("[cleanup] failed (non-fatal): %s", exc)
    remove_worktree(checkout)


def _touches_component_pins(worktree: Path) -> bool:
    """Check if the branch modifies component repository or version pins."""
    try:
        res = subprocess.run(
            ["git", "-C", str(worktree), "diff", "--name-only", "main...HEAD"],
            capture_output=True,
            text=True,
        )
        for path in res.stdout.splitlines():
            parts = Path(path).parts
            if (
                len(parts) >= 4
                and parts[0] == "build-scripts"
                and parts[1] == "components"
                and parts[3] in ("repository", "version")
            ):
                return True
    except Exception as exc:
        log.warning("[pipeline] diff check against main failed: %s", exc)
    return False


def _test_reference(reproducer: ReproducerResult) -> str:
    """Return selector or path reference string."""
    ref = reproducer.test_selector or reproducer.test_path
    return f" (`{ref}`)" if ref else ""


def _open_pr(rt, issue, fix: FixResult, reproducer: ReproducerResult) -> Optional[str]:
    """Push the agent's local branch and open a draft PR, idempotently.

    The skills commit ``triage/fix-<n>`` in the checkout but cannot push (their
    shell holds no credentials); pushing and PR creation happen here in the
    trusted orchestrator. The branch is force-pushed first so a re-run updates
    an already-open PR to the new commit (never leaves it pointing at a stale,
    already-rejected diff); then an existing PR is reused, else a draft is
    opened. A branch the skills never committed degrades to no-PR, and any
    GitHub error degrades to no-PR -- a PR hiccup must never misreport a real
    fix as an internal error. Returns the PR URL, or None if none was opened.

    A failed fix still opens a PR, carrying the failing test alone: it is a
    reproducer a maintainer can run, and it is expected to fail CI until the
    bug is fixed, so it never claims to close the issue. A fix that was
    written but never rebuild-verified (``verification_blocked``) also opens
    a PR, carrying both the test and the unverified candidate -- distinct
    wording throughout makes clear it is not confirmed working.
    """
    branch = _branch(issue)
    worktree_dir = rt.worktree(issue.number)
    if _touches_component_pins(worktree_dir):
        log.warning(
            "[pipeline] issue #%s: refusing to open PR: commit modifies component repository/version pin",
            issue.number,
        )
        return None
    try:
        if not rt.gh.push_branch(branch, str(worktree_dir)):
            return None
        existing = rt.gh.find_pull_request(branch)
        if existing is not None:
            return existing.get("html_url")
        if fix.fixed:
            title = fix.commit_message or f"fix: resolve issue #{issue.number}"
            body = (
                f"Automated draft fix for #{issue.number}, with the end-to-end "
                f"test that reproduces it{_test_reference(reproducer)}.\n\n"
                "Prepared by the triage bot; awaiting maintainer verification.\n\n"
                f"Closes #{issue.number}"
            )
        elif fix.verification_blocked:
            title = (
                fix.commit_message
                or f"wip: candidate fix for #{issue.number} (unverified)"
            )
            reason = f" ({fix.blocked_reason})" if fix.blocked_reason else ""
            body = (
                f"Draft fix for #{issue.number}, with the end-to-end test "
                f"that reproduces it{_test_reference(reproducer)}.\n\n"
                "**Not verified**: the triage bot diagnosed a root cause and "
                f"committed a candidate fix, but could not rebuild and "
                f"re-run the test to confirm it{reason}. Review the change "
                "and confirm locally before merging.\n\n"
                f"Refs #{issue.number}"
            )
        elif reproducer.fails_before_fix:
            title = f"test: add failing reproducer for #{issue.number}"
            body = (
                f"End-to-end test reproducing #{issue.number}"
                f"{_test_reference(reproducer)}, observed to fail against "
                "current `main`.\n\nThe triage bot could not prepare a "
                "confident fix, so only the test is proposed here. **CI is "
                "expected to fail until the underlying bug is fixed.**\n\n"
                f"Refs #{issue.number}"
            )
        else:
            # Crash salvage (salvage_reproducer): the run stopped before the
            # reproducer stage confirmed anything. Publish what was
            # committed without claiming a fact nobody verified.
            title = f"wip: salvaged triage branch for #{issue.number}"
            body = (
                f"The triage bot's run for #{issue.number} did not finish. "
                "This branch carries whatever it had already committed "
                "before stopping, salvaged rather than discarded. It has "
                "not been verified to reproduce the issue or fail as "
                "expected -- review before relying on it.\n\n"
                f"Refs #{issue.number}"
            )
        pr = rt.gh.create_pull_request(head=branch, base="main", title=title, body=body)
        return (pr or {}).get("html_url")
    except GitHubError:
        return None


def salvage_reproducer(rt, issue) -> Optional[str]:
    """Push whatever the crashed run already committed, if anything.

    A run can die between committing the reproducer test and finishing the fix
    (an LLM timeout mid-fix is the common case, hours in). The test is the
    expensive part and it is already on the branch, so publish it rather than
    discard the whole run: the next attempt reuses the PR. Degrades to None
    when there is no branch, which is the normal case for an early crash.
    """
    return _open_pr(rt, issue, FixResult(fixed=False), ReproducerResult())
