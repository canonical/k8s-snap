#
# Copyright 2026 Canonical, Ltd.
#
"""Shared plumbing for the event handlers.

The handlers are the action-owned orchestration: each one is selected by the
router for a given event and drives the GitHub side effects (label swaps,
comments, PRs) around one or more skill invocations. Everything a handler needs
is bundled in :class:`Runtime`, and the two expensive, non-deterministic seams
-- the skill pipeline and the LLM classifier -- are injected so the handlers run
fully offline in tests with canned results.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Callable, Optional

from .. import triage_core
from ..context import ActionContext
from ..github import GitHubClient
from ..report import Report
from ..schema import (
    Classification,
    EnhancementProposal,
    ExistingSupport,
    FixVerification,
    RetriageDecision,
    TriageResult,
)
from ..skills import repo_root

# Hidden marker appended to every failure comment. Re-triage counts these to
# enforce the ``max_triage_failures`` cap without any external state.
FAILURE_MARKER = "<!-- triage-bot:failure -->"

# Marker on the bot's own triage comments, so its comments are recognisable
# even when posted under a user account in local runs.
BOT_MARKER = "<!-- triage-bot -->"

# Alpha disclaimer appended to every comment.
ALPHA_DISCLAIMER = (
    "\n\n---\n*🤖 This automated triage bot is in alpha. We are open to feedback!*"
)


@dataclass
class Comment:
    """A single issue comment: body plus the author's login."""

    body: str
    author: str = ""


@dataclass
class IssueContext:
    """Everything about the issue the handlers read, fetched once up front."""

    number: int
    title: str
    body: str
    comments: list[Comment] = field(default_factory=list)
    labels: list[str] = field(default_factory=list)


@dataclass
class HandlerResult:
    """What a handler did, for structured logging and the CLI summary."""

    action: str
    outcome: str
    label: Optional[str] = None
    actions_taken: list[str] = field(default_factory=list)


VerifyFn = Callable[..., FixVerification]
SupportFn = Callable[..., ExistingSupport]
ProposeFn = Callable[..., EnhancementProposal]
# Production wires the real skill runner; tests inject a canned function.
PipelineFn = Callable[["Runtime", IssueContext], TriageResult]
ClassifyFn = Callable[..., Classification]
RetriageFn = Callable[..., RetriageDecision]


@dataclass
class Runtime:
    """Bundle threaded through every handler."""

    ctx: ActionContext
    gh: GitHubClient
    pipeline: PipelineFn
    classify: ClassifyFn = triage_core.classify
    decide_retriage: RetriageFn = triage_core.decide_retriage
    verify_fix: VerifyFn = triage_core.verify_fix
    existing_support: SupportFn = triage_core.check_existing_support
    propose_enhancement: ProposeFn = triage_core.propose_enhancement

    def workdir(self, number: int) -> Path:
        # A relative workdir_root is anchored to the checkout root (the shell's
        # cwd), not the process cwd, so scratch (report.md) and the agent shell
        # share one base regardless of where the CLI is invoked; an absolute
        # root (e.g. a test's tmp_path) is used as-is.
        root = Path(self.ctx.workdir_root)
        if not root.is_absolute():
            root = repo_root() / root
        return root / f"issue-{number}"

    def worktree(self, number: int) -> Path:
        """The isolated checkout the agent works in, inside the scratch dir."""
        return self.workdir(number) / "checkout"

    def report(self, issue: IssueContext) -> Report:
        # Non-truncating: the pipeline creates and fills report.md, then
        # retriage/verify-fix read those accumulated findings. Calling
        # .start() here would overwrite them with a bare header (the seams
        # would then decide on empty context).
        return Report(self.workdir(issue.number))


def fetch_issue(gh: GitHubClient, number: int) -> IssueContext:
    """Read issue title/body/labels/comments through the GitHub boundary."""
    issue = gh.get_issue(number)
    comments = [
        Comment(
            body=c.get("body") or "",
            author=(c.get("user") or {}).get("login") or "",
        )
        for c in gh.get_comments(number)
    ]
    labels = [
        label.get("name", "")
        for label in issue.get("labels", [])
        if isinstance(label, dict) and label.get("name")
    ]
    return IssueContext(
        number=number,
        title=issue.get("title") or "",
        body=issue.get("body") or "",
        comments=comments,
        labels=labels,
    )


def count_failures(comments: list[Comment], bot_logins: tuple[str, ...]) -> int:
    """Prior failed triage attempts, counted by the hidden failure marker.

    Only markers on the bot's OWN comments count. The marker lives in an HTML
    comment invisible in rendered markdown, so any user could otherwise paste
    it to forge the cap; GitHub-set authorship (``bot_logins``) is the guard.
    """
    return sum(
        1
        for c in comments
        if FAILURE_MARKER in (c.body or "") and c.author in bot_logins
    )


def last_bot_comment(comments: list[Comment], bot_logins: tuple[str, ...]) -> str:
    """The bot's most recent own comment, or "" if it has never posted one.

    Used to give the retriage classifier the actual prior ask (e.g. "please
    attach an inspection tarball") when ``report.md`` doesn't exist yet -- a
    park via the cheap gates never reaches the pipeline stage that writes it.
    """
    for c in reversed(comments):
        if BOT_MARKER in (c.body or "") and c.author in bot_logins:
            return c.body
    return ""


def maintainer_ping(rt: "Runtime") -> str:
    """``cc @team: `` prefix for a comment that needs maintainer action.

    "" when ``maintainer_team`` is unset, so the opt-out stays silent rather
    than pinging an empty name.
    """
    team = rt.ctx.maintainer_team
    return f"cc @{team}: " if team else ""


def with_marker(body: str, *, failure: bool = False) -> str:
    """Append the bot marker (and optionally the failure marker) to a comment."""
    markers = BOT_MARKER + (FAILURE_MARKER if failure else "")
    return f"{body}{ALPHA_DISCLAIMER}\n\n{markers}"
