#
# Copyright 2026 Canonical, Ltd.
#
"""Thin GitHub Actions API client for the metrics pipeline.

Deliberately small: the pipeline only needs runs, jobs, attempts and job logs.
The client is responsible for pagination, rate-limit courtesy and retries, so
callers can treat the API as a plain iterator.

Measured cost basis (run 34172139128, 2555 jobs): 26 paginated calls at
roughly 4.5s each, so a full nightly ingest is ~2 minutes and ~160 API calls
including logs -- about 3% of the 5000/hour token budget.
"""

import logging
import os
import time
from typing import Any, Dict, Iterator, List, Optional, Tuple

import requests

LOG = logging.getLogger(__name__)

DEFAULT_API_ROOT = "https://api.github.com"
DEFAULT_REPO = "canonical/k8s-snap"
PER_PAGE = 100

# Leave headroom so a metrics run never starves other CI automation of quota.
RATE_LIMIT_FLOOR = 200

RETRY_STATUSES = (500, 502, 503, 504)
MAX_RETRIES = 4


class GitHubError(RuntimeError):
    """Raised for non-retryable API failures."""


class GitHubClient:
    """Minimal, rate-limit-aware GitHub Actions API client."""

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
            token or os.environ.get("GH_TOKEN") or os.environ.get("GITHUB_TOKEN")
        )
        self.session = session or requests.Session()
        # Tier-1 log fetching runs a thread pool over this session; urllib3's
        # default of 10 pooled connections thrashes above that.
        self.session.mount(
            "https://",
            requests.adapters.HTTPAdapter(pool_connections=32, pool_maxsize=32),
        )
        self.calls_made = 0

    # -- plumbing ---------------------------------------------------------

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
        """Sleep if we are close to exhausting the quota.

        Backing off here (rather than failing) keeps long backfills unattended:
        a 90-day backfill is ~20k calls and will cross at least one reset window.
        """
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
            "Rate limit nearly exhausted (%s remaining); sleeping %ss until reset",
            remaining_i,
            wait,
        )
        time.sleep(wait)

    def request(
        self, path: str, params: Optional[Dict[str, Any]] = None, raw: bool = False
    ) -> Tuple[int, Any]:
        """Perform a GET, retrying transient failures.

        Returns ``(status_code, payload)``. 404 is returned rather than raised:
        it is a meaningful, expected answer for job logs (see ``get_job_log``).
        """
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
                LOG.warning(
                    "Request error %s (attempt %s); retrying in %ss",
                    exc,
                    attempt + 1,
                    sleep_for,
                )
                time.sleep(sleep_for)
                continue

            self.calls_made += 1
            self._respect_rate_limit(response)

            if response.status_code in RETRY_STATUSES:
                sleep_for = 2**attempt
                LOG.warning(
                    "HTTP %s from %s (attempt %s); retrying in %ss",
                    response.status_code,
                    url,
                    attempt + 1,
                    sleep_for,
                )
                time.sleep(sleep_for)
                continue

            # Secondary rate limit / abuse detection.
            if response.status_code == 403 and "rate limit" in response.text.lower():
                sleep_for = 60 * (attempt + 1)
                LOG.warning("Secondary rate limit hit; sleeping %ss", sleep_for)
                time.sleep(sleep_for)
                continue

            if response.status_code == 404:
                return 404, None

            if response.status_code >= 400:
                raise GitHubError(
                    f"HTTP {response.status_code} from {url}: {response.text[:400]}"
                )

            if raw:
                return response.status_code, response.content
            return response.status_code, response.json()

        raise GitHubError(f"Exhausted retries for {url}: {last_exc}")

    def get(self, path: str, params: Optional[Dict[str, Any]] = None) -> Any:
        _, payload = self.request(path, params=params)
        return payload

    def paginate(
        self, path: str, key: str, params: Optional[Dict[str, Any]] = None
    ) -> Iterator[Dict[str, Any]]:
        """Yield every item of a paginated list endpoint.

        Uses ``total_count`` where the endpoint provides it (jobs, runs,
        artifacts all do) and falls back to an empty-page check otherwise.
        """
        params = dict(params or {})
        params.setdefault("per_page", PER_PAGE)
        page = 1
        seen = 0

        while True:
            params["page"] = page
            payload = self.get(path, params=params)
            if not payload:
                return
            items: List[Dict[str, Any]] = payload.get(key, [])
            if not items:
                return
            for item in items:
                yield item
            seen += len(items)
            total = payload.get("total_count")
            if total is not None and seen >= total:
                return
            page += 1

    # -- endpoints --------------------------------------------------------

    def get_run(self, run_id: int) -> Dict[str, Any]:
        return self.get(f"repos/{self.repo}/actions/runs/{run_id}")

    def list_workflows(self) -> List[Dict[str, Any]]:
        return list(self.paginate(f"repos/{self.repo}/actions/workflows", "workflows"))

    def list_workflow_runs(
        self,
        workflow_id: Any,
        created: Optional[str] = None,
        event: Optional[str] = None,
        limit: Optional[int] = None,
    ) -> List[Dict[str, Any]]:
        """List runs of a workflow, newest first.

        ``created`` accepts the GitHub date-range syntax, e.g.
        ``2026-07-01..2026-09-09`` -- verified working against the live API.
        """
        params: Dict[str, Any] = {}
        if created:
            params["created"] = created
        if event:
            params["event"] = event

        runs: List[Dict[str, Any]] = []
        path = f"repos/{self.repo}/actions/workflows/{workflow_id}/runs"
        for run in self.paginate(path, "workflow_runs", params=params):
            runs.append(run)
            if limit is not None and len(runs) >= limit:
                break
        return runs

    def list_run_jobs(
        self, run_id: int, attempt: Optional[int] = None
    ) -> List[Dict[str, Any]]:
        """List jobs for a run.

        With ``attempt`` set, returns that specific attempt's jobs (used for
        flake detection). Without it, returns only the latest attempt of each
        job via ``filter=latest``, which is what run-level accounting wants.
        """
        if attempt is not None:
            path = f"repos/{self.repo}/actions/runs/{run_id}/attempts/{attempt}/jobs"
            return list(self.paginate(path, "jobs"))
        path = f"repos/{self.repo}/actions/runs/{run_id}/jobs"
        return list(self.paginate(path, "jobs", params={"filter": "latest"}))

    def get_job_log(self, job_id: int) -> Optional[str]:
        """Fetch a job's log, or ``None`` when GitHub has no blob for it.

        A 404 here is signal, not an error: jobs whose runner vanished
        mid-execution have no log blob at all, which is precisely how the
        ``infra.runner/runner_lost`` class is detected.
        """
        status, payload = self.request(
            f"repos/{self.repo}/actions/jobs/{job_id}/logs", raw=True
        )
        if status == 404 or payload is None:
            return None
        if isinstance(payload, bytes):
            return payload.decode("utf-8", errors="replace")
        return str(payload)
