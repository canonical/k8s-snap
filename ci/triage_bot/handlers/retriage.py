#
# Copyright 2026 Canonical, Ltd.
#
"""``created`` on a re-triageable issue -> maybe re-run triage."""

from __future__ import annotations

from .base import (
    HandlerResult,
    IssueContext,
    Runtime,
    count_failures,
    last_bot_comment,
    maintainer_ping,
    with_marker,
)
from .triage import handle_triage

# Trust-gated states bypass LLM classification.
_TRUST_GATED = "needs_triage", "failed"
# Information-gated states where decline moves to manual review.
_DECLINABLE = "needs_reproduction", "unable_to_reproduce"


def handle_retriage(
    rt: Runtime,
    issue: IssueContext,
    current_label: str,
    *,
    trusted: bool = False,
    comment_body: str = "",
) -> HandlerResult:
    labels = rt.ctx.labels

    if count_failures(issue.comments, rt.ctx.bot_logins) >= rt.ctx.max_triage_failures:
        return HandlerResult(
            "retriage",
            "skipped-max-failures",
            current_label,
            [f"skip:max_failures>={rt.ctx.max_triage_failures}"],
        )

    if current_label in (getattr(labels, name) for name in _TRUST_GATED):
        if not trusted:
            return HandlerResult(
                "retriage", "untrusted-comment", current_label, ["skip:untrusted"]
            )
        result = handle_triage(rt, issue, trusted=True, recheck_gates=False)
        result.action = "retriage"
        result.actions_taken = ["retriage:maintainer_comment", *result.actions_taken]
        return result

    # Classify the triggering comment.
    latest = comment_body
    report = rt.report(issue).read()
    prior_request = last_bot_comment(issue.comments, rt.ctx.bot_logins)
    decision = rt.decide_retriage(
        latest_comment=latest,
        report=report,
        prior_request=prior_request,
        model_spec=rt.ctx.verification_model,
    )

    if decision.outcome == "declined" and current_label in (
        getattr(labels, name) for name in _DECLINABLE
    ):
        rt.gh.swap_label(issue.number, current_label, labels.needs_manual_review)
        rt.gh.add_comment(
            issue.number,
            with_marker(
                f"{maintainer_ping(rt)}the reporter is unable to provide what "
                "automated reproduction needs. Please reproduce this by hand "
                "using the details already on the issue, or close it if "
                "there isn't enough information to proceed."
            ),
        )
        return HandlerResult(
            "retriage", "declined", labels.needs_manual_review, ["retriage:declined"]
        )

    if decision.outcome != "retriage":
        return HandlerResult(
            "retriage", "no-new-info", current_label, ["skip:no_new_info"]
        )

    # Reset to needs-triage and re-run triage.
    rt.gh.swap_label(issue.number, current_label, labels.needs_triage)
    issue.labels = [label for label in issue.labels if label != current_label] + [
        labels.needs_triage
    ]
    result = handle_triage(rt, issue, trusted=trusted, recheck_gates=False)
    result.action = "retriage"
    result.actions_taken = ["retriage:new_info", *result.actions_taken]
    return result
