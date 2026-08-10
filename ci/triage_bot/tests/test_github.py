#
# Copyright 2026 Canonical, Ltd.
#
"""Guards on the GitHub boundary that don't need a live ``gh``.

These pin the write-suppression contract: in ``dry_run`` no network write may
leave the process. ``_api`` is stubbed to explode so any leak fails loudly.
"""

from __future__ import annotations

import os
import subprocess

import pytest

from triage_bot import github
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


def test_push_branch_keeps_the_token_out_of_argv(monkeypatch):
    # The whole point: a token embedded in the remote URL sits in this
    # process's own argv (readable via ps/proc) for as long as the push
    # runs. It must only ever reach git through the askpass child's env.
    calls = []

    class _Done:
        returncode = 0
        stdout = ""
        stderr = ""

    def fake_run(cmd, *a, **k):
        calls.append((cmd, k.get("env")))
        return _Done()

    monkeypatch.setattr("triage_bot.github.subprocess.run", fake_run)
    monkeypatch.setenv("GH_TOKEN", "secrettoken")
    gh = GitHubClient(repo="canonical/k8s-snap", dry_run=False)
    assert gh.push_branch("triage/fix-42", cwd="/repo") is True
    # 1) verify branch exists locally, 2) push to authenticated URL.
    assert calls[0][0][:4] == ["git", "-C", "/repo", "rev-parse"]
    push_cmd, push_env = calls[1]
    assert push_cmd[:5] == ["git", "-C", "/repo", "push", "--force"]
    assert push_cmd[5] == "https://x-access-token@github.com/canonical/k8s-snap.git"
    assert push_cmd[6] == "refs/heads/triage/fix-42:refs/heads/triage/fix-42"
    assert "secrettoken" not in push_cmd
    assert " ".join(push_cmd).count("secrettoken") == 0
    # The token instead flows through the askpass child's environment.
    assert push_env["TRIAGE_PUSH_TOKEN"] == "secrettoken"
    assert push_env["GIT_TERMINAL_PROMPT"] == "0"
    askpass_path = push_env["GIT_ASKPASS"]
    # The helper is cleaned up once the push completes.
    assert not os.path.exists(askpass_path)


def test_askpass_helper_answers_password_with_the_token_only(monkeypatch):
    monkeypatch.setenv("TRIAGE_PUSH_TOKEN", "secrettoken")
    path = github._write_askpass_helper()
    try:
        password = subprocess.run(
            [path, "Password for 'https://x-access-token@github.com': "],
            capture_output=True,
            text=True,
        )
        username = subprocess.run(
            [path, "Username for 'https://github.com': "],
            capture_output=True,
            text=True,
        )
        assert password.stdout == "secrettoken"
        assert username.stdout == "x-access-token"
        assert "secrettoken" not in username.stdout
    finally:
        os.unlink(path)


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
