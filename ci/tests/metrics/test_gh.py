#
# Copyright 2026 Canonical, Ltd.
#
"""GitHub Actions API client tests."""

import sys
import unittest
from pathlib import Path
from typing import Any, Dict, List, Optional

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.gh import GitHubClient  # noqa: E402


class FakeResponse:
    def __init__(self, payload: Dict[str, Any]) -> None:
        self.status_code = 200
        self.headers: Dict[str, str] = {}
        self.text = ""
        self.content = b""
        self._payload = payload

    def json(self) -> Dict[str, Any]:
        return self._payload


class FakeSession:
    def __init__(self) -> None:
        self.calls: List[Dict[str, Any]] = []

    def mount(self, *args: Any, **kwargs: Any) -> None:
        return None

    def get(
        self,
        url: str,
        headers: Optional[Dict[str, str]] = None,
        params: Optional[Dict[str, Any]] = None,
        timeout: int = 0,
        allow_redirects: bool = True,
    ) -> FakeResponse:
        del headers, timeout, allow_redirects
        params = dict(params or {})
        self.calls.append({"url": url, "params": params})
        page = int(params.get("page", 1))
        if page == 1:
            return FakeResponse({"total_count": 3, "jobs": [{"id": 1}]})
        if page == 2:
            return FakeResponse({"total_count": 3, "jobs": [{"id": 2}]})
        if page == 3:
            return FakeResponse({"total_count": 3, "jobs": [{"id": 3}]})
        return FakeResponse({"total_count": 3, "jobs": []})


class GitHubClientTest(unittest.TestCase):
    def test_paginate_all_uses_total_count_to_fetch_remaining_pages(self) -> None:
        session = FakeSession()
        client = GitHubClient(session=session, token="token")

        items = client.paginate_all(
            "repos/o/r/actions/runs/1/jobs",
            "jobs",
            params={"per_page": 1},
            workers=2,
        )

        self.assertEqual(items, [{"id": 1}, {"id": 2}, {"id": 3}])
        self.assertEqual([call["params"]["page"] for call in session.calls], [1, 2, 3])

    def test_list_run_jobs_uses_attempt_endpoint(self) -> None:
        session = FakeSession()
        client = GitHubClient(repo="o/r", session=session, token="token")

        client.list_run_jobs(123, attempt=2)

        self.assertIn(
            "repos/o/r/actions/runs/123/attempts/2/jobs",
            session.calls[0]["url"],
        )


if __name__ == "__main__":
    unittest.main()
