#
# Copyright 2026 Canonical, Ltd.
#
"""``created`` on a re-triageable issue -> maybe re-run triage.

A new, non-bot comment on an issue resting at a re-triageable label can mean
one of three different things, depending on why the bot is waiting:

- **needs-triage / failed** -- the issue is already fully classified, or the
  pipeline crashed for reasons unrelated to the report's content. Either way
  the open question is trust, not information, so any comment from a trusted
  actor is enough to re-enter triage; there is nothing about its content to
  classify.
- **needs-reproduction / unable-to-reproduce** -- the bot is waiting on a
  specific thing (usually an inspection report) that the reporter may not be
  able to provide. A new comment is classified -- conservatively, via the
  injected ``decide_retriage`` seam -- as carrying the missing detail, an
  explicit decline (flagged for manual review instead of asked again), or
  neither.
- **unable-to-fix / fix-rejected** -- the bot is inviting genuinely new
  diagnostic detail; the same classifier decides whether the comment carries
  it.

The ``max_triage_failures`` cap applies uniformly so a back-and-forth thread
cannot loop forever.
"""

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

# Waiting on trust, not information: a comment's content is never classified
# here, only whether its author is trusted.
_TRUST_GATED = "needs_triage", "failed"
# Waiting on a specific thing the reporter may be unable to provide; a clear
# decline moves to manual review instead of re-asking forever.
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

    # The router only constructs Retriage for a `created` event, which
    # always carries a real comment_body: no fallback needed, and falling
    # back to comments[-1] would risk classifying a different comment than
    # the one that actually triggered this run.
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

    # New info: reset to needs-triage and re-run triage. The pipeline itself
    # runs only for a trusted (maintainer) trigger; an untrusted reporter's
    # new info is re-classified and re-parked, never sent to the cluster.
    rt.gh.swap_label(issue.number, current_label, labels.needs_triage)
    issue.labels = [label for label in issue.labels if label != current_label] + [
        labels.needs_triage
    ]
    result = handle_triage(rt, issue, trusted=trusted, recheck_gates=False)
    result.action = "retriage"
    result.actions_taken = ["retriage:new_info", *result.actions_taken]
    return result
