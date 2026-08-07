#
# Copyright 2026 Canonical, Ltd.
#
"""Guards on the GitHub boundary that don't need a live ``gh``.

These pin the write-suppression contract: in ``dry_run`` no network write may
leave the process. ``_api`` is stubbed to explode so any leak fails loudly.
"""

from __future__ import annotations

import pytest

from triage_bot.github import GitHubClient, GitHubError


def _no_network(client: GitHubClient) -> None:
    def explode(*a, **k):
        raise AssertionError("dry-run must not touch the network")

    client._api = explode  # type: ignore[assignment]


def test_dry_run_create_pull_request_returns_none_without_api():
    gh = GitHubClient(dry_run=True)
    _no_network(gh)
    assert gh.create_pull_request(head="h", base="main", title="t", body="b") is None


def test_dry_run_push_branch_returns_false():
    gh = GitHubClient(dry_run=True)
    # push_branch must short-circuit before invoking git at all.
    assert gh.push_branch("triage/fix-1", cwd=".") is False


def test_dry_run_label_and_comment_writes_suppressed():
    gh = GitHubClient(dry_run=True)
    _no_network(gh)
    # None of these may reach _api under dry-run.
    gh.add_labels(1, ["kind/bug"])
    gh.add_comment(1, "hello")
    gh.remove_label(1, "kind/bug")
    gh.delete_branch("triage/fix-1")


def test_create_pull_request_raises_on_bad_response():
    gh = GitHubClient(dry_run=False)
    gh._api = lambda *a, **k: ["not", "a", "dict"]  # type: ignore[assignment]
    with pytest.raises(GitHubError):
        gh.create_pull_request(head="h", base="main", title="t", body="b")


def test_push_branch_uses_token_authenticated_remote(monkeypatch):
    # Pin the real argv: persist-credentials:false means the push must carry the
    # token in the URL (not rely on `origin`), with an explicit refspec.
    calls = []

    class _Done:
        returncode = 0
        stdout = ""
        stderr = ""

    def fake_run(cmd, *a, **k):
        calls.append(cmd)
        return _Done()

    monkeypatch.setattr("triage_bot.github.subprocess.run", fake_run)
    monkeypatch.setenv("GH_TOKEN", "secrettoken")
    gh = GitHubClient(repo="canonical/k8s-snap", dry_run=False)
    assert gh.push_branch("triage/fix-42", cwd="/repo") is True
    # 1) verify branch exists locally, 2) push to authenticated URL.
    assert calls[0][:4] == ["git", "-C", "/repo", "rev-parse"]
    push = calls[1]
    assert push[:5] == ["git", "-C", "/repo", "push", "--force"]
    assert (
        push[5]
        == "https://x-access-token:secrettoken@github.com/canonical/k8s-snap.git"
    )
    assert push[6] == "refs/heads/triage/fix-42:refs/heads/triage/fix-42"


def test_push_branch_missing_local_branch_returns_false(monkeypatch):
    calls = []

    class _Missing:
        returncode = 1
        stdout = ""
        stderr = ""

    def fake_run(cmd, *a, **k):
        calls.append(cmd)
        return _Missing()

    monkeypatch.setattr("triage_bot.github.subprocess.run", fake_run)
    gh = GitHubClient(repo="canonical/k8s-snap", dry_run=False)
    # A branch the skill never committed: no push attempted, degrades to no-PR.
    assert gh.push_branch("triage/fix-42", cwd="/repo") is False
    assert len(calls) == 1
