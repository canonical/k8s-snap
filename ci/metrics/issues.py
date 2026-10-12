#
# Copyright 2026 Canonical, Ltd.
#
"""Bridge classified CI signatures to GitHub issues."""

from __future__ import annotations

import dataclasses
import os
import re
import sqlite3
import subprocess
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Protocol, Sequence, Tuple

import requests

from .gh import DEFAULT_API_ROOT, DEFAULT_REPO
from .scrub import find_secrets, scrub

ISSUE_MARKER_RE = re.compile(
    r"<!-- ci-signature: (?P<signature>[0-9a-f]{16}) "
    r"ctx=(?P<substrate>[A-Za-z0-9_.-]+|\*) "
    r"key=(?P<key>[^ ]+) -->"
)
SHA_RE = re.compile(r"^[0-9a-f]{7,40}$", re.IGNORECASE)

ELIGIBLE_CLASSES = ("product.bug", "test.bug")
DEFAULT_MIN_OCCURRENCES = 3
DEFAULT_MIN_RUNS = 2
DEFAULT_DAILY_BUDGET = 3
DEFAULT_REOPEN_COMMENT_WINDOW_DAYS = 7
DEFAULT_RESOLVED_COOLDOWN_DAYS = 14


class IssueBridgeError(RuntimeError):
    """Raised when issue bridge safety or API checks fail."""


@dataclasses.dataclass(frozen=True)
class IssueMarker:
    signature_id: str
    substrate: str
    idempotency_key: str


@dataclasses.dataclass(frozen=True)
class Candidate:
    signature_id: str
    substrate: str
    failure_class: str
    rule_id: str
    confidence: float
    occurrences: int
    runs_seen: int
    tests_affected: int
    configs_affected: int
    first_seen: str
    last_seen: str
    exemplar_excerpt: str
    rule_notes: str
    flaky: bool
    run_urls: Tuple[str, ...]
    matrix_cells: Tuple[str, ...]
    test_selectors: Tuple[str, ...]
    source_paths: Tuple[str, ...]

    @property
    def idempotency_key(self) -> str:
        return idempotency_key(self.signature_id, self.substrate)

    @property
    def signature_label(self) -> str:
        return f"signature/{self.signature_id}"

    @property
    def substrate_label(self) -> str:
        return f"substrate/{self.substrate}"


@dataclasses.dataclass(frozen=True)
class BridgeResult:
    created: Tuple[int, ...] = ()
    adopted: Tuple[int, ...] = ()
    reopened: Tuple[int, ...] = ()
    commented: Tuple[int, ...] = ()
    skipped: Tuple[str, ...] = ()
    dry_run: Tuple[str, ...] = ()

    def as_dict(self) -> Dict[str, Any]:
        return dataclasses.asdict(self)


class IssueClient(Protocol):
    """Minimal GitHub Issues client used by the bridge and faked in tests."""

    def current_user(self) -> str:
        raise NotImplementedError

    def list_issues(
        self, labels: Sequence[str], state: str = "open"
    ) -> List[Dict[str, Any]]:
        raise NotImplementedError

    def create_issue(
        self, title: str, body: str, labels: Sequence[str]
    ) -> Dict[str, Any]:
        raise NotImplementedError

    def get_issue(self, number: int) -> Dict[str, Any]:
        raise NotImplementedError

    def edit_issue(self, number: int, **fields: Any) -> Dict[str, Any]:
        raise NotImplementedError

    def add_comment(self, number: int, body: str) -> Dict[str, Any]:
        raise NotImplementedError


class GitHubIssueClient:
    """Small write-capable Issues API client."""

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
            or os.environ.get("BOT_TOKEN")
            or os.environ.get("GITHUB_TOKEN")
        )
        if not self.token:
            raise IssueBridgeError(
                "metrics open-issues requires GH_TOKEN, BOT_TOKEN, or --token"
            )
        self.session = session or requests.Session()

    def _headers(self) -> Dict[str, str]:
        return {
            "Accept": "application/vnd.github+json",
            "Authorization": f"Bearer {self.token}",
            "X-GitHub-Api-Version": "2022-11-28",
            "User-Agent": "k8s-ci-metrics-issue-bridge",
        }

    def _request(
        self,
        method: str,
        path: str,
        *,
        params: Optional[Dict[str, Any]] = None,
        json: Optional[Dict[str, Any]] = None,
    ) -> Any:
        url = f"{self.api_root}/{path.lstrip('/')}"
        response = self.session.request(
            method,
            url,
            headers=self._headers(),
            params=params,
            json=json,
            timeout=60,
        )
        if response.status_code >= 400:
            raise IssueBridgeError(
                f"GitHub {method} {path} failed with HTTP {response.status_code}: "
                f"{response.text[:400]}"
            )
        return response.json() if response.content else None

    def current_user(self) -> str:
        payload = self._request("GET", "user")
        return str(payload["login"])

    def list_issues(
        self, labels: Sequence[str], state: str = "open"
    ) -> List[Dict[str, Any]]:
        payload = self._request(
            "GET",
            f"repos/{self.repo}/issues",
            params={"state": state, "labels": ",".join(labels), "per_page": 100},
        )
        return list(payload or [])

    def create_issue(
        self, title: str, body: str, labels: Sequence[str]
    ) -> Dict[str, Any]:
        return dict(
            self._request(
                "POST",
                f"repos/{self.repo}/issues",
                json={"title": title, "body": body, "labels": list(labels)},
            )
        )

    def get_issue(self, number: int) -> Dict[str, Any]:
        return dict(self._request("GET", f"repos/{self.repo}/issues/{number}"))

    def edit_issue(self, number: int, **fields: Any) -> Dict[str, Any]:
        return dict(
            self._request("PATCH", f"repos/{self.repo}/issues/{number}", json=fields)
        )

    def add_comment(self, number: int, body: str) -> Dict[str, Any]:
        return dict(
            self._request(
                "POST",
                f"repos/{self.repo}/issues/{number}/comments",
                json={"body": _safe_body(body)},
            )
        )


def idempotency_key(signature_id: str, substrate: str) -> str:
    return f"{signature_id}:{substrate}"


def issue_marker(signature_id: str, substrate: str) -> str:
    key = idempotency_key(signature_id, substrate)
    return f"<!-- ci-signature: {signature_id} ctx={substrate} key={key} -->"


def parse_marker(body: str) -> Optional[IssueMarker]:
    match = ISSUE_MARKER_RE.search(body or "")
    if match is None:
        return None
    return IssueMarker(
        signature_id=match.group("signature"),
        substrate=match.group("substrate"),
        idempotency_key=match.group("key"),
    )


def _safe_body(body: str) -> str:
    cleaned = scrub(body)
    findings = find_secrets(cleaned)
    if findings:
        first = findings[0]
        raise IssueBridgeError(
            f"refusing to post issue body with secret-like value at line "
            f"{first.line_number} ({first.rule_name})"
        )
    return cleaned


def _utc_now() -> datetime:
    return datetime.now(timezone.utc)


def _issue_labels(candidate: Candidate) -> Tuple[str, ...]:
    return (
        "ci-failure",
        candidate.signature_label,
        candidate.substrate_label,
        candidate.failure_class.replace(".", "-"),
        "kind/bug",
        "flaky" if candidate.flaky else "systemic",
    )


def _issue_title(candidate: Candidate) -> str:
    kind = "product bug" if candidate.failure_class == "product.bug" else "test bug"
    return f"CI {kind}: {candidate.signature_id} on {candidate.substrate}"


def _bullet_list(values: Sequence[str], limit: int = 8) -> str:
    if not values:
        return "- n/a"
    shown = list(values[:limit])
    extra = len(values) - len(shown)
    lines = [f"- {value}" for value in shown]
    if extra:
        lines.append(f"- ... and {extra} more")
    return "\n".join(lines)


def _issue_body(candidate: Candidate) -> str:
    lines = [
        issue_marker(candidate.signature_id, candidate.substrate),
        "",
        "## Summary",
        "",
        f"The CI metrics bridge found a recurring `{candidate.failure_class}` "
        f"signature in the `{candidate.substrate}` substrate.",
        "",
        "Opening this issue triggers the cheap triage classifier only. A "
        "maintainer must comment `/triage` or `/reproduce` before the "
        "self-hosted reproduction pipeline runs.",
        "",
        "## Evidence",
        "",
        f"- Signature: `{candidate.signature_id}`",
        f"- Substrate: `{candidate.substrate}`",
        f"- Class: `{candidate.failure_class}`",
        f"- Rule: `{candidate.rule_id}` (confidence {candidate.confidence:.2f})",
        f"- Rule notes: {candidate.rule_notes}",
        f"- First seen: {candidate.first_seen}",
        f"- Last seen: {candidate.last_seen}",
        f"- Occurrences: {candidate.occurrences}",
        f"- Runs seen: {candidate.runs_seen}",
        f"- Affected tests: {candidate.tests_affected}",
        f"- Affected matrix cells: {candidate.configs_affected}",
        f"- Retry outcome: {'flaky' if candidate.flaky else 'systemic'}",
        "",
        "## Matrix cells",
        "",
        _bullet_list(candidate.matrix_cells),
        "",
        "## Failing selectors",
        "",
        _bullet_list(candidate.test_selectors),
        "",
        "## Source paths",
        "",
        _bullet_list(candidate.source_paths),
        "",
        "## Runs",
        "",
        _bullet_list(candidate.run_urls),
        "",
        "## Scrubbed exemplar excerpt",
        "",
        "```text",
        candidate.exemplar_excerpt,
        "```",
    ]
    return _safe_body("\n".join(lines))


def _comment_body(candidate: Candidate, reason: str) -> str:
    lines = [
        f"CI metrics saw this signature again ({reason}).",
        "",
        f"- Signature: `{candidate.signature_id}`",
        f"- Substrate: `{candidate.substrate}`",
        f"- Occurrences now recorded: {candidate.occurrences}",
        f"- Runs seen: {candidate.runs_seen}",
        f"- Last seen: {candidate.last_seen}",
        "",
        "Newest runs:",
        _bullet_list(candidate.run_urls, limit=3),
    ]
    return _safe_body("\n".join(lines))


def _label_names(issue: Dict[str, Any]) -> Tuple[str, ...]:
    names = []
    for label in issue.get("labels", []):
        if isinstance(label, dict):
            names.append(str(label.get("name", "")))
        else:
            names.append(str(label))
    return tuple(name for name in names if name)


def _author(issue: Dict[str, Any]) -> str:
    user = issue.get("user") or {}
    return str(user.get("login", ""))


def _state_reason(issue: Dict[str, Any]) -> Optional[str]:
    value = issue.get("state_reason")
    return str(value) if value else None


def _parse_dt(value: Optional[str]) -> Optional[datetime]:
    if not value:
        return None
    normalised = value.replace("Z", "+00:00")
    return datetime.fromisoformat(normalised)


def _row_exists(conn: sqlite3.Connection, issue_number: int) -> bool:
    row = conn.execute(
        "SELECT 1 FROM issues WHERE issue_number = ?", (issue_number,)
    ).fetchone()
    return row is not None


def _recorded_issue(
    conn: sqlite3.Connection, signature_id: str, substrate: str
) -> Optional[sqlite3.Row]:
    conn.row_factory = sqlite3.Row
    return conn.execute(
        """
        SELECT * FROM issues
        WHERE signature_id = ? AND substrate = ?
        ORDER BY (issue_state IN ('claiming', 'open')) DESC, claimed_at DESC
        LIMIT 1
        """,
        (signature_id, substrate),
    ).fetchone()


def _claim_issue(conn: sqlite3.Connection, candidate: Candidate, now: datetime) -> bool:
    try:
        conn.execute(
            """
            INSERT INTO issues (
                signature_id, substrate, idempotency_key, issue_state, claimed_at,
                opened_by_ruleset_version
            ) VALUES (?, ?, ?, 'claiming', ?, (
                SELECT MAX(ruleset_version) FROM signature_classifications
                WHERE signature_id = ? AND substrate IN ('*', ?)
            ))
            """,
            (
                candidate.signature_id,
                candidate.substrate,
                candidate.idempotency_key,
                now.isoformat(),
                candidate.signature_id,
                candidate.substrate,
            ),
        )
        return True
    except sqlite3.IntegrityError:
        return False


def _adopt_issue(
    conn: sqlite3.Connection,
    issue: Dict[str, Any],
    marker: IssueMarker,
    *,
    opened_at: Optional[str] = None,
) -> None:
    state = str(issue.get("state", "open"))
    if state not in ("open", "closed"):
        state = "open"
    issue_state = "open" if state == "open" else "closed"
    closed_at = issue.get("closed_at") if issue_state == "closed" else None
    closed_reason = _state_reason(issue) if issue_state == "closed" else None
    timestamp = _utc_now().isoformat()
    existing = _recorded_issue(conn, marker.signature_id, marker.substrate)
    if existing is not None and existing["issue_state"] == "claiming":
        conn.execute(
            """
            UPDATE issues
            SET issue_number = ?, issue_state = ?, opened_at = ?, closed_at = ?,
                closed_reason = ?
            WHERE claim_id = ?
            """,
            (
                int(issue["number"]),
                issue_state,
                opened_at or issue.get("created_at") or timestamp,
                closed_at,
                closed_reason,
                existing["claim_id"],
            ),
        )
        return
    conn.execute(
        """
        INSERT OR IGNORE INTO issues (
            signature_id, substrate, idempotency_key, issue_number, issue_state,
            claimed_at, opened_at, closed_at, closed_reason
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        """,
        (
            marker.signature_id,
            marker.substrate,
            marker.idempotency_key,
            int(issue["number"]),
            issue_state,
            timestamp,
            opened_at or issue.get("created_at") or timestamp,
            closed_at,
            closed_reason,
        ),
    )


def _verified_marker(issue: Dict[str, Any]) -> Optional[IssueMarker]:
    marker = parse_marker(str(issue.get("body") or ""))
    if marker is None:
        return None
    if marker.idempotency_key != idempotency_key(marker.signature_id, marker.substrate):
        return None
    return marker


def _verify_adoptable(
    issue: Dict[str, Any], expected_author: str, substrate: Optional[str] = None
) -> Optional[IssueMarker]:
    marker = _verified_marker(issue)
    if marker is None:
        return None
    if expected_author and _author(issue) != expected_author:
        return None
    if substrate is not None and marker.substrate != substrate:
        return None
    return marker


def reconcile_live_issues(
    conn: sqlite3.Connection,
    client: IssueClient,
    issue_author: str,
) -> Tuple[int, ...]:
    """Adopt live CI issues with markers that are missing from local storage."""

    adopted = []
    for issue in client.list_issues(["ci-failure"], state="open"):
        marker = _verify_adoptable(issue, issue_author)
        if marker is None or _row_exists(conn, int(issue["number"])):
            continue
        _adopt_issue(conn, issue, marker)
        adopted.append(int(issue["number"]))
    return tuple(adopted)


def _lookup_issue(
    conn: sqlite3.Connection,
    client: IssueClient,
    candidate: Candidate,
    issue_author: str,
) -> Optional[Dict[str, Any]]:
    labels = [candidate.signature_label, candidate.substrate_label]
    for issue in client.list_issues(labels, state="all"):
        marker = _verify_adoptable(issue, issue_author, candidate.substrate)
        if marker is None or marker.signature_id != candidate.signature_id:
            continue
        if not _row_exists(conn, int(issue["number"])):
            _adopt_issue(conn, issue, marker)
        return issue
    return None


def _ensure_labels(
    client: IssueClient, issue_number: int, issue: Dict[str, Any], labels: Sequence[str]
) -> None:
    present = set(_label_names(issue))
    missing = [label for label in labels if label not in present]
    if missing:
        client.edit_issue(issue_number, labels=sorted(present.union(labels)))


def _today_opened_count(conn: sqlite3.Connection, now: datetime) -> int:
    day = now.date().isoformat()
    row = conn.execute(
        """
        SELECT COUNT(*) FROM issues
        WHERE opened_at IS NOT NULL AND substr(opened_at, 1, 10) = ?
        """,
        (day,),
    ).fetchone()
    return int(row[0])


def _resolved_in_cooldown(row: sqlite3.Row, now: datetime, cooldown_days: int) -> bool:
    if row["issue_state"] != "closed" or row["closed_reason"] != "completed":
        return False
    closed_at = _parse_dt(row["closed_at"])
    if closed_at is None:
        return False
    return closed_at >= now - timedelta(days=cooldown_days)


def _candidate_rows(
    conn: sqlite3.Connection,
    *,
    min_occurrences: int,
    min_runs: int,
    rule_notes: Optional[Dict[str, str]] = None,
) -> Iterable[Candidate]:
    rows = conn.execute(
        """
        WITH job_ctx AS (
            SELECT vjc.signature_id, cfg.substrate, vjc.failure_class,
                   s.exemplar_excerpt, j.job_id, j.run_id, j.test_id, j.config_id,
                   j.html_url, j.job_name, r.started_at, t.nodeid,
                   COALESCE(t.file, '') AS source_path
            FROM v_job_class vjc
            JOIN jobs j USING (job_id)
            JOIN configs cfg USING (config_id)
            JOIN runs r ON r.run_id = j.run_id AND r.attempt = j.attempt
            JOIN tests t USING (test_id)
            JOIN signatures s USING (signature_id)
            WHERE vjc.classified_by = 'rules'
              AND vjc.failure_class IN ('product.bug', 'test.bug')
        )
        SELECT signature_id, substrate, failure_class, exemplar_excerpt,
               COUNT(*) AS occurrences, COUNT(DISTINCT run_id) AS runs_seen,
               COUNT(DISTINCT test_id) AS tests_affected,
               COUNT(DISTINCT config_id) AS configs_affected,
               MIN(started_at) AS first_seen, MAX(started_at) AS last_seen,
               GROUP_CONCAT(DISTINCT html_url) AS run_urls,
               GROUP_CONCAT(DISTINCT job_name) AS matrix_cells,
               GROUP_CONCAT(DISTINCT nodeid) AS test_selectors,
               GROUP_CONCAT(DISTINCT source_path) AS source_paths,
               EXISTS (
                   SELECT 1 FROM v_flakes vf
                   WHERE vf.signature_id = job_ctx.signature_id
               ) AS flaky
        FROM job_ctx
        GROUP BY signature_id, substrate, failure_class, exemplar_excerpt
        HAVING occurrences >= ? AND runs_seen >= ?
        ORDER BY last_seen DESC, signature_id, substrate
        """,
        (min_occurrences, min_runs),
    )
    for row in rows:
        class_row = conn.execute(
            """
            SELECT sc.rule_id, sc.confidence
            FROM v_signature_class sc
            WHERE sc.signature_id = ? AND sc.substrate IN ('*', ?)
            ORDER BY (sc.substrate <> '*') DESC
            LIMIT 1
            """,
            (row[0], row[1]),
        ).fetchone()
        if class_row is None:
            continue
        rule_id = str(class_row[0] or "unknown")
        yield Candidate(
            signature_id=row[0],
            substrate=row[1],
            failure_class=row[2],
            exemplar_excerpt=row[3],
            occurrences=int(row[4]),
            runs_seen=int(row[5]),
            tests_affected=int(row[6]),
            configs_affected=int(row[7]),
            first_seen=str(row[8]),
            last_seen=str(row[9]),
            run_urls=tuple(filter(None, str(row[10] or "").split(","))),
            matrix_cells=tuple(filter(None, str(row[11] or "").split(","))),
            test_selectors=tuple(filter(None, str(row[12] or "").split(","))),
            source_paths=tuple(filter(None, str(row[13] or "").split(","))),
            flaky=bool(row[14]),
            rule_id=rule_id,
            confidence=float(class_row[1] if class_row[1] is not None else 0.0),
            rule_notes=(rule_notes or {}).get(rule_id, ""),
        )


def _candidate_tested_commits(
    conn: sqlite3.Connection, candidate: Candidate
) -> Tuple[Tuple[str, ...], bool]:
    rows = conn.execute(
        """
        SELECT cfg.channel, cfg.arch, j.tested_revision
        FROM jobs j
        JOIN configs cfg USING (config_id)
        JOIN v_job_class vjc USING (job_id)
        WHERE j.signature_id = ? AND cfg.substrate = ?
          AND vjc.failure_class = ? AND vjc.classified_by = 'rules'
        """,
        (candidate.signature_id, candidate.substrate, candidate.failure_class),
    )
    commits = []
    unknown = False
    for channel, arch, revision in rows:
        if not channel or not arch or not revision:
            unknown = True
            continue
        provenance = conn.execute(
            """
            SELECT built_sha FROM snap_provenance
            WHERE channel = ? AND arch = ? AND revision = ?
            """,
            (channel, arch, int(revision)),
        ).fetchone()
        if provenance is None:
            unknown = True
        else:
            commits.append(str(provenance[0]))
    return tuple(commits), unknown


def _is_ancestor(repo: Path, ancestor: str, descendant: str) -> bool:
    if not SHA_RE.match(ancestor) or not SHA_RE.match(descendant):
        raise IssueBridgeError("freshness commits must look like git SHAs")
    result = subprocess.run(  # noqa: S603
        ["git", "-C", str(repo), "merge-base", "--is-ancestor", ancestor, descendant],
        check=False,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        timeout=30,
    )
    if result.returncode == 0:
        return True
    if result.returncode == 1:
        return False
    raise IssueBridgeError(
        f"git merge-base --is-ancestor failed with {result.returncode}"
    )


def _freshness(
    conn: sqlite3.Connection,
    candidate: Candidate,
    fixed_at_sha: Optional[str],
    repo: Path,
) -> str:
    if not fixed_at_sha:
        return "unknown"
    commits, unknown = _candidate_tested_commits(conn, candidate)
    if unknown or not commits:
        return "unknown"
    if any(_is_ancestor(repo, fixed_at_sha, commit) for commit in commits):
        return "fresh"
    return "stale"


def _comment_window_elapsed(row: sqlite3.Row, now: datetime, window_days: int) -> bool:
    last = _parse_dt(row["last_comment_at"])
    return last is None or last <= now - timedelta(days=window_days)


def open_issues(
    conn: sqlite3.Connection,
    client: IssueClient,
    *,
    dry_run: bool = True,
    issue_author: Optional[str] = None,
    repo_path: Path = Path("."),
    min_occurrences: int = DEFAULT_MIN_OCCURRENCES,
    min_runs: int = DEFAULT_MIN_RUNS,
    daily_budget: int = DEFAULT_DAILY_BUDGET,
    resolved_cooldown_days: int = DEFAULT_RESOLVED_COOLDOWN_DAYS,
    reopen_comment_window_days: int = DEFAULT_REOPEN_COMMENT_WINDOW_DAYS,
    rule_notes: Optional[Dict[str, str]] = None,
) -> BridgeResult:
    """Open or update GitHub issues for eligible signature/substrate pairs."""

    conn.row_factory = sqlite3.Row
    now = _utc_now()
    author = issue_author or ("" if dry_run else client.current_user())
    created: List[int] = []
    adopted: List[int] = []
    reopened: List[int] = []
    commented: List[int] = []
    skipped: List[str] = []
    dry_run_actions: List[str] = []

    if not dry_run:
        adopted.extend(reconcile_live_issues(conn, client, author))

    budget_remaining = max(0, daily_budget - _today_opened_count(conn, now))

    for candidate in _candidate_rows(
        conn,
        min_occurrences=min_occurrences,
        min_runs=min_runs,
        rule_notes=rule_notes,
    ):
        row = _recorded_issue(conn, candidate.signature_id, candidate.substrate)
        if row is not None and row["issue_state"] in ("open", "claiming"):
            skipped.append(f"{candidate.idempotency_key}: already tracked")
            continue
        if row is not None and _resolved_in_cooldown(row, now, resolved_cooldown_days):
            skipped.append(f"{candidate.idempotency_key}: closed in cooldown")
            continue
        if row is not None and row["issue_state"] == "closed":
            freshness = _freshness(
                conn, candidate, row["fixed_at_sha"], Path(repo_path).resolve()
            )
            if freshness == "stale":
                skipped.append(f"{candidate.idempotency_key}: stale")
                continue
            if freshness == "unknown":
                if _comment_window_elapsed(row, now, reopen_comment_window_days):
                    if dry_run:
                        dry_run_actions.append(
                            f"would comment on closed {row['issue_number']} "
                            f"for unknown freshness"
                        )
                    else:
                        client.add_comment(
                            int(row["issue_number"]),
                            _comment_body(candidate, "freshness unknown"),
                        )
                        conn.execute(
                            "UPDATE issues SET last_comment_at = ? WHERE claim_id = ?",
                            (now.isoformat(), row["claim_id"]),
                        )
                        commented.append(int(row["issue_number"]))
                skipped.append(f"{candidate.idempotency_key}: freshness unknown")
                continue
            if dry_run:
                dry_run_actions.append(f"would reopen issue {row['issue_number']}")
                continue
            number = int(row["issue_number"])
            client.edit_issue(number, state="open")
            client.add_comment(number, _comment_body(candidate, "provably fresh"))
            conn.execute(
                """
                UPDATE issues
                SET issue_state = 'open', closed_at = NULL, closed_reason = NULL,
                    reopened_count = reopened_count + 1, last_comment_at = ?
                WHERE claim_id = ?
                """,
                (now.isoformat(), row["claim_id"]),
            )
            reopened.append(number)
            continue

        if budget_remaining <= 0:
            skipped.append(f"{candidate.idempotency_key}: daily budget exhausted")
            continue

        if dry_run:
            dry_run_actions.append(f"would create {candidate.idempotency_key}")
            budget_remaining -= 1
            continue

        if not _claim_issue(conn, candidate, now):
            skipped.append(f"{candidate.idempotency_key}: claim conflict")
            continue
        adopted_now = reconcile_live_issues(conn, client, author)
        adopted.extend(adopted_now)
        live_issue = _lookup_issue(conn, client, candidate, author)
        if live_issue is not None:
            number = int(live_issue["number"])
            if number not in adopted:
                skipped.append(f"{candidate.idempotency_key}: already live")
            continue

        labels = _issue_labels(candidate)
        issue = client.create_issue(
            _issue_title(candidate), _issue_body(candidate), labels
        )
        number = int(issue["number"])
        confirmed = client.get_issue(number)
        marker = _verify_adoptable(confirmed, author, candidate.substrate)
        if marker is None or marker.signature_id != candidate.signature_id:
            raise IssueBridgeError(f"created issue {number} does not carry marker")
        _ensure_labels(client, number, confirmed, labels)
        _adopt_issue(conn, confirmed, marker, opened_at=now.isoformat())
        conn.execute(
            """
            UPDATE issues
            SET occurrences_at_open = ?, last_reported_occurrences = ?,
                post_attempts = post_attempts + 1
            WHERE issue_number = ?
            """,
            (candidate.occurrences, candidate.occurrences, number),
        )
        created.append(number)
        budget_remaining -= 1

    return BridgeResult(
        created=tuple(created),
        adopted=tuple(dict.fromkeys(adopted)),
        reopened=tuple(reopened),
        commented=tuple(commented),
        skipped=tuple(skipped),
        dry_run=tuple(dry_run_actions),
    )
