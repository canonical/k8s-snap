#
# Copyright 2026 Canonical, Ltd.
#
"""``opened``/``reopened`` -> full triage.

The triage handler is the FSM entry transition. It always applies the
deterministic ``kind/``/``area/`` labels, then runs three cheap gates before
spending a cluster run:

1. **duplicate** -- a strong title match to another open issue is flagged for
   a maintainer and parked at ``needs-triage`` (retriageable via a follow-up
   comment, never terminal, since a false positive must stay reachable).
2. **already supported** -- a request for behaviour that already ships (or a
   bug already fixed) is answered with links to the documentation, or, when
   nothing documents it, with working commands plus a ``docs-change-needed``
   label. Runs before the missing-details gate, since a feature request has no
   reproduction steps to be missing.
3. **missing details** -- a bug with no inspection tarball / no reproduction
   steps is sent back to the reporter and parked at ``needs-reproduction``.

Only when all three gates pass does it invoke the reproduce -> verify ->
reproducer -> diagnose -> fix skill pipeline and map the :class:`TriageResult`
onto the terminal (or fix-pending) triage label.
"""

from __future__ import annotations

import logging
from typing import Optional

from .. import triage_core
from ..labels import current_triage_label
from ..pipeline import salvage_reproducer
from ..schema import TriageResult
from ..skills import repo_root
from .base import HandlerResult, IssueContext, Runtime, with_marker

_INSPECT = "sudo /snap/k8s/current/k8s/scripts/inspect.sh"

# Only the exception *type* reaches the issue (a message could echo attacker
# text or credentials); the detail an operator needs goes to the job log.
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

    # The cheap dedup/missing-info gates run on first triage only. On a retriage
    # re-entry a human already judged the issue worth another look, so re-firing
    # them would re-park (and re-comment) forever, wedging the pipeline behind a
    # sticky false-positive duplicate or an unchanged missing-info verdict.
    if recheck_gates:
        duplicate = _duplicate(rt, issue)
        if duplicate is not None:
            # Park at needs-triage (retriageable), never terminal not-actionable:
            # a false positive must remain reachable by a follow-up comment.
            return _park(
                rt,
                issue,
                current,
                labels.needs_triage,
                f"This looks like a possible duplicate of #{duplicate}. "
                "A maintainer will confirm and close it if so; comment if not.",
                actions + [f"duplicate:{duplicate}"],
                "duplicate",
            )

        # Before asking for reproduction details, check the report is not
        # asking for something that already ships (or a bug already fixed).
        # This runs ahead of the missing-info gate on purpose: a feature
        # request has no reproduction steps, so that gate would bounce it back
        # to the reporter instead of just answering the question.
        answered = _already_supported(rt, issue, current, actions)
        if answered is not None:
            return answered

        # For enhancement requests, propose workarounds and implementation
        # ideas before the missing-info gate -- a feature request has nothing
        # to reproduce, so the pipeline would just return not-actionable
        # without giving the reporter anything useful.
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

    # The reproduce->verify->reproducer->diagnose->fix pipeline drives a
    # shell-wielding agent on a live cluster, so reporter-controlled text
    # must never reach it.
    # It runs only for a maintainer-triggered (trusted) event; for anyone else
    # the issue rests, classified, at needs-triage until a maintainer starts it.
    if not trusted:
        return _park(
            rt,
            issue,
            current,
            labels.needs_triage,
            "Thanks for the report. It has been classified and is ready for a "
            "maintainer to start automated reproduction.",
            actions + ["awaiting_maintainer"],
            "awaiting_maintainer",
        )

    try:
        result = rt.pipeline(rt, issue)
    except Exception as exc:
        log.exception("triage pipeline failed for #%s", issue.number)
        # A cluster/skill failure must not leave the issue in limbo with no
        # transition or feedback. Park at the retriageable failed state.
        rt.gh.swap_label(issue.number, current, labels.failed)
        # The crash may have landed hours after the reproducer test was
        # committed. Publish that branch rather than throw the run away.
        salvaged = salvage_reproducer(rt, issue) if rt.ctx.auto_pr else None
        note = (
            f" The end-to-end test it had already written is pushed here: {salvaged}."
            if salvaged
            else ""
        )
        rt.gh.add_comment(
            issue.number,
            with_marker(
                "Automated triage hit an internal error and will retry on the "
                f"next maintainer comment.{note}",
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
        {"number": it.get("number"), "title": it.get("title", "")}
        for it in items
        if it.get("number") is not None and it.get("number") != issue.number
    ]
    match = triage_core.find_duplicate(issue.title, candidates)
    return match["number"] if match else None


def _already_supported(
    rt: Runtime, issue: IssueContext, current: Optional[str], actions: list[str]
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
    if links:
        body.append("Documentation:\n" + "\n".join(f"- {url}" for url in links))
    else:
        if support.instructions:
            body.append(f"How to do it today:\n\n```\n{support.instructions}\n```")
        body.append(
            "No documentation page covers this yet, so it is labelled "
            f"`{labels.docs_needed}` for a docs update."
        )
        rt.gh.add_labels(issue.number, [labels.docs_needed])
        taken.append("docs_needed")
    return _park(
        rt,
        issue,
        current,
        labels.not_actionable,
        "\n\n".join(body),
        taken,
        "already_supported",
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
        workaround = "**What you can do today:**\n\n" + proposal.workaround_instructions
        if proposal.workaround_doc_paths:
            links = "\n".join(
                f"- {triage_core.doc_url(p)}" for p in proposal.workaround_doc_paths
            )
            workaround += f"\n\nDocumentation:\n{links}"
        reporter_parts.append(workaround)

    # Second part: the proposal, at the end. The `cc` is the only signal this
    # is for maintainers -- no need to say so explicitly.
    team = rt.ctx.maintainer_team
    proposal_parts: list[str] = ["---"]
    if proposal.ideas:
        lines = [
            f"cc @{team}" if team else "",
            "**Possible implementation paths:**",
        ]
        for i, idea in enumerate(proposal.ideas, 1):
            block = f"{i}. **{idea.title}** · effort: {idea.effort}\n\n   {idea.description}"
            if idea.example:
                block += f"\n\n   ```\n   {idea.example}\n   ```"
            lines.append(block)
        proposal_parts.append("\n\n".join(part for part in lines if part))

    comment = "\n\n".join(reporter_parts) + "\n\n" + "\n\n".join(proposal_parts)

    return _park(
        rt,
        issue,
        current,
        labels.not_actionable,
        comment,
        taken,
        "enhancement_proposal",
    )


def _resolve(result: TriageResult, rt: Runtime) -> tuple[str, str, str]:
    """Map a pipeline result to (new triage label, comment body, outcome)."""
    labels = rt.ctx.labels
    if result.skipped:
        reason = result.skipped_reason or "not-actionable"
        return (
            labels.skipped,
            f"Automated triage skipped this issue ({reason}).",
            f"skipped:{reason}",
        )
    if not result.reproducible:
        return (
            labels.unable_to_reproduce,
            "Automated triage could not reproduce this issue. Please attach an "
            f"inspection report (`{_INSPECT}`) and the exact steps, versions, "
            "and environment used.",
            "unable_to_reproduce",
        )
    if result.verdict == "intended-behavior":
        return (
            labels.not_actionable,
            "Automated triage reproduced the described behaviour and found it "
            "matches the documented, intended behaviour. Marking not-actionable; "
            "a maintainer will review and close it, or reopen the discussion if "
            "this is wrong.",
            "intended_behavior",
        )
    if result.completed_stage == "reproducer":
        return (
            labels.unable_to_fix,
            "Automated triage reproduced this issue by hand but could not turn "
            "it into an end-to-end test that fails, so there is nothing to "
            "verify a fix against. No code was changed. A maintainer will need "
            "to take it from here.",
            "no_reproducer",
        )
    if result.fixed:
        if result.pr_url:
            body = (
                "Automated triage reproduced the issue and opened a draft fix "
                f"PR: {result.pr_url}. A maintainer will review, then comment "
                "to confirm or reject once verified."
            )
        else:
            body = (
                "Automated triage reproduced the issue and prepared a candidate "
                "fix. A maintainer will review and open a PR."
            )
        return (labels.fix_pending, body, "fix_pending")
    if result.pr_url:
        body = (
            "Automated triage reproduced the issue and pushed the end-to-end "
            f"test that captures it: {result.pr_url}. That test fails against "
            "`main` by design. A confident fix could not be prepared, so a "
            "maintainer will need to take it from here."
        )
    else:
        body = (
            "Automated triage reproduced the issue but could not prepare a "
            "confident fix. A maintainer will need to take it from here."
        )
    return (labels.unable_to_fix, body, "unable_to_fix")


def _needs_info_comment(missing: list[str]) -> str:
    bullets = "\n".join(f"- {m}" for m in missing)
    return (
        "Thanks for the report. To help us reproduce and triage this issue, "
        "could you please provide:\n\n"
        f"{bullets}\n\n"
        f"For the inspection tarball, run `{_INSPECT}` and attach the "
        "generated file."
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
