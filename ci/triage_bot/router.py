#
# Copyright 2026 Canonical, Ltd.
#
"""FSM router: event + current label -> the handler action to run.

Pure function of the parsed GitHub event and the issue's current triage label.
No I/O here so the routing table is exhaustively unit-testable offline; the
handlers it selects perform all the GitHub and LLM side effects.
"""

from __future__ import annotations

import re
from dataclasses import dataclass
from typing import Literal, Optional, Union

from .labels import LabelConfig, current_triage_label

TRUSTED_ASSOCIATIONS = frozenset({"OWNER", "MEMBER", "COLLABORATOR"})

# The bot's own draft-fix branch, in either naming convention it has used
# (see handlers/verify_fix.py's _tag_fix_pr and pipeline.py's _branch).
_FIX_BRANCH_RE = re.compile(r"^(?:triage/fix|fix/issue)-(\d+)$")


@dataclass(frozen=True)
class GitHubEvent:
    action: str  # opened | reopened | closed | created (comment)
    issue_number: int
    issue_labels: list[str]
    is_pull_request: bool = False
    comment_author: Optional[str] = None
    bot_logins: tuple[str, ...] = ()
    # author_association of the actor that triggered THIS event (the commenter
    # for a comment, else the issue author). Drives the trust gate: only a
    # maintainer-associated actor may run the shell pipeline / verify a fix.
    author_association: str = ""
    # Body of the triggering comment, threaded onto the verdict/retriage action
    # so the handler classifies the exact trust-checked comment rather than
    # whatever is newest at fetch time (which a reporter could race in).
    comment_body: str = ""
    # Set only when this event IS a native ``pull_request`` webhook for one
    # of the bot's own draft fix branches: merging or closing that PR is
    # itself the maintainer's verdict, so this bypasses the LLM classifier
    # and the comment-based path entirely -- no issue comment is needed.
    pr_verdict: Optional[Literal["confirmed", "rejected"]] = None

    @property
    def trusted(self) -> bool:
        return self.author_association in TRUSTED_ASSOCIATIONS

    @classmethod
    def from_payload(
        cls, payload: dict, bot_logins: tuple[str, ...] = ()
    ) -> "GitHubEvent":
        """Parse a GitHub ``issues``/``issue_comment``/``pull_request`` webhook.

        Unknown or missing fields degrade to a benign no-op event (action
        ``""`` with issue number 0), which the router turns into a Skip.
        """
        pr = payload.get("pull_request")
        if pr is not None:
            # A native pull_request webhook -- the bot's own draft fix PR
            # closing (merged or not) IS the maintainer's verdict, so this
            # never falls through to the issues/issue_comment parsing below.
            head = pr.get("head") or {}
            base = pr.get("base") or {}
            same_repo = (head.get("repo") or {}).get("full_name") == (
                base.get("repo") or {}
            ).get("full_name")
            match = _FIX_BRANCH_RE.match(head.get("ref", ""))
            # Same-repo only: a fork PR's head.ref is entirely attacker-
            # controlled, so an external contributor could otherwise name a
            # branch triage/fix-<n> for an arbitrary existing issue and
            # self-close their own PR to forge that issue's verdict.
            if payload.get("action") != "closed" or not match or not same_repo:
                return cls(action="", issue_number=0, issue_labels=[])
            return cls(
                action="closed",
                issue_number=int(match.group(1)),
                issue_labels=[],
                bot_logins=bot_logins,
                pr_verdict="confirmed" if pr.get("merged") else "rejected",
            )

        issue = payload.get("issue") or {}
        labels = [
            label.get("name", "")
            for label in issue.get("labels", [])
            if isinstance(label, dict) and label.get("name")
        ]
        comment = payload.get("comment") or {}
        author = (comment.get("user") or {}).get("login")
        # Trust binds to the actor of THIS event: a comment event uses only the
        # commenter's association (never the issue author's, or an untrusted
        # commenter on a maintainer's issue would inherit the author's trust); an
        # issue event uses the issue author's association. That is exactly the
        # actor for `opened`; for `reopened` GitHub does not expose the actual
        # reopener's own association here, so a maintainer reopening someone
        # else's untrusted issue reads as untrusted. Only the author or a
        # collaborator can reopen another user's issue, so this can only
        # under-grant trust, never over-grant it.
        is_comment = "comment" in payload
        association = (
            comment.get("author_association")
            if is_comment
            else issue.get("author_association")
        ) or ""
        return cls(
            action=payload.get("action", ""),
            issue_number=int(issue.get("number", 0) or 0),
            issue_labels=labels,
            is_pull_request=bool(issue.get("pull_request")),
            comment_author=author,
            bot_logins=bot_logins,
            author_association=association,
            comment_body=comment.get("body") or "",
        )


@dataclass(frozen=True)
class Triage:
    issue_number: int
    trusted: bool = False


@dataclass(frozen=True)
class Retriage:
    issue_number: int
    current_label: str
    trusted: bool = False
    comment_body: str = ""


@dataclass(frozen=True)
class VerifyFix:
    issue_number: int
    comment_body: str = ""
    # Set directly from a PR merge/close (GitHubEvent.pr_verdict), bypassing
    # the LLM classifier entirely -- see handle_verify_fix.
    verdict: Optional[Literal["confirmed", "rejected"]] = None


@dataclass(frozen=True)
class Cleanup:
    issue_number: int


@dataclass(frozen=True)
class Skip:
    reason: str


Action = Union[Triage, Retriage, VerifyFix, Cleanup, Skip]


def route(event: GitHubEvent, labels: LabelConfig) -> Action:
    """Decide which handler should run for this event.

    Trust (maintainer ``author_association``) is threaded onto the actions that
    can drive privileged work: only a trusted actor's event runs the shell
    pipeline (gated in ``handle_triage``) or is accepted as a fix verdict.
    """
    # A PR-driven verdict is the most specific, authoritative signal: check
    # it before anything else, since such an event also carries action ==
    # "closed" (the same value a plain issue-closed cleanup event uses) --
    # it must never fall into the Cleanup branch below.
    if event.pr_verdict is not None:
        return VerifyFix(event.issue_number, verdict=event.pr_verdict)

    if event.is_pull_request:
        return Skip("event is on a pull request, not an issue")

    if event.action in ("opened", "reopened"):
        return Triage(event.issue_number, trusted=event.trusted)

    if event.action == "closed":
        return Cleanup(event.issue_number)

    if event.action == "created":
        author = event.comment_author
        if author and author in event.bot_logins:
            return Skip(f"comment from bot ({author})")

        current = current_triage_label(event.issue_labels, labels)
        if current == labels.fix_pending:
            # Only a maintainer comment is a fix verdict; anyone else could
            # otherwise force triage/fix-verified.
            if not event.trusted:
                return Skip("fix verdict requires a maintainer comment")
            return VerifyFix(event.issue_number, comment_body=event.comment_body)
        if current is not None and current in labels.retriageable_labels():
            return Retriage(
                event.issue_number,
                current,
                trusted=event.trusted,
                comment_body=event.comment_body,
            )
        return Skip(
            f"terminal label: {current}" if current else "no triage label on issue"
        )

    return Skip(f"unhandled event action: {event.action}")
