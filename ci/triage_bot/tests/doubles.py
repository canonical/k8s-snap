#
# Copyright 2026 Canonical, Ltd.
#
"""Shared test doubles for the triage bot.

The whole FSM is exercised offline by substituting these for the two real
side-effect surfaces: ``FakeGitHub`` for the GitHub boundary (records every
label swap, comment, PR, and branch delete) and small canned callables for the
LLM-backed seams (classify, pipeline, retriage, verify).
"""

from __future__ import annotations

from urllib.parse import unquote

from triage_bot.schema import (
    Classification,
    FixVerification,
    RetriageDecision,
    TriageResult,
)


class FakeGitHub:
    """In-memory stand-in for ``GitHubClient`` that records all writes."""

    def __init__(
        self,
        *,
        repo: str = "canonical/k8s-snap",
        issue: dict | None = None,
        comments: list | None = None,
        search: list[dict] | None = None,
        labels: list[str] | None = None,
        branches: list[str] | None = None,
        local_branches: list[str] | None = None,
        dry_run: bool = False,
        bot_login: str = "github-actions[bot]",
    ):
        self.repo = repo
        self.dry_run = dry_run
        self._issue = issue or {}
        # Comments seed as bare strings (authored by a reporter) or as
        # ``(body, author)`` tuples when authorship matters (e.g. bot-authored
        # failure markers that the failure cap must count).
        self._bot_login = bot_login
        self._comments = [self._norm_comment(c) for c in (comments or [])]
        self._search = search or []
        # Live label set on the issue, seeded and then mutated by swaps.
        self.labels: list[str] = list(
            labels if labels is not None else self._issue.get("labels_list", [])
        )
        self._branches = list(branches or [])
        # Branches the fix skill committed locally in the checkout. The real
        # client refuses to push a branch that was never committed; None means
        # "unmodelled" -> push always succeeds, matching pre-existing tests.
        self._local_branches = local_branches
        # Recorded write history for assertions.
        self.added_labels: list[str] = []
        self.removed_labels: list[str] = []
        self.comments_posted: list[str] = []
        self.pulls_created: list[dict] = []
        self.deleted_branches: list[str] = []
        self.pushed_branches: list[str] = []
        self.pr_labels: dict[int, list[str]] = {}

    @staticmethod
    def _norm_comment(c) -> dict:
        if isinstance(c, tuple):
            return {"body": c[0], "author": c[1]}
        return {"body": c, "author": "reporter"}

    # --- reads ---

    def get_issue(self, number: int) -> dict:
        data = dict(self._issue)
        data.setdefault("number", number)
        data["labels"] = [{"name": name} for name in self.labels]
        return data

    def get_comments(self, number: int) -> list[dict]:
        return [
            {"body": c["body"], "user": {"login": c["author"]}} for c in self._comments
        ]

    def search_issues(self, query: str) -> list[dict]:
        return self._search

    def list_labels(self, number: int) -> list[str]:
        return list(self.labels)

    # --- writes ---

    def add_labels(self, number: int, labels: list[str]) -> None:
        if self.dry_run or not labels:
            return
        for label in labels:
            # A number seeded in pr_labels is a PR, kept separate so issue
            # label assertions stay clean.
            if number in self.pr_labels:
                self.pr_labels[number].append(label)
                continue
            self.added_labels.append(label)
            if label not in self.labels:
                self.labels.append(label)

    def add_comment(self, number: int, body: str) -> None:
        if self.dry_run or not body:
            return
        self.comments_posted.append(body)
        # The bot posts as its own login, so a failure marker it writes is
        # attributable and counts toward the cap; a reporter cannot forge it.
        self._comments.append({"body": body, "author": self._bot_login})

    def remove_label(self, number: int, label: str) -> None:
        if self.dry_run:
            return
        self.removed_labels.append(label)
        self.labels = [x for x in self.labels if x != label]

    def swap_label(self, number: int, old, new: str) -> None:
        if old and old != new:
            self.remove_label(number, old)
        self.add_labels(number, [new])

    def push_branch(self, branch: str, cwd: str) -> bool:
        if self.dry_run:
            return False
        if self._local_branches is not None and branch not in self._local_branches:
            return False
        self.pushed_branches.append(branch)
        if branch not in self._branches:
            self._branches.append(branch)
        return True

    def create_pull_request(self, *, head: str, base: str, title: str, body: str):
        if self.dry_run:
            return None
        number = 9000 + len(self.pulls_created)
        pr = {
            "number": number,
            "head": head,
            "base": base,
            "title": title,
            "body": body,
            "html_url": f"https://github.com/{self.repo}/pull/{number}",
        }
        self.pulls_created.append(pr)
        self.pr_labels[number] = []
        if head not in self._branches:
            self._branches.append(head)
        return pr

    def find_pull_request(self, head: str, state: str = "open"):
        for pr in self.pulls_created:
            if pr["head"] == head:
                return pr
        return None

    def find_branch(self, candidates: list[str]):
        for branch in candidates:
            if branch in self._branches:
                return branch
        return None

    def delete_branch(self, branch: str) -> None:
        if self.dry_run:
            return
        self.deleted_branches.append(branch)
        self._branches = [b for b in self._branches if b != branch]


def make_classifier(classification: Classification):
    """A classify seam that ignores its inputs and returns a canned result."""

    def classify(**kwargs) -> Classification:
        return classification

    return classify


def make_pipeline(result: TriageResult):
    """A pipeline seam returning a canned :class:`TriageResult`."""

    def pipeline(rt, issue) -> TriageResult:
        return result

    return pipeline


def make_retriage(decision: RetriageDecision):
    def decide(**kwargs) -> RetriageDecision:
        return decision

    return decide


def make_verifier(verification: FixVerification):
    def verify(**kwargs) -> FixVerification:
        return verification

    return verify


# quote/unquote round-trips are asserted in a couple of label tests.
__all__ = [
    "FakeGitHub",
    "make_classifier",
    "make_pipeline",
    "make_retriage",
    "make_verifier",
    "unquote",
]
