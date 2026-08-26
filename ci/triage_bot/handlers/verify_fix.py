#
# Copyright 2026 Canonical, Ltd.
#
"""``created`` on a fix-pending issue or PR close -> verify the fix."""

from __future__ import annotations

from typing import Literal, Optional

from ..labels import current_triage_label
from ..schema import FixVerification
from .base import HandlerResult, IssueContext, Runtime, maintainer_ping, with_marker


def handle_verify_fix(
    rt: Runtime,
    issue: IssueContext,
    *,
    comment_body: str = "",
    verdict: Optional[Literal["confirmed", "rejected"]] = None,
) -> HandlerResult:
    labels = rt.ctx.labels
    current = current_triage_label(issue.labels, labels)
    if current != labels.fix_pending:
        # Ensure issue is still at fix-pending before applying verdict.
        return HandlerResult("verify_fix", "not-fix-pending", current, [])

    if verdict is not None:
        verification = FixVerification(status=verdict)
    else:
        # Classify the trust-checked triggering comment.
        latest = comment_body
        report = rt.report(issue).read()
        verification = rt.verify_fix(
            latest_comment=latest,
            report=report,
            model_spec=rt.ctx.verification_model,
        )

    if verification.status == "confirmed":
        rt.gh.swap_label(issue.number, labels.fix_pending, labels.fix_verified)
        pr = _find_fix_pr(rt, issue.number, state="all" if verdict else "open")
        _tag_fix_pr(rt, pr)
        merge_note = ""
        if verdict is None:
            merge_note = (
                f" Merge {pr['html_url']} to complete this."
                if pr and pr.get("html_url")
                else " Merge the draft fix PR to complete this."
            )
        rt.gh.add_comment(
            issue.number,
            with_marker(f"Fix confirmed. Thanks for verifying.{merge_note}"),
        )
        return HandlerResult(
            "verify_fix", "confirmed", labels.fix_verified, ["fix:confirmed"]
        )

    if verification.status == "rejected":
        rt.gh.swap_label(issue.number, labels.fix_pending, labels.fix_rejected)
        pr = _find_fix_pr(rt, issue.number, state="all" if verdict else "open")
        close_note = ""
        if verdict is None:
            close_note = (
                f" Please close {pr['html_url']} without merging."
                if pr and pr.get("html_url")
                else " Please close the draft fix PR without merging."
            )
        rt.gh.add_comment(
            issue.number,
            with_marker(
                f"Thanks, marking the proposed fix as rejected.{close_note} "
                f"{maintainer_ping(rt)}further details on what still fails "
                "are welcome.",
                failure=True,
            ),
        )
        return HandlerResult(
            "verify_fix", "rejected", labels.fix_rejected, ["fix:rejected"]
        )

    # Avoid nagging maintainers discussing the fix in-thread repeatedly.
    already_nagged = any(
        "didn't clearly confirm or reject the fix" in c.body for c in issue.comments
    )
    if not already_nagged:
        rt.gh.add_comment(
            issue.number,
            with_marker(
                "That comment didn't clearly confirm or reject the fix. "
                f"{maintainer_ping(rt)}reply with a clear confirmation or "
                "rejection, or merge the draft PR to confirm it / close the PR "
                "without merging to reject it."
            ),
        )
    return HandlerResult(
        "verify_fix", "inconclusive", labels.fix_pending, ["fix:inconclusive"]
    )


def _find_fix_pr(rt: Runtime, number: int, state: str = "open") -> Optional[dict]:
    """Locate the bot's draft PR for this issue."""
    branch = rt.gh.find_branch([f"triage/fix-{number}", f"fix/issue-{number}"])
    if not branch:
        return None
    return rt.gh.find_pull_request(branch, state=state)


def _tag_fix_pr(rt: Runtime, pr: Optional[dict]) -> None:
    """Tag the open fix PR so the CI side can key off the confirmation."""
    if pr and pr.get("number") is not None:
        rt.gh.add_labels(pr["number"], [rt.ctx.labels.pr_fix_verified])
