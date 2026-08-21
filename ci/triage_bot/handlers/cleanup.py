#
# Copyright 2026 Canonical, Ltd.
#
"""``closed`` -> clean up the fix branch.

When an issue is closed (fix merged, or closed by a maintainer) any dangling
draft-fix branch the bot pushed should be removed. Best-effort and idempotent:
a missing branch is not an error.
"""

from __future__ import annotations

from .base import HandlerResult, IssueContext, Runtime


def handle_cleanup(rt: Runtime, issue: IssueContext) -> HandlerResult:
    branch = rt.gh.find_branch(
        [f"triage/fix-{issue.number}", f"fix/issue-{issue.number}"]
    )
    if not branch:
        return HandlerResult("cleanup", "no-branch", None, ["cleanup:none"])
    rt.gh.delete_branch(branch)
    return HandlerResult("cleanup", "branch-deleted", None, [f"cleanup:{branch}"])
