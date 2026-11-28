#
# Copyright 2026 Canonical, Ltd.
#
"""Read-only GitHub Actions API client for the metrics pipeline."""

from __future__ import annotations

import logging
import math
import os
import shutil
import subprocess
import time
from concurrent import futures
from typing import TYPE_CHECKING, Any, Dict, Iterator, List, Optional, Tuple

if TYPE_CHECKING:
    import requests

LOG = logging.getLogger(__name__)

DEFAULT_API_ROOT = "https://api.github.com"
DEFAULT_REPO = "canonical/k8s-snap"
PER_PAGE = 100
# Deep pages of very large runs (Nightly lists ~2,500 jobs) intermittently
# 502 at 100 per page while the same items load fine in smaller pages.
FALLBACK_PER_PAGE = 25
RATE_LIMIT_FLOOR = 200
RETRY_STATUSES = (500, 502, 503, 504)
MAX_RETRIES = 4


class GitHubError(RuntimeError):
    """Raised when the GitHub API returns an unrecoverable error."""


class GitHubUnavailable(GitHubError):
    """Raised when a request still fails after every retry."""


def _token_from_gh_cli() -> Optional[str]:
    if not shutil.which("gh"):
        return None
    try:
        result = subprocess.run(  # noqa: S603
            ["gh", "auth", "token"],
            capture_output=True,
            text=True,
            timeout=10,
            check=False,
        )
    except (OSError, subprocess.SubprocessError):
        return None
    token = result.stdout.strip()
    return token or None


class GitHubClient:
    """Minimal, rate-limit-aware, read-only GitHub Actions API client."""

    def __init__(
        self,
        repo: str = DEFAULT_REPO,
        token: Optional[str] = None,
        api_root: str = DEFAULT_API_ROOT,
        session: Optional[requests.Session] = None,
    ) -> None:
        self.repo = repo
        self.api_root = api_root.rstrip("/")
        self.token = (
            token
            or os.environ.get("GH_TOKEN")
            or os.environ.get("GITHUB_TOKEN")
            or _token_from_gh_cli()
        )
        if not self.token:
            LOG.warning(
                "No GitHub token found; using unauthenticated read-only API access."
            )
        # Imported lazily: k8s-ci.py loads every command module at startup, and
        # the e2e workflow runs its commands with a bare python3 without requests.
        import requests

        self.session = session or requests.Session()
        self.session.mount(
            "https://",
            requests.adapters.HTTPAdapter(pool_connections=32, pool_maxsize=32),
        )
        self.calls_made = 0

    def _headers(self) -> Dict[str, str]:
        headers = {
            "Accept": "application/vnd.github+json",
            "X-GitHub-Api-Version": "2022-11-28",
            "User-Agent": "k8s-ci-metrics",
        }
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        return headers

    def _respect_rate_limit(self, response: requests.Response) -> None:
        remaining = response.headers.get("x-ratelimit-remaining")
        reset = response.headers.get("x-ratelimit-reset")
        if remaining is None or reset is None:
            return
        try:
            remaining_i = int(remaining)
            reset_i = int(reset)
        except ValueError:
            return
        if remaining_i > RATE_LIMIT_FLOOR:
            return
        wait = max(0, reset_i - int(time.time())) + 5
        LOG.warning(
            "Rate limit nearly exhausted (%s remaining); sleeping %ss",
            remaining_i,
            wait,
        )
        time.sleep(wait)

    def request(
        self, path: str, params: Optional[Dict[str, Any]] = None, raw: bool = False
    ) -> Tuple[int, Any]:
        """Perform a read-only GET request, retrying transient failures."""

        import requests

        url = path if path.startswith("http") else f"{self.api_root}/{path.lstrip('/')}"
        last_exc: Optional[Exception] = None

        for attempt in range(MAX_RETRIES):
            try:
                response = self.session.get(
                    url,
                    headers=self._headers(),
                    params=params,
                    timeout=60,
                    allow_redirects=True,
                )
            except requests.RequestException as exc:  # pragma: no cover - network
                last_exc = exc
                sleep_for = 2**attempt
                LOG.warning("Request error %s; retrying in %ss", exc, sleep_for)
                time.sleep(sleep_for)
                continue

            self.calls_made += 1
            self._respect_rate_limit(response)

            if response.status_code in RETRY_STATUSES:
                sleep_for = 2**attempt
                LOG.warning(
                    "HTTP %s from %s; retrying in %ss",
                    response.status_code,
                    url,
                    sleep_for,
                )
                time.sleep(sleep_for)
                continue

            if response.status_code == 403 and "rate limit" in response.text.lower():
                sleep_for = 60 * (attempt + 1)
                LOG.warning("Secondary rate limit hit; sleeping %ss", sleep_for)
                time.sleep(sleep_for)
                continue

            if response.status_code in (404, 410):
                return response.status_code, None

            if response.status_code >= 400:
                raise GitHubError(
                    f"HTTP {response.status_code} from {url}: {response.text[:400]}"
                )

            if raw:
                return response.status_code, response.content
            return response.status_code, response.json()

        raise GitHubUnavailable(f"Exhausted retries for {url}: {last_exc}")

    def get(self, path: str, params: Optional[Dict[str, Any]] = None) -> Any:
        _, payload = self.request(path, params=params)
        return payload

    def paginate(
        self, path: str, key: str, params: Optional[Dict[str, Any]] = None
    ) -> Iterator[Dict[str, Any]]:
        """Yield every item from a paginated endpoint."""

        params = dict(params or {})
        params.setdefault("per_page", PER_PAGE)
        page = 1
        seen = 0

        while True:
            params["page"] = page
            payload = self.get(path, params=params)
            if not payload:
                return
            items = payload.get(key, [])
            if not items:
                return
            for item in items:
                yield item
            seen += len(items)
            total = payload.get("total_count")
            if total is not None and seen >= total:
                return
            page += 1

    def paginate_all(
        self,
        path: str,
        key: str,
        params: Optional[Dict[str, Any]] = None,
        workers: int = 8,
    ) -> List[Dict[str, Any]]:
        """Fetch page 1, then fetch pages 2..N in parallel when possible."""

        params = dict(params or {})
        params.setdefault("per_page", PER_PAGE)
        params["page"] = 1
        first = self._page(path, key, params, 1)
        if not first:
            return []

        items: List[Dict[str, Any]] = list(first.get(key, []))
        total = first.get("total_count")
        if total is None:
            rest = dict(params)
            rest.pop("page", None)
            return list(self.paginate(path, key, params=rest))
        if len(items) >= total:
            return items

        last_page = math.ceil(total / int(params["per_page"]))
        pages = range(2, last_page + 1)
        results: Dict[int, List[Dict[str, Any]]] = {}
        with futures.ThreadPoolExecutor(max_workers=workers) as pool:
            calls = {
                pool.submit(self._page, path, key, params, page): page for page in pages
            }
            for future in futures.as_completed(calls):
                results[calls[future]] = (future.result() or {}).get(key, [])

        for page in pages:
            items.extend(results.get(page, []))
        return items

    def _page(
        self, path: str, key: str, params: Dict[str, Any], page: int
    ) -> Optional[Dict[str, Any]]:
        """Fetch one page, re-fetching its items in smaller pages if it fails."""

        try:
            return self.get(path, {**params, "page": page})
        except GitHubUnavailable:
            size = int(params["per_page"])
            if size <= FALLBACK_PER_PAGE or size % FALLBACK_PER_PAGE:
                raise
        factor = size // FALLBACK_PER_PAGE
        LOG.warning(
            "Page %s of %s failed at %s per page; retrying as %s pages of %s",
            page,
            path,
            size,
            factor,
            FALLBACK_PER_PAGE,
        )
        merged: Dict[str, Any] = {key: []}
        first = (page - 1) * factor + 1
        for sub_page in range(first, first + factor):
            payload = self.get(
                path, {**params, "per_page": FALLBACK_PER_PAGE, "page": sub_page}
            )
            if not payload:
                break
            if "total_count" in payload:
                merged["total_count"] = payload["total_count"]
            merged[key].extend(payload.get(key, []))
        return merged

    def get_run(self, run_id: int) -> Dict[str, Any]:
        return self.get(f"repos/{self.repo}/actions/runs/{run_id}")

    def list_workflows(self) -> List[Dict[str, Any]]:
        return self.paginate_all(f"repos/{self.repo}/actions/workflows", "workflows")

    def list_workflow_runs(
        self,
        workflow_id: Any,
        created: Optional[str] = None,
        event: Optional[str] = None,
        status: Optional[str] = None,
        limit: Optional[int] = None,
    ) -> List[Dict[str, Any]]:
        params: Dict[str, Any] = {}
        if created:
            params["created"] = created
        if event:
            params["event"] = event
        if status:
            params["status"] = status

        runs: List[Dict[str, Any]] = []
        path = f"repos/{self.repo}/actions/workflows/{workflow_id}/runs"
        for run in self.paginate(path, "workflow_runs", params=params):
            runs.append(run)
            if limit is not None and len(runs) >= limit:
                break
        return runs

    def list_run_jobs(self, run_id: int, attempt: int) -> List[Dict[str, Any]]:
        path = f"repos/{self.repo}/actions/runs/{run_id}/attempts/{attempt}/jobs"
        return self.paginate_all(path, "jobs")

    def download_job_log(self, job_id: int) -> Tuple[int, Optional[str]]:
        """Return ``(status, log)``; ``log`` is ``None`` when GitHub has none (404/410)."""

        status, payload = self.request(
            f"repos/{self.repo}/actions/jobs/{job_id}/logs", raw=True
        )
        if payload is None:
            return status, None
        return status, payload.decode("utf-8", errors="replace")
