#
# Copyright 2026 Canonical, Ltd.
#
"""``created`` on a fix-pending issue -> verify the fix.

When the bot has posted a candidate fix (issue resting at ``fix-pending``), the
next non-bot comment is a maintainer verdict. The injected ``verify_fix`` seam
classifies it as confirmed / rejected / inconclusive, and the handler swaps to
the matching terminal (or re-triageable) label. On confirmation it also tags the
open fix PR so the CI side can pick it up.
"""

from __future__ import annotations

from .base import HandlerResult, IssueContext, Runtime, with_marker


def handle_verify_fix(
    rt: Runtime, issue: IssueContext, *, comment_body: str = ""
) -> HandlerResult:
    labels = rt.ctx.labels
    # Classify the exact trust-checked triggering comment, not whatever is
    # newest at fetch time: the router only routes a fix verdict here for a
    # maintainer comment (a `created` event, always carrying a real
    # comment_body), so there is nothing to fall back to -- and falling back
    # to comments[-1] would reintroduce the very race this guards against (a
    # reporter racing a later comment in during the multi-minute checkout
    # window), or silently swap in a different comment for an intentionally
    # empty one (e.g. `--issue --action created` with no --comment-body).
    latest = comment_body
    report = rt.report(issue).read()

    verification = rt.verify_fix(
        latest_comment=latest,
        report=report,
        model_spec=rt.ctx.verification_model,
    )

    if verification.status == "confirmed":
        rt.gh.swap_label(issue.number, labels.fix_pending, labels.fix_verified)
        _tag_fix_pr(rt, issue.number)
        rt.gh.add_comment(
            issue.number,
            with_marker("Fix confirmed. Thanks for verifying."),
        )
        return HandlerResult(
            "verify_fix", "confirmed", labels.fix_verified, ["fix:confirmed"]
        )

    if verification.status == "rejected":
        rt.gh.swap_label(issue.number, labels.fix_pending, labels.fix_rejected)
        rt.gh.add_comment(
            issue.number,
            with_marker(
                "Thanks, marking the proposed fix as rejected. A maintainer "
                "will revisit; further details on what still fails are welcome.",
                failure=True,
            ),
        )
        return HandlerResult(
            "verify_fix", "rejected", labels.fix_rejected, ["fix:rejected"]
        )

    return HandlerResult(
        "verify_fix", "inconclusive", labels.fix_pending, ["fix:inconclusive"]
    )


def _tag_fix_pr(rt: Runtime, number: int) -> None:
    """Best-effort: tag the open fix PR so CI can key off the confirmation."""
    branch = rt.gh.find_branch([f"triage/fix-{number}", f"fix/issue-{number}"])
    if not branch:
        return
    pr = rt.gh.find_pull_request(branch)
    if pr and pr.get("number") is not None:
        rt.gh.add_labels(pr["number"], [rt.ctx.labels.pr_fix_verified])
