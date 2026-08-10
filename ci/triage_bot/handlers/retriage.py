#
# Copyright 2026 Canonical, Ltd.
#
"""``created`` on a re-triageable issue -> maybe re-run triage.

A new, non-bot comment on an issue resting at a re-triageable label (needs
reproduction, unable to reproduce/fix, failed, fix rejected) may carry the
detail that was missing. This handler decides -- conservatively, via the
injected ``decide_retriage`` seam -- whether to spend another triage run, and
enforces the ``max_triage_failures`` cap so a back-and-forth thread cannot loop
forever.
"""

from __future__ import annotations

from .base import HandlerResult, IssueContext, Runtime, count_failures
from .triage import handle_triage


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

    # The router only constructs Retriage for a `created` event, which
    # always carries a real comment_body: no fallback needed, and falling
    # back to comments[-1] would risk classifying a different comment than
    # the one that actually triggered this run.
    latest = comment_body
    report = rt.report(issue).read()
    decision = rt.decide_retriage(
        latest_comment=latest,
        report=report,
        model_spec=rt.ctx.verification_model,
    )
    if not decision.retriage:
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
