#
# Copyright 2026 Canonical, Ltd.
#
"""``created`` on a fix-pending issue, or its own PR closing -> verify the fix.

A fix-pending issue's verdict comes from whichever happens first: a
maintainer comment (classified confirmed / rejected / inconclusive by the
injected ``verify_fix`` seam), or the bot's own draft PR being merged or
closed (ground truth, no classification needed -- see
``router.GitHubEvent.pr_verdict``). Either way the handler swaps to the
matching terminal (or re-triageable) label, and on confirmation also tags
the open fix PR so the CI side can pick it up.
"""

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
        # Defence in depth: the comment path is already gated by route()
        # (a verdict is only routed for an issue at fix-pending), but the
        # PR-driven path (GitHubEvent.pr_verdict) has no such issue-state
        # gate upstream -- a pull_request webhook carries no issue labels
        # at all, only the closing PR's own metadata. Re-checking here,
        # against the issue's real current label, is what actually stops
        # an unrelated PR that merely happens to be named triage/fix-<n>
        # (or fix/issue-<n>) from mutating that issue's state.
        return HandlerResult("verify_fix", "not-fix-pending", current, [])

    if verdict is not None:
        # Driven directly by the PR's own merge/close outcome: that is
        # already ground truth, so skip the LLM classifier entirely.
        verification = FixVerification(status=verdict)
    else:
        # Classify the exact trust-checked triggering comment, not whatever
        # is newest at fetch time: the router only routes a fix verdict here
        # for a maintainer comment (a `created` event, always carrying a
        # real comment_body), so there is nothing to fall back to -- and
        # falling back to comments[-1] would reintroduce the very race this
        # guards against (a reporter racing a later comment in during the
        # pipeline's multi-minute checkout window), or silently swap in a
        # different comment for an intentionally empty one (e.g.
        # `--issue --action created` with no --comment-body).
        latest = comment_body
        report = rt.report(issue).read()
        verification = rt.verify_fix(
            latest_comment=latest,
            report=report,
            model_spec=rt.ctx.verification_model,
        )

    if verification.status == "confirmed":
        rt.gh.swap_label(issue.number, labels.fix_pending, labels.fix_verified)
        pr = _find_fix_pr(rt, issue.number)
        _tag_fix_pr(rt, pr)
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
        pr = _find_fix_pr(rt, issue.number)
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


def _find_fix_pr(rt: Runtime, number: int) -> Optional[dict]:
    """Best-effort: locate the bot's own open draft PR for this issue."""
    branch = rt.gh.find_branch([f"triage/fix-{number}", f"fix/issue-{number}"])
    if not branch:
        return None
    return rt.gh.find_pull_request(branch)


def _tag_fix_pr(rt: Runtime, pr: Optional[dict]) -> None:
    """Tag the open fix PR so the CI side can key off the confirmation."""
    if pr and pr.get("number") is not None:
        rt.gh.add_labels(pr["number"], [rt.ctx.labels.pr_fix_verified])
