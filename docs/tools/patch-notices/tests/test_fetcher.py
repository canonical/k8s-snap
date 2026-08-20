# Copyright 2026 Canonical, Ltd.
# See LICENSE file for licensing details.

from __future__ import annotations

from urllib.parse import parse_qs, urlparse
from typing import Any

import pytest

from patch_notices import fetcher


class _Response:
    def __init__(self, payload: dict[str, Any]) -> None:
        self._payload = payload

    def raise_for_status(self) -> None:
        return None

    def json(self) -> dict[str, Any]:
        return self._payload


def _commit(sha: str, message: str | None = None) -> dict[str, Any]:
    return {
        "sha": sha,
        "html_url": f"https://github.com/canonical/k8s-snap/commit/{sha}",
        "author": {"login": "octocat"},
        "commit": {
            "message": message or f"fix: update {sha[:8]}",
            "author": {"name": "Octo Cat", "date": "2026-08-19T10:00:00Z"},
        },
    }


def test_github_commits_single_page(monkeypatch: pytest.MonkeyPatch) -> None:
    commits = [_commit("a" * 40), _commit("b" * 40)]

    def fake_get(url: str, **_: Any) -> _Response:
        query = parse_qs(urlparse(url).query)
        assert query == {"per_page": ["250"]}
        return _Response({"total_commits": 2, "commits": commits})

    monkeypatch.setattr(fetcher.requests, "get", fake_get)
    monkeypatch.setattr(fetcher, "_github_commit_diff", lambda *_: "diff")

    result = fetcher._github_commits("base", "head")

    assert [entry["sha"] for entry in result] == ["a" * 40, "b" * 40]
    assert all(entry["diff"] == "diff" for entry in result)


def test_github_commits_uses_github_token(monkeypatch: pytest.MonkeyPatch) -> None:
    commit = _commit("a" * 40)

    def fake_get(url: str, **kwargs: Any) -> _Response:
        assert kwargs["headers"]["Authorization"] == "Bearer github-token"
        return _Response({"total_commits": 1, "commits": [commit]})

    monkeypatch.setenv("GITHUB_TOKEN", "github-token")
    monkeypatch.setattr(fetcher.requests, "get", fake_get)
    monkeypatch.setattr(fetcher, "_github_commit_diff", lambda *_: "diff")

    assert fetcher._github_commits("base", "head")[0]["sha"] == "a" * 40


def test_github_commits_detects_truncated_compare_results(monkeypatch: pytest.MonkeyPatch) -> None:
    commits = [_commit(f"{index:040x}") for index in range(250)]
    request_count = 0

    def fake_get(url: str, **_: Any) -> _Response:
        nonlocal request_count
        request_count += 1
        return _Response({"total_commits": 400, "commits": commits})

    monkeypatch.setattr(fetcher.requests, "get", fake_get)
    monkeypatch.setattr(fetcher, "_github_commit_diff", lambda *_: "diff")

    with pytest.raises(ValueError, match="GitHub compare returned 250 of 400 commits"):
        fetcher._github_commits("base", "head")
    assert request_count == 1


def test_github_commits_raises_on_incomplete_compare(monkeypatch: pytest.MonkeyPatch) -> None:
    commits = [_commit(f"{index:040x}") for index in range(10)]

    def fake_get(url: str, **_: Any) -> _Response:
        return _Response({"total_commits": 11, "commits": commits})

    monkeypatch.setattr(fetcher.requests, "get", fake_get)
    monkeypatch.setattr(fetcher, "_github_commit_diff", lambda *_: "diff")

    with pytest.raises(ValueError, match="GitHub compare returned 10 of 11 commits"):
        fetcher._github_commits("base", "head")


def test_github_commits_empty_compare_skips_diff_fetch(monkeypatch: pytest.MonkeyPatch) -> None:
    def fake_get(url: str, **_: Any) -> _Response:
        return _Response({"status": "identical", "total_commits": 0, "commits": []})

    def fail_diff(*_: Any) -> str:
        raise AssertionError("diff fetch should not run for an empty compare")

    monkeypatch.setattr(fetcher.requests, "get", fake_get)
    monkeypatch.setattr(fetcher, "_github_commit_diff", fail_diff)

    assert fetcher._github_commits("base", "head") == []


def test_github_commits_preserves_pr_number_extraction(monkeypatch: pytest.MonkeyPatch) -> None:
    commit = _commit("c" * 40, message="fix: repair the cluster (#1234)\n\nMore details")

    def fake_get(url: str, **_: Any) -> _Response:
        return _Response({"total_commits": 1, "commits": [commit]})

    monkeypatch.setattr(fetcher.requests, "get", fake_get)
    monkeypatch.setattr(fetcher, "_github_commit_diff", lambda *_: "diff")

    result = fetcher._github_commits("base", "head")

    assert result[0]["pr_number"] == 1234
    assert result[0]["pr_url"] == "https://github.com/canonical/k8s-snap/pull/1234"
    assert result[0]["body"] == "More details"