#
# Copyright 2026 Canonical, Ltd.
#
"""The pure routing table: (event, current label) -> action, no I/O."""

from __future__ import annotations

from triage_bot.labels import LabelConfig
from triage_bot.router import (
    Cleanup,
    GitHubEvent,
    Retriage,
    Skip,
    Triage,
    VerifyFix,
    route,
)

LABELS = LabelConfig()


def _event(action, *, labels=None, is_pr=False, author=None, bots=(), association=""):
    return GitHubEvent(
        action=action,
        issue_number=7,
        issue_labels=labels or [],
        is_pull_request=is_pr,
        comment_author=author,
        bot_logins=bots,
        author_association=association,
    )


def test_opened_routes_to_triage():
    assert route(_event("opened"), LABELS) == Triage(7)


def test_reopened_routes_to_triage():
    assert route(_event("reopened"), LABELS) == Triage(7)


def test_closed_routes_to_cleanup():
    assert route(_event("closed"), LABELS) == Cleanup(7)


def test_pull_request_is_skipped():
    action = route(_event("opened", is_pr=True), LABELS)
    assert isinstance(action, Skip)


def test_comment_on_fix_pending_verifies():
    action = route(
        _event("created", labels=[LABELS.fix_pending], association="MEMBER"), LABELS
    )
    assert action == VerifyFix(7)


def test_untrusted_comment_on_fix_pending_skips():
    # A fix verdict must come from a maintainer; a reporter cannot self-verify.
    action = route(_event("created", labels=[LABELS.fix_pending]), LABELS)
    assert isinstance(action, Skip)


def test_comment_on_retriageable_retriages():
    action = route(_event("created", labels=[LABELS.needs_reproduction]), LABELS)
    assert action == Retriage(7, LABELS.needs_reproduction)


def test_comment_on_terminal_is_skipped():
    action = route(_event("created", labels=[LABELS.fix_verified]), LABELS)
    assert isinstance(action, Skip)


def test_comment_without_triage_label_is_skipped():
    action = route(_event("created", labels=["kind/bug"]), LABELS)
    assert isinstance(action, Skip)


def test_bot_comment_is_skipped():
    action = route(
        _event(
            "created",
            labels=[LABELS.needs_reproduction],
            author="triage[bot]",
            bots=("triage[bot]",),
        ),
        LABELS,
    )
    assert isinstance(action, Skip)


def test_fix_pending_takes_precedence_over_retriage():
    # fix-pending is not in the retriageable set, so a maintainer comment verifies.
    action = route(
        _event("created", labels=[LABELS.fix_pending], association="OWNER"), LABELS
    )
    assert isinstance(action, VerifyFix)


def test_unknown_action_is_skipped():
    assert isinstance(route(_event("labeled"), LABELS), Skip)


def test_from_payload_null_pull_request_is_issue():
    # GitHub omits pull_request for issues; a defensive null must not read as PR.
    event = GitHubEvent.from_payload(
        {"action": "opened", "issue": {"number": 3, "pull_request": None}},
        bot_logins=(),
    )
    assert event.is_pull_request is False
    assert isinstance(route(event, LABELS), Triage)


def test_from_payload_dict_pull_request_is_pr():
    event = GitHubEvent.from_payload(
        {"action": "opened", "issue": {"number": 3, "pull_request": {"url": "x"}}},
        bot_logins=(),
    )
    assert event.is_pull_request is True
    assert isinstance(route(event, LABELS), Skip)


def test_from_payload_extracts_comment_author_and_labels():
    event = GitHubEvent.from_payload(
        {
            "action": "created",
            "issue": {"number": 3, "labels": [{"name": "triage/needs-reproduction"}]},
            "comment": {"user": {"login": "octocat"}},
        },
        bot_logins=(),
    )
    assert event.comment_author == "octocat"
    assert event.issue_labels == ["triage/needs-reproduction"]
