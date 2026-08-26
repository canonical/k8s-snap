#
# Copyright 2026 Canonical, Ltd.
#
"""``opened``/``reopened`` -> full triage handler."""

from __future__ import annotations

import logging
from typing import Optional

from .. import triage_core
from ..labels import current_triage_label
from ..pipeline import salvage_reproducer
from ..schema import TriageResult
from ..skills import repo_root
from .base import HandlerResult, IssueContext, Runtime, maintainer_ping, with_marker

_INSPECT = "sudo /snap/k8s/current/k8s/scripts/inspect.sh"

log = logging.getLogger("triage_bot")


def handle_triage(
    rt: Runtime,
    issue: IssueContext,
    *,
    trusted: bool = False,
    recheck_gates: bool = True,
) -> HandlerResult:
    labels = rt.ctx.labels
    actions: list[str] = []

    fields = triage_core.parse_template(issue.body)
    comment_bodies = [c.body for c in issue.comments]
    tarball = triage_core.has_tarball(issue.body, comment_bodies)
    classification = rt.classify(
        title=issue.title,
        fields=fields,
        tarball=tarball,
        model_spec=rt.ctx.triage_model,
    )
    kind_area = [*classification.kind_labels, *classification.area_labels]
    if kind_area:
        rt.gh.add_labels(issue.number, kind_area)
        actions.append("labels:" + ",".join(kind_area))

    current = current_triage_label(issue.labels, labels)

    # Skip dedup/missing-info gates on retriage re-entries.
    if recheck_gates:
        duplicate = _duplicate(rt, issue)
        if duplicate is not None:
            # Park at needs-triage so false positives stay reachable by comment.
            return _park(
                rt,
                issue,
                current,
                labels.needs_triage,
                f"This looks like a possible duplicate of #{duplicate}. "
                f"{maintainer_ping(rt)}close this issue as a duplicate if "
                "confirmed, or comment saying it isn't -- that retriages it "
                "automatically.",
                actions + [f"duplicate:{duplicate}"],
                "duplicate",
            )

        # Answer already-supported requests before the missing-info gate.
        answered = _already_supported(rt, issue, current, actions, kind_area)
        if answered is not None:
            return answered

        # Propose enhancement workarounds before the missing-info gate.
        if any(lbl == "kind/enhancement" for lbl in kind_area):
            proposal = _propose_enhancement(rt, issue, current, actions)
            if proposal is not None:
                return proposal

        if classification.missing_info:
            return _park(
                rt,
                issue,
                current,
                labels.needs_reproduction,
                _needs_info_comment(classification.missing_info),
                actions + ["needs_info:" + ",".join(classification.missing_info)],
                "needs_info",
            )

    # Self-hosted pipeline runs only for maintainer-triggered (trusted) events.
    if not trusted:
        return _park(
            rt,
            issue,
            current,
            labels.needs_triage,
            "Thanks for the report. It has been classified. "
            f"{maintainer_ping(rt)}comment on this issue to start automated "
            "reproduction.",
            actions + ["awaiting_maintainer"],
            "awaiting_maintainer",
        )

    if "kind/bug" not in kind_area:
        return _park(
            rt,
            issue,
            current,
            labels.needs_human,
            "Automated triage classified this as a non-bug issue. "
            "The automated reproduce and fix pipeline only runs for defects. "
            f"{maintainer_ping(rt)}please review and process manually.",
            actions + ["skip_pipeline_non_bug"],
            "needs_human",
        )
    try:
        result = rt.pipeline(rt, issue)
    except Exception as exc:
        # Park at failed state on internal error.
        salvaged = salvage_reproducer(rt, issue) if rt.ctx.auto_pr else None
        note = (
            f" The end-to-end test it had already written is pushed here: {salvaged}."
            if salvaged
            else ""
        )
        rt.gh.add_comment(
            issue.number,
            with_marker(
                f"Automated triage hit an internal error.{note} "
                f"{maintainer_ping(rt)}comment on this issue to retry.",
                failure=True,
            ),
        )
        actions.append(f"triage:error:{type(exc).__name__}")
        return HandlerResult("triage", "error", labels.failed, actions)

    new_label, comment, outcome = _resolve(result, rt)
    rt.gh.swap_label(issue.number, current, new_label)
    failure = new_label in (labels.unable_to_reproduce, labels.unable_to_fix)
    rt.gh.add_comment(issue.number, with_marker(comment, failure=failure))
    actions.append(f"triage:{outcome}")
    return HandlerResult("triage", outcome, new_label, actions)


def _duplicate(rt: Runtime, issue: IssueContext) -> Optional[int]:
    tokens = triage_core.dedup_query_tokens(issue.title)
    if not tokens:
        return None
    query = f"repo:{rt.gh.repo}+is:issue+is:open+" + "+".join(tokens[:6])
    try:
        items = rt.gh.search_issues(query)
    except Exception:
        return None
    candidates = [
        {"number": it.get("number"), "title": it.get("title") or ""}
        for it in items
        if it.get("number") is not None and it.get("number") != issue.number
    ]
    match = triage_core.find_duplicate(issue.title, candidates)
    return match["number"] if match else None


def _already_supported(
    rt: Runtime,
    issue: IssueContext,
    current: Optional[str],
    actions: list[str],
    kind_area: list[str],
) -> Optional[HandlerResult]:
    """Answer a report asking for something k8s-snap already does.

    Returns a parked result when the request is already satisfied, else None so
    triage carries on. A link to the documentation resolves these far faster
    than a cluster run; when nothing documents the feature the reporter still
    gets working commands, and the gap is labelled for a docs update rather
    than left implicit.
    """
    labels = rt.ctx.labels
    pages = triage_core.doc_inventory(repo_root())
    if not pages:
        return None
    try:
        support = rt.existing_support(
            title=issue.title,
            body=issue.body,
            pages=pages,
            model_spec=rt.ctx.triage_model,
            root=repo_root(),
        )
    except Exception:
        # Never let an advisory check block triage.
        log.exception("existing-support check failed for #%s", issue.number)
        return None
    if not support.already_supported or not support.explanation:
        return None

    taken = actions + ["already_supported"]
    body = [support.explanation]
    links = [triage_core.doc_url(p) for p in support.doc_paths]
    if support.instructions:
        body.append(f"How to do it today:\n\n```\n{support.instructions}\n```")

    if links:
        body.append("Documentation:\n" + "\n".join(f"- {url}" for url in links))
    else:
        body.append(
            "No documentation page covers this yet, so it is labelled "
            f"`{labels.docs_needed}`. {maintainer_ping(rt)}add a docs page "
            "covering it."
        )
        rt.gh.add_labels(issue.number, [labels.docs_needed])
        taken.append("docs_needed")
    if "kind/bug" in kind_area:
        body.append(
            "Since this is a supported feature but you are experiencing an issue, "
            "we need more information to investigate why it is failing in your environment. "
            "Please attach an inspection report (`sudo /snap/k8s/current/k8s/scripts/inspect.sh`) "
            "and any relevant specific component logs."
        )
        new_label = labels.needs_reproduction
        outcome = "needs_info"
    else:
        new_label = labels.needs_human
        outcome = "already_supported"

    return _park(
        rt,
        issue,
        current,
        new_label,
        "\n\n".join(body),
        taken,
        outcome,
    )


def _propose_enhancement(
    rt: Runtime, issue: IssueContext, current: Optional[str], actions: list[str]
) -> Optional[HandlerResult]:
    """Surface workarounds and implementation ideas for a feature request.

    Returns a parked result so the reporter gets something concrete rather than
    a generic "noted" message. Never raises: a failure is logged and triage
    continues as if this gate were absent.
    """
    labels = rt.ctx.labels
    pages = triage_core.doc_inventory(repo_root())
    try:
        proposal = rt.propose_enhancement(
            title=issue.title,
            body=issue.body,
            pages=pages,
            model_spec=rt.ctx.triage_model,
            root=repo_root(),
        )
    except Exception:
        log.exception("enhancement proposal failed for #%s", issue.number)
        return None
    if not proposal.ideas and not proposal.workaround_exists:
        return None

    taken = actions + ["enhancement_proposal"]

    # First part: for the reporter -- followable steps, not framing.
    reporter_parts: list[str] = ["This feature is not yet part of k8s-snap."]
    if proposal.workaround_exists and proposal.workaround_instructions:
        workaround = (
            "**What you can do today:**\n\n"
            f"```\n{proposal.workaround_instructions}\n```"
        )
        if proposal.workaround_doc_paths:
            links = "\n".join(
                f"- {triage_core.doc_url(p)}" for p in proposal.workaround_doc_paths
            )
            workaround += f"\n\nDocumentation:\n{links}"
        reporter_parts.append(workaround)

    # Second part: the proposal, at the end. The `cc` is the only signal this
    # is for maintainers -- no need to say so explicitly. Omitted entirely
    # (no dangling separator) when there are no ideas to present.
    team = rt.ctx.maintainer_team
    proposal_parts: list[str] = []
    if proposal.ideas:
        lines = [
            f"cc @{team}" if team else "",
            "**Possible implementation paths:**",
        ]
        for i, idea in enumerate(proposal.ideas, 1):
            block = (
                f"{i}. **{idea.title}** · effort: {idea.effort}\n\n"
                f"   {idea.description}"
            )
            if idea.example:
                indented = "\n".join(
                    f"   {ln}" if ln else "" for ln in idea.example.splitlines()
                )
                block += f"\n\n   ```\n{indented}\n   ```"
            lines.append(block)
        proposal_parts = ["---", "\n\n".join(part for part in lines if part)]

    comment = "\n\n".join([*reporter_parts, *proposal_parts])

    return _park(
        rt,
        issue,
        current,
        labels.needs_human,
        comment,
        taken,
        "enhancement_proposal",
    )


def _resolve(result: TriageResult, rt: Runtime) -> tuple[str, str, str]:
    """Map a pipeline result to (new triage label, comment body, outcome)."""
    labels = rt.ctx.labels
    if result.skipped:
        reason = result.skipped_reason or "not-actionable"
        if reason == "missing-details":
            return (
                labels.needs_reproduction,
                "Automated triage could not attempt to reproduce this issue due to missing details. Please attach "
                f"an inspection report (`{_INSPECT}`) and the exact commands, "
                "configuration files, and environment used. If you're not able to provide "
                "this, let us know in a comment and this will be flagged for "
                "manual review instead.",
                "needs_info",
            )
        return (
            labels.skipped,
            f"Automated triage skipped this issue ({reason}).",
            f"skipped:{reason}",
        )
    if not result.reproducible:
        return (
            labels.unable_to_reproduce,
            "Automated triage could not reproduce this issue. Please attach "
            f"an inspection report (`{_INSPECT}`) and the exact steps, "
            "versions, and environment used. If you're not able to provide "
            "this, let us know in a comment and this will be flagged for "
            "manual review instead.",
            "unable_to_reproduce",
        )
    if result.verdict == "intended-behavior":
        return (
            labels.needs_human,
            "Automated triage reproduced the described behaviour and found "
            "it matches the documented, intended behaviour. "
            f"{maintainer_ping(rt)}close this issue if that's correct, or "
            "reopen it with why the behaviour is actually wrong.",
            "intended_behavior",
        )
    if result.completed_stage == "reproducer":
        return (
            labels.unable_to_fix,
            "Automated triage reproduced this issue by hand but could not "
            "turn it into an end-to-end test that fails, so there is "
            "nothing to verify a fix against. No code was changed. "
            f"{maintainer_ping(rt)}please continue manually, or close this "
            "if it isn't worth pursuing.",
            "no_reproducer",
        )
    if result.fixed:
        if result.pr_url:
            body = (
                "Automated triage reproduced the issue and opened a draft "
                f"fix PR: {result.pr_url}. {maintainer_ping(rt)}review it, "
                "then merge to confirm the fix or close it without merging "
                "to reject -- a clear comment works too."
            )
        else:
            body = (
                "Automated triage reproduced the issue and prepared a "
                "candidate fix, but no PR was opened automatically. "
                f"{maintainer_ping(rt)}re-run with auto-PR enabled, or open "
                "one from the bot's branch manually."
            )
        return (labels.fix_pending, body, "fix_pending")
    if result.verification_blocked:
        reason = f" ({result.blocked_reason})" if result.blocked_reason else ""
        if result.pr_url:
            body = (
                "Automated triage diagnosed the issue and committed a "
                f"candidate fix, but could not rebuild and re-run the test "
                f"to confirm it{reason}: {result.pr_url}. "
                f"{maintainer_ping(rt)}review the change and verify it "
                "locally, then merge to confirm the fix or close it "
                "without merging to reject."
            )
        else:
            body = (
                "Automated triage diagnosed the issue and prepared a "
                f"candidate fix, but could not verify it{reason}, and no "
                f"PR was opened automatically. {maintainer_ping(rt)}re-run "
                "with auto-PR enabled, or open one from the bot's branch "
                "manually."
            )
        return (labels.fix_pending, body, "fix_pending_unverified")
    if result.pr_url:
        body = (
            "Automated triage reproduced the issue and pushed the "
            f"end-to-end test that captures it: {result.pr_url}. That test "
            "fails against `main` by design. A confident fix could not be "
            f"prepared. {maintainer_ping(rt)}please continue from that "
            "branch, or close the PR if this isn't worth pursuing."
        )
    else:
        body = (
            "Automated triage reproduced the issue but could not prepare a "
            f"confident fix. {maintainer_ping(rt)}please continue manually, "
            "or close this if it isn't worth pursuing."
        )
    return (labels.unable_to_fix, body, "unable_to_fix")


def _needs_info_comment(missing: list[str]) -> str:
    bullets = "\n".join(f"- {m}" for m in missing)
    return (
        "Thanks for the report. To help us reproduce and triage this issue, "
        "could you please provide:\n\n"
        f"{bullets}\n\n"
        f"For the inspection tarball, run `{_INSPECT}` and attach the "
        "generated file. If you're not able to provide this, let us know "
        "in a comment and this will be flagged for manual review instead."
    )


def _park(
    rt: Runtime,
    issue: IssueContext,
    current: Optional[str],
    new_label: str,
    comment: str,
    actions: list[str],
    outcome: str,
) -> HandlerResult:
    """Swap to a resting label and post a single comment."""
    rt.gh.swap_label(issue.number, current, new_label)
    rt.gh.add_comment(issue.number, with_marker(comment))
    return HandlerResult("triage", outcome, new_label, actions)
