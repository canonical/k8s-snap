#
# Copyright 2026 Canonical, Ltd.
#
"""GitHub Actions API client tests."""

import sys
import unittest
from pathlib import Path
from typing import Any, Dict, List, Optional
from unittest import mock

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

    def test_list_workflow_runs_passes_status_filter(self) -> None:
        session = FakeSession()
        client = GitHubClient(repo="o/r", session=session, token="token")

        client.list_workflow_runs("nightly.yaml", status="completed")

        self.assertEqual(session.calls[0]["params"].get("status"), "completed")

    def test_list_workflow_runs_omits_status_when_unset(self) -> None:
        session = FakeSession()
        client = GitHubClient(repo="o/r", session=session, token="token")

        client.list_workflow_runs("nightly.yaml")

        self.assertNotIn("status", session.calls[0]["params"])

    def test_failing_page_is_refetched_in_smaller_pages(self) -> None:
        session = FakeSession()
        jobs = [{"id": index} for index in range(1, 151)]

        def get(url: str, params: Optional[Dict[str, Any]] = None, **_: Any):
            params = dict(params or {})
            session.calls.append({"url": url, "params": params})
            size, page = int(params["per_page"]), int(params["page"])
            if size == 100 and page == 2:
                response = FakeResponse({})
                response.status_code = 502
                return response
            start = (page - 1) * size
            return FakeResponse({"total_count": len(jobs), "jobs": jobs[start:][:size]})

        session.get = get
        client = GitHubClient(repo="o/r", session=session, token="token")

        with mock.patch("metrics.gh.time.sleep"):
            items = client.list_run_jobs(123, attempt=1)

        self.assertEqual(items, jobs)
        fallback = [
            call["params"]["page"]
            for call in session.calls
            if call["params"]["per_page"] == 25
        ]
        self.assertEqual(fallback, [5, 6, 7, 8])

    def test_download_job_log_decodes_log_body(self) -> None:
        session = FakeSession()
        response = FakeResponse({})
        response.content = "AssertionError: boom \u2713\n".encode("utf-8")
        session.get = lambda url, **_: (session.calls.append(url), response)[1]
        client = GitHubClient(repo="o/r", session=session, token="token")

        log = client.download_job_log(42)

        self.assertEqual(log, (200, "AssertionError: boom \u2713\n"))
        self.assertTrue(session.calls[0].endswith("repos/o/r/actions/jobs/42/logs"))

    def test_download_job_log_returns_none_when_log_is_gone(self) -> None:
        for status in (404, 410):
            session = FakeSession()
            response = FakeResponse({})
            response.status_code = status
            session.get = lambda url, **_: response
            client = GitHubClient(repo="o/r", session=session, token="token")

            self.assertEqual(client.download_job_log(42), (status, None))


if __name__ == "__main__":
    unittest.main()
