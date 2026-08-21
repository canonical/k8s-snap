#
# Copyright 2026 Canonical, Ltd.
#
"""Guards on the CLI's manual/local invocation path (``--issue``).

``--issue`` mode builds a synthetic webhook-style payload instead of reading
one from disk; these pin that the synthetic payload is faithful enough for
downstream handlers, not a plausible-looking shape that silently loses data
(e.g. the triggering comment body for a simulated ``created`` event).
"""

from __future__ import annotations

from types import SimpleNamespace

from cmds.triage import _load_event
from triage_bot.tests.doubles import FakeGitHub

ISSUE = 42


def _args(**overrides):
    defaults = dict(issue=ISSUE, action="opened", event_file=None, comment_body=None)
    return SimpleNamespace(**(defaults | overrides))


def test_issue_mode_builds_a_trusted_opened_payload():
    gh = FakeGitHub(issue={"title": "x"})

    payload = _load_event(_args(), gh)

    assert payload["action"] == "opened"
    assert payload["issue"]["number"] == ISSUE
    assert payload["issue"]["author_association"] == "OWNER"
    assert "comment" not in payload


def test_issue_mode_created_carries_the_given_comment_body():
    gh = FakeGitHub(issue={"title": "x"})

    payload = _load_event(
        _args(action="created", comment_body="the fix works, thanks"), gh
    )

    assert payload["comment"]["body"] == "the fix works, thanks"
    assert payload["comment"]["author_association"] == "OWNER"


def test_issue_mode_created_without_a_comment_body_is_an_empty_string():
    # Falling back to whatever a downstream handler treats as "latest
    # comment" would silently simulate a different comment than none; an
    # explicit empty string is the honest representation of "not given".
    gh = FakeGitHub(issue={"title": "x"})

    payload = _load_event(_args(action="created"), gh)

    assert payload["comment"]["body"] == ""
