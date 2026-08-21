#
# Copyright 2026 Canonical, Ltd.
#
"""Event handlers and the router-to-handler dispatch.

:func:`dispatch` is the single entry point the CLI calls: it routes a parsed
:class:`~triage_bot.router.GitHubEvent` to the matching handler, fetching the
issue only when a handler will actually run. The handlers themselves live in
sibling modules and receive a :class:`~triage_bot.handlers.base.Runtime`.
"""

from __future__ import annotations

from ..router import (
    Action,
    Cleanup,
    GitHubEvent,
    Retriage,
    Skip,
    Triage,
    VerifyFix,
    route,
)
from .base import HandlerResult, Runtime, fetch_issue
from .cleanup import handle_cleanup
from .retriage import handle_retriage
from .triage import handle_triage
from .verify_fix import handle_verify_fix

__all__ = ["dispatch", "HandlerResult", "Runtime"]


def dispatch(event: GitHubEvent, rt: Runtime) -> HandlerResult:
    """Route an event and run the selected handler."""
    action: Action = route(event, rt.ctx.labels)

    if isinstance(action, Skip):
        return HandlerResult("skip", action.reason, None, [])

    issue = fetch_issue(rt.gh, action.issue_number)

    if isinstance(action, Triage):
        return handle_triage(rt, issue, trusted=action.trusted)
    if isinstance(action, Retriage):
        return handle_retriage(
            rt,
            issue,
            action.current_label,
            trusted=action.trusted,
            comment_body=action.comment_body,
        )
    if isinstance(action, VerifyFix):
        return handle_verify_fix(
            rt, issue, comment_body=action.comment_body, verdict=action.verdict
        )
    if isinstance(action, Cleanup):
        return handle_cleanup(rt, issue)

    return HandlerResult("skip", "unhandled action", None, [])
