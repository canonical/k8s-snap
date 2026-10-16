#
# Copyright 2026 Canonical, Ltd.
#
"""Signature-to-issue bridge safety tests."""

from __future__ import annotations

import sqlite3
import sys
import tempfile
import unittest
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Any, Dict, List, Optional, Sequence
from unittest.mock import ANY, patch

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.issues import (  # noqa: E402
    QUIET_CLOSE_REASON,
    GitHubIssueClient,
    IssueBridgeError,
    _closer_sha,
    issue_marker,
    open_issues,
    sync_issue_states,
)
from metrics.models import SCHEMA_SQL  # noqa: E402
from metrics.scrub import SecretFinding  # noqa: E402

QUIET_CLOSED_AT = "2026-11-01T00:00:00Z"


def _db() -> sqlite3.Connection:
    conn = sqlite3.connect(":memory:")
    conn.execute("PRAGMA foreign_keys = ON")
    conn.executescript(SCHEMA_SQL)
    conn.execute(
        "INSERT INTO workflows (workflow_id, name, file) VALUES (1, 'nightly', 'nightly.yaml')"
    )
    return conn


def _insert_candidate(
    conn: sqlite3.Connection,
    signature_id: str,
    substrate: str,
    failure_class: str = "product.bug",
    *,
    start_job: int = 1,
    excerpt: str = "safe excerpt",
) -> None:
    conn.execute(
        """
        INSERT OR IGNORE INTO signatures (
            signature_id, normalised_pattern, exemplar_excerpt,
            normaliser_version, extractor_version
        ) VALUES (?, 'pattern', ?, 1, 1)
        """,
        (signature_id, excerpt),
    )
    conn.execute(
        """
        INSERT INTO signature_classifications (
            signature_id, substrate, ruleset_version, failure_class, rule_id,
            confidence, classified_by, classified_at
        ) VALUES (?, ?, 1, ?, ?, 1.0, 'rules', '2026-10-12T00:00:00Z')
        """,
        (signature_id, substrate, failure_class, f"rule-{substrate}"),
    )
    for offset, run_id in enumerate((100, 200, 200), start=0):
        job_id = start_job + offset
        config_id = start_job + offset
        test_id = start_job + offset
        conn.execute(
            """
            INSERT OR IGNORE INTO runs (
                run_id, attempt, workflow_id, completed_at, conclusion, jobs_total,
                jobs_success, jobs_failure, jobs_skipped, jobs_cancelled,
                runner_minutes_total, runner_minutes_failed, ingest_version
            ) VALUES (?, 1, 1, '2026-10-12T00:00:00Z', 'failure', 1, 0, 1, 0, 0, 1.0, 1.0, 1)
            """,
            (run_id,),
        )
        conn.execute(
            """
            INSERT INTO configs (config_id, os, arch, channel, flavor, substrate)
            VALUES (?, 'noble', ?, '1.37-classic/edge', 'classic', ?)
            """,
            (config_id, f"{job_id}", substrate),
        )
        conn.execute(
            "INSERT INTO tests (test_id, nodeid, file) VALUES (?, ?, ?)",
            (
                test_id,
                f"tests/test_{substrate}.py::test_{job_id}",
                f"tests/test_{substrate}.py",
            ),
        )
        conn.execute(
            """
            INSERT INTO jobs (
                job_id, run_id, attempt, config_id, test_id, job_name, html_url,
                tested_revision, signature_id
            ) VALUES (?, ?, 1, ?, ?, ?, ?, '42', ?)
            """,
            (
                job_id,
                run_id,
                config_id,
                test_id,
                f"{substrate} job",
                f"https://example.invalid/runs/{run_id}/jobs/{job_id}",
                signature_id,
            ),
        )


def _issue(
    number: int, signature_id: str, substrate: str, user: str = "ci-bot"
) -> Dict[str, Any]:
    key = f"{signature_id}:{substrate}"
    return {
        "number": number,
        "state": "open",
        "state_reason": None,
        "created_at": "2026-10-12T00:00:00Z",
        "closed_at": None,
        "user": {"login": user},
        "labels": [
            {"name": f"signature/{signature_id}"},
            {"name": f"substrate/{substrate}"},
        ],
        "body": f"<!-- ci-signature: {signature_id} ctx={substrate} key={key} -->",
    }


def _insert_opened_issue(
    conn: sqlite3.Connection,
    *,
    signature_id: str,
    substrate: str = "lxd",
    issue_number: int,
    opened_at: str,
) -> None:
    conn.execute(
        """
        INSERT OR IGNORE INTO signatures (
            signature_id, normalised_pattern, exemplar_excerpt,
            normaliser_version, extractor_version
        ) VALUES (?, 'pattern', 'safe excerpt', 1, 1)
        """,
        (signature_id,),
    )
    conn.execute(
        """
        INSERT INTO issues (
            signature_id, substrate, idempotency_key, issue_number, issue_state,
            claimed_at, opened_at
        ) VALUES (?, ?, ?, ?, 'open', ?, ?)
        """,
        (
            signature_id,
            substrate,
            f"{signature_id}:{substrate}",
            issue_number,
            opened_at,
            opened_at,
        ),
    )


class FakeIssueClient:
    def __init__(self) -> None:
        self.sweep_issues: List[Dict[str, Any]] = []
        self.lookup_issues: List[Dict[str, Any]] = []
        self.created: List[Dict[str, Any]] = []
        self.edits: List[tuple[int, Dict[str, Any]]] = []
        self.comments: List[tuple[int, str]] = []
        self.closers: Dict[int, str] = {}
        self.closer_calls: List[int] = []
        self._next = 1000

    def current_user(self) -> str:
        return "ci-bot"

    def list_issues(
        self,
        labels: Sequence[str],
        state: str = "open",
        creator: Optional[str] = None,
    ) -> List[Dict[str, Any]]:
        if creator:
            pool = {
                issue["number"]: issue
                for issue in self.sweep_issues + self.lookup_issues + self.created
            }
            return [
                dict(issue)
                for issue in pool.values()
                if issue["user"]["login"] == creator
                and state in ("all", issue["state"])
                and set(labels).issubset({label["name"] for label in issue["labels"]})
            ]
        if list(labels) == ["ci-failure"]:
            return list(self.sweep_issues)
        required = set(labels)
        return [
            issue
            for issue in self.lookup_issues
            if required.issubset({label["name"] for label in issue["labels"]})
        ]

    def create_issue(
        self, title: str, body: str, labels: Sequence[str]
    ) -> Dict[str, Any]:
        issue = {
            "number": self._next,
            "state": "open",
            "state_reason": None,
            "created_at": "2026-10-12T00:00:00Z",
            "closed_at": None,
            "user": {"login": "ci-bot"},
            "labels": [],
            "title": title,
            "body": body,
            "requested_labels": list(labels),
        }
        self._next += 1
        self.created.append(issue)
        return dict(issue)

    def get_issue(self, number: int) -> Dict[str, Any]:
        for issue in self.created + self.lookup_issues + self.sweep_issues:
            if issue["number"] == number:
                return dict(issue)
        raise IssueBridgeError(f"GitHub GET issues/{number} failed with HTTP 404")

    def closing_commit(self, number: int) -> Optional[str]:
        self.closer_calls.append(number)
        closer = self.closers.get(number)
        if isinstance(closer, Exception):
            raise closer
        return closer

    def edit_issue(self, number: int, **fields: Any) -> Dict[str, Any]:
        self.edits.append((number, fields))
        for issue in self.created:
            if issue["number"] == number:
                if "labels" in fields:
                    issue["labels"] = [{"name": label} for label in fields["labels"]]
                issue.update(
                    {key: value for key, value in fields.items() if key != "labels"}
                )
                return dict(issue)
        for issue in self.lookup_issues:
            if issue["number"] == number and "state" in fields:
                closing = fields["state"] == "closed"
                issue.update(
                    {
                        "state": fields["state"],
                        "state_reason": fields.get("state_reason"),
                        "closed_at": QUIET_CLOSED_AT if closing else None,
                    }
                )
                return dict(issue)
        return {"number": number, **fields}

    def add_comment(self, number: int, body: str) -> Dict[str, Any]:
        self.comments.append((number, body))
        return {"id": len(self.comments)}


class IssueBridgeTest(unittest.TestCase):
    def test_lookup_present_without_local_record_is_adopted_not_skipped(self) -> None:
        conn = _db()
        _insert_candidate(conn, "aaaaaaaaaaaaaaaa", "lxd")
        client = FakeIssueClient()
        client.lookup_issues.append(_issue(42, "aaaaaaaaaaaaaaaa", "lxd"))

        with conn:
            result = open_issues(conn, client, dry_run=False, issue_author="ci-bot")

        row = conn.execute(
            """
            SELECT issue_number, issue_state FROM issues
            WHERE signature_id = 'aaaaaaaaaaaaaaaa' AND substrate = 'lxd'
            """
        ).fetchone()
        self.assertEqual(tuple(row), (42, "open"))
        self.assertEqual(result.created, ())
        self.assertEqual(client.created, [])

    def test_unknown_freshness_never_reopens_closed_issue(self) -> None:
        conn = _db()
        _insert_candidate(conn, "bbbbbbbbbbbbbbbb", "lxd")
        conn.execute(
            """
            INSERT INTO issues (
                signature_id, substrate, idempotency_key, issue_number, issue_state,
                claimed_at, opened_at, closed_at, closed_reason, fixed_at_sha,
                last_comment_at
            ) VALUES (
                'bbbbbbbbbbbbbbbb', 'lxd', 'bbbbbbbbbbbbbbbb:lxd', 77, 'closed',
                '2026-10-01T00:00:00Z', '2026-10-01T00:00:00Z',
                '2026-10-11T00:00:00Z', 'not_planned',
                '1111111111111111111111111111111111111111',
                '2999-01-01T00:00:00+00:00'
            )
            """
        )
        client = FakeIssueClient()

        with tempfile.TemporaryDirectory() as tmp, conn:
            result = open_issues(
                conn,
                client,
                dry_run=False,
                issue_author="ci-bot",
                repo_path=Path(tmp),
            )

        self.assertEqual(result.reopened, ())
        self.assertEqual(client.edits, [])
        self.assertEqual(client.comments, [])
        row = conn.execute(
            "SELECT issue_state FROM issues WHERE issue_number = 77"
        ).fetchone()
        self.assertEqual(tuple(row), ("closed",))

    def test_daily_budget_caps_new_issues_at_three(self) -> None:
        conn = _db()
        for index, signature_id in enumerate(
            (
                "1111111111111111",
                "2222222222222222",
                "3333333333333333",
                "4444444444444444",
            ),
            start=1,
        ):
            _insert_candidate(conn, signature_id, "lxd", start_job=index * 10)
        client = FakeIssueClient()

        with conn:
            result = open_issues(
                conn,
                client,
                dry_run=False,
                issue_author="ci-bot",
                daily_budget=3,
            )

        self.assertEqual(len(result.created), 3)
        self.assertEqual(len(client.created), 3)
        self.assertIn("daily budget exhausted", "\n".join(result.skipped))

    def test_daily_budget_already_consumed_today_blocks_new_invocation(self) -> None:
        conn = _db()
        _insert_candidate(conn, "1212121212121212", "lxd")
        for index, signature_id in enumerate(
            ("9000000000000001", "9000000000000002", "9000000000000003"),
            start=1,
        ):
            _insert_opened_issue(
                conn,
                signature_id=signature_id,
                issue_number=900 + index,
                opened_at="2026-10-12T00:00:00+00:00",
            )
        client = FakeIssueClient()

        with patch(
            "metrics.issues._utc_now",
            return_value=datetime(2026, 10, 12, 12, 0, tzinfo=timezone.utc),
        ):
            with conn:
                result = open_issues(
                    conn,
                    client,
                    dry_run=False,
                    issue_author="ci-bot",
                    daily_budget=3,
                )

        self.assertEqual(result.created, ())
        self.assertEqual(client.created, [])
        self.assertIn("daily budget exhausted", "\n".join(result.skipped))

    def test_daily_budget_partially_consumed_allows_remaining_capacity(self) -> None:
        conn = _db()
        _insert_candidate(conn, "1313131313131313", "lxd", start_job=10)
        _insert_candidate(conn, "1414141414141414", "lxd", start_job=20)
        for index, signature_id in enumerate(
            ("9000000000000011", "9000000000000012"),
            start=1,
        ):
            _insert_opened_issue(
                conn,
                signature_id=signature_id,
                issue_number=910 + index,
                opened_at="2026-10-12T08:30:00+00:00",
            )
        client = FakeIssueClient()

        with patch(
            "metrics.issues._utc_now",
            return_value=datetime(2026, 10, 12, 12, 0, tzinfo=timezone.utc),
        ):
            with conn:
                result = open_issues(
                    conn,
                    client,
                    dry_run=False,
                    issue_author="ci-bot",
                    daily_budget=3,
                )

        self.assertEqual(len(result.created), 1)
        self.assertEqual(len(client.created), 1)
        self.assertIn("daily budget exhausted", "\n".join(result.skipped))

    def test_daily_budget_ignores_yesterday_utc_rows(self) -> None:
        conn = _db()
        _insert_candidate(conn, "1515151515151515", "lxd")
        for index, signature_id in enumerate(
            ("9000000000000021", "9000000000000022", "9000000000000023"),
            start=1,
        ):
            _insert_opened_issue(
                conn,
                signature_id=signature_id,
                issue_number=920 + index,
                opened_at="2026-10-11T23:59:59+00:00",
            )
        client = FakeIssueClient()

        with patch(
            "metrics.issues._utc_now",
            return_value=datetime(2026, 10, 12, 0, 0, tzinfo=timezone.utc),
        ):
            with conn:
                result = open_issues(
                    conn,
                    client,
                    dry_run=False,
                    issue_author="ci-bot",
                    daily_budget=3,
                )

        self.assertEqual(len(result.created), 1)
        self.assertEqual(len(client.created), 1)

    def test_issue_body_scrubs_join_token_before_posting(self) -> None:
        conn = _db()
        _insert_candidate(
            conn,
            "dddddddddddddddd",
            "lxd",
            excerpt="k8s join failed with token abcdef.0123456789abcdef on worker",
        )
        client = FakeIssueClient()

        with conn:
            open_issues(conn, client, dry_run=False, issue_author="ci-bot")

        self.assertEqual(len(client.created), 1)
        body = client.created[0]["body"]
        self.assertIn("k8s join failed with token <redacted> on worker", body)
        self.assertNotIn("abcdef.0123456789abcdef", body)

    def test_detected_secret_blocks_post_and_comment_calls(self) -> None:
        conn = _db()
        _insert_candidate(conn, "eeeeeeeeeeeeeeee", "lxd")
        client = FakeIssueClient()

        with patch(
            "metrics.issues.find_secrets",
            return_value=(SecretFinding(1, "forced-secret"),),
        ):
            with self.assertRaises(IssueBridgeError):
                with conn:
                    open_issues(conn, client, dry_run=False, issue_author="ci-bot")

        self.assertEqual(client.created, [])
        self.assertEqual(client.comments, [])

    def test_body_uses_scrubbed_text_not_raw_assignment_secret(self) -> None:
        conn = _db()
        raw = "token=abc123def456"
        _insert_candidate(
            conn,
            "ffffffffffffffff",
            "lxd",
            excerpt=f"setup failed after {raw} was printed",
        )
        client = FakeIssueClient()

        with conn:
            open_issues(conn, client, dry_run=False, issue_author="ci-bot")

        self.assertEqual(len(client.created), 1)
        body = client.created[0]["body"]
        self.assertIn("setup failed after token=<redacted> was printed", body)
        self.assertNotIn(raw, body)

    def test_eligibility_is_keyed_by_signature_and_substrate(self) -> None:
        conn = _db()
        _insert_candidate(
            conn,
            "cccccccccccccccc",
            "multipass",
            failure_class="infra.provisioning",
            start_job=1,
        )
        _insert_candidate(
            conn,
            "cccccccccccccccc",
            "lxd",
            failure_class="product.bug",
            start_job=10,
        )
        client = FakeIssueClient()

        with conn:
            result = open_issues(conn, client, dry_run=False, issue_author="ci-bot")

        self.assertEqual(len(result.created), 1)
        self.assertIn("substrate/lxd", client.created[0]["requested_labels"])
        self.assertNotIn("substrate/multipass", client.created[0]["requested_labels"])
        self.assertIn("- Occurrences: 3", client.created[0]["body"])
        self.assertNotIn("multipass job", client.created[0]["body"])


class _PagedResponse:
    def __init__(self, payload: List[Dict[str, Any]]) -> None:
        self.status_code = 200
        self.content = b"x"
        self.text = ""
        self._payload = payload

    def json(self) -> List[Dict[str, Any]]:
        return self._payload


class _PagedSession:
    def __init__(self, pages: List[List[Dict[str, Any]]]) -> None:
        self.pages = pages
        self.params: List[Dict[str, Any]] = []

    def request(self, method: str, url: str, **kwargs: Any) -> _PagedResponse:
        params = dict(kwargs["params"])
        self.params.append(params)
        return _PagedResponse(self.pages[params["page"] - 1])


class IssueBridgeIdentityTest(unittest.TestCase):
    def test_list_issues_reads_every_page_and_drops_pull_requests(self) -> None:
        full_page = [{"number": n} for n in range(100)]
        last_page = [{"number": 100}, {"number": 101, "pull_request": {}}]
        session = _PagedSession([full_page, last_page])
        client = GitHubIssueClient(repo="o/r", token="t", session=session)

        issues = client.list_issues([], state="all", creator="ci-bot")

        self.assertEqual([i["number"] for i in issues], list(range(101)))
        self.assertEqual([p["page"] for p in session.params], [1, 2])
        self.assertEqual(session.params[0]["creator"], "ci-bot")
        self.assertNotIn("labels", session.params[0])

    def test_unlabelled_bot_issue_is_found_by_marker_not_duplicated(self) -> None:
        conn = _db()
        _insert_candidate(conn, "dddddddddddddddd", "lxd")
        client = FakeIssueClient()
        unlabelled = _issue(77, "dddddddddddddddd", "lxd")
        unlabelled["labels"] = []
        unlabelled["state"] = "closed"
        client.lookup_issues.append(unlabelled)

        with conn:
            result = open_issues(conn, client, dry_run=False, issue_author="ci-bot")

        self.assertEqual(client.created, [])
        self.assertEqual(
            tuple(
                conn.execute("SELECT issue_number, issue_state FROM issues").fetchone()
            ),
            (77, "closed"),
        )
        self.assertIn("dddddddddddddddd:lxd: already live", result.skipped)

    def test_sweep_adopts_open_bot_issue_whose_labels_were_dropped(self) -> None:
        conn = _db()
        client = FakeIssueClient()
        unlabelled = _issue(78, "eeeeeeeeeeeeeeee", "lxd")
        unlabelled["labels"] = []
        client.lookup_issues.append(unlabelled)
        conn.execute(
            """
            INSERT INTO signatures (
                signature_id, normalised_pattern, exemplar_excerpt,
                normaliser_version, extractor_version
            ) VALUES ('eeeeeeeeeeeeeeee', 'p', 'e', 1, 1)
            """
        )

        with conn:
            result = open_issues(conn, client, dry_run=False, issue_author="ci-bot")

        self.assertEqual(result.adopted, (78,))

    def test_empty_substrate_candidate_is_skipped_before_claim(self) -> None:
        conn = _db()
        _insert_candidate(conn, "ffffffffffffffff", "")
        client = FakeIssueClient()

        with conn:
            result = open_issues(conn, client, dry_run=False, issue_author="ci-bot")

        self.assertEqual(client.created, [])
        self.assertIn("ffffffffffffffff:: substrate not markable", result.skipped)
        self.assertEqual(conn.execute("SELECT COUNT(*) FROM issues").fetchone()[0], 0)
        with self.assertRaises(IssueBridgeError):
            issue_marker("ffffffffffffffff", "")

    def test_matrix_cells_and_selectors_with_commas_render_intact(self) -> None:
        conn = _db()
        _insert_candidate(conn, "1111111111111111", "lxd")
        conn.execute(
            "UPDATE jobs SET job_name = 'Integration (noble, amd64) / ' || job_id"
        )
        conn.execute("UPDATE tests SET nodeid = nodeid || '[a,b]'")
        client = FakeIssueClient()

        with conn:
            open_issues(conn, client, dry_run=False, issue_author="ci-bot")

        body = client.created[0]["body"]
        self.assertIn("- Integration (noble, amd64) / 1\n", body)
        self.assertIn("- tests/test_lxd.py::test_1[a,b]\n", body)
        self.assertNotIn("\n- amd64)", body)


FIX_SHA = "a" * 40
BUILT_SHA = "b" * 40


def _closed(issue: Dict[str, Any], reason: str, closed_at: str) -> Dict[str, Any]:
    issue.update({"state": "closed", "state_reason": reason, "closed_at": closed_at})
    return issue


class IssueStateSyncTest(unittest.TestCase):
    def _tracked(self, client: FakeIssueClient, number: int = 42) -> sqlite3.Connection:
        conn = _db()
        _insert_opened_issue(
            conn,
            signature_id="abababababababab",
            issue_number=number,
            opened_at="2026-01-01T00:00:00+00:00",
        )
        client.lookup_issues.append(_issue(number, "abababababababab", "lxd"))
        return conn

    def _row(self, conn: sqlite3.Connection, number: int = 42) -> tuple:
        return tuple(
            conn.execute(
                """
                SELECT issue_state, closed_at, closed_reason, fixed_at_sha,
                       reopened_count
                FROM issues WHERE issue_number = ?
                """,
                (number,),
            ).fetchone()
        )

    def test_completed_close_records_reason_and_fixing_commit(self) -> None:
        client = FakeIssueClient()
        conn = self._tracked(client)
        _closed(client.lookup_issues[0], "completed", "2026-02-01T00:00:00Z")
        client.closers[42] = FIX_SHA

        with conn:
            result = sync_issue_states(conn, client, "ci-bot")

        self.assertEqual(result.closed, (42,))
        self.assertEqual(
            self._row(conn), ("closed", "2026-02-01T00:00:00Z", "completed", FIX_SHA, 0)
        )

    def test_not_planned_close_has_no_fixing_commit(self) -> None:
        client = FakeIssueClient()
        conn = self._tracked(client)
        _closed(client.lookup_issues[0], "not_planned", "2026-02-01T00:00:00Z")

        with conn:
            sync_issue_states(conn, client, "ci-bot")

        self.assertEqual(client.closer_calls, [])
        self.assertEqual(self._row(conn)[:4], ("closed", ANY, "not_planned", None))

    def test_unchanged_closed_issue_is_not_resynced(self) -> None:
        client = FakeIssueClient()
        conn = self._tracked(client)
        _closed(client.lookup_issues[0], "completed", "2026-02-01T00:00:00Z")
        client.closers[42] = FIX_SHA

        with conn:
            sync_issue_states(conn, client, "ci-bot")
            second = sync_issue_states(conn, client, "ci-bot")

        self.assertEqual(second.closed, ())
        self.assertEqual(client.closer_calls, [42])

    def test_already_closed_row_without_closer_gets_it_backfilled(self) -> None:
        client = FakeIssueClient()
        conn = self._tracked(client)
        _closed(client.lookup_issues[0], "completed", "2026-02-01T00:00:00Z")
        conn.execute(
            """
            UPDATE issues SET issue_state = 'closed',
                closed_at = '2026-02-01T00:00:00Z', closed_reason = 'completed'
            """
        )
        client.closers[42] = FIX_SHA

        with conn:
            sync_issue_states(conn, client, "ci-bot")

        self.assertEqual(self._row(conn)[3], FIX_SHA)

    def test_closer_lookup_retries_on_error_but_not_after_a_clean_miss(self) -> None:
        client = FakeIssueClient()
        conn = self._tracked(client)
        _closed(client.lookup_issues[0], "completed", "2026-02-01T00:00:00Z")
        client.closers[42] = IssueBridgeError("HTTP 502")

        with conn:
            first = sync_issue_states(conn, client, "ci-bot")
        self.assertIsNone(self._row(conn)[3])
        self.assertIn("closing commit unknown", first.skipped[0])

        del client.closers[42]
        with conn:
            sync_issue_states(conn, client, "ci-bot")
            sync_issue_states(conn, client, "ci-bot")

        self.assertEqual(self._row(conn)[3], "")
        self.assertEqual(client.closer_calls, [42, 42])

    def test_human_reopen_makes_row_live_again(self) -> None:
        client = FakeIssueClient()
        conn = self._tracked(client)
        _closed(client.lookup_issues[0], "completed", "2026-02-01T00:00:00Z")
        client.closers[42] = FIX_SHA
        with conn:
            sync_issue_states(conn, client, "ci-bot")
        client.lookup_issues[0].update(
            {"state": "open", "state_reason": "reopened", "closed_at": None}
        )

        with conn:
            result = sync_issue_states(conn, client, "ci-bot")

        self.assertEqual(result.reopened, (42,))
        self.assertEqual(self._row(conn), ("open", None, None, None, 1))

    def test_unavailable_issue_is_skipped_not_fatal(self) -> None:
        client = FakeIssueClient()
        conn = _db()
        _insert_opened_issue(
            conn,
            signature_id="abababababababab",
            issue_number=404,
            opened_at="2026-01-01T00:00:00+00:00",
        )

        with conn:
            result = sync_issue_states(conn, client, "ci-bot")

        self.assertIn("issue 404: state unavailable", result.skipped[0])
        self.assertEqual(self._row(conn, 404)[0], "open")

    def test_closed_upstream_then_fresh_recurrence_reopens(self) -> None:
        client = FakeIssueClient()
        conn = _db()
        _insert_candidate(conn, "abababababababab", "lxd")
        conn.execute(
            """
            INSERT INTO snap_provenance (channel, arch, revision, built_sha, observed_at)
            SELECT channel, arch, 42, ?, '2026-10-12T00:00:00Z' FROM configs
            """,
            (BUILT_SHA,),
        )
        _insert_opened_issue(
            conn,
            signature_id="abababababababab",
            issue_number=42,
            opened_at="2026-01-01T00:00:00+00:00",
        )
        client.lookup_issues.append(
            _closed(
                _issue(42, "abababababababab", "lxd"),
                "completed",
                "2026-02-01T00:00:00Z",
            )
        )
        client.closers[42] = FIX_SHA

        with patch("metrics.issues._is_ancestor", return_value=True) as ancestor:
            with conn:
                result = open_issues(conn, client, dry_run=False, issue_author="ci-bot")

        self.assertEqual(result.closed_upstream, (42,))
        self.assertEqual(result.reopened, (42,))
        self.assertEqual(client.edits, [(42, {"state": "open"})])
        ancestor.assert_called_with(ANY, FIX_SHA, BUILT_SHA)
        self.assertEqual(self._row(conn)[0], "open")

    def test_closed_upstream_suppresses_recurrence_during_cooldown(self) -> None:
        client = FakeIssueClient()
        conn = _db()
        _insert_candidate(conn, "abababababababab", "lxd")
        _insert_opened_issue(
            conn,
            signature_id="abababababababab",
            issue_number=42,
            opened_at="2026-01-01T00:00:00+00:00",
        )
        client.lookup_issues.append(
            _closed(
                _issue(42, "abababababababab", "lxd"),
                "completed",
                datetime.now(timezone.utc).isoformat(),
            )
        )

        with conn:
            result = open_issues(conn, client, dry_run=False, issue_author="ci-bot")

        self.assertIn("abababababababab:lxd: closed in cooldown", result.skipped)
        self.assertEqual(client.edits, [])

    def test_closer_sha_prefers_merged_pull_request_commit(self) -> None:
        def payload(closer: Dict[str, Any]) -> Dict[str, Any]:
            return {
                "data": {
                    "repository": {
                        "issue": {"timelineItems": {"nodes": [{"closer": closer}]}}
                    }
                }
            }

        repo = {"repository": {"nameWithOwner": "o/r"}}
        merged = {"merged": True, "mergeCommit": {"oid": FIX_SHA}, **repo}
        self.assertEqual(_closer_sha(payload(merged), "o/r"), FIX_SHA)
        self.assertIsNone(_closer_sha(payload(merged), "o/other"))
        self.assertIsNone(
            _closer_sha(payload({"merged": False, "mergeCommit": None, **repo}), "o/r")
        )
        self.assertEqual(
            _closer_sha(payload({"oid": BUILT_SHA, **repo}), "o/r"), BUILT_SHA
        )
        self.assertIsNone(_closer_sha(payload({}), "o/r"))
        self.assertIsNone(_closer_sha({"errors": [{"message": "nope"}]}, "o/r"))


SIG = "cdcdcdcdcdcdcdcd"


def _at(day: int) -> datetime:
    return datetime(2026, 10, 1, tzinfo=timezone.utc) + timedelta(days=day)


class QuietCloseTest(unittest.TestCase):
    """Candidate jobs run on 2026-10-12 (run completed_at); "now" is patched."""

    def setUp(self) -> None:
        self.client = FakeIssueClient()
        self.conn = _db()
        _insert_candidate(self.conn, SIG, "lxd")
        _insert_opened_issue(
            self.conn,
            signature_id=SIG,
            issue_number=42,
            opened_at="2026-10-12T00:00:00+00:00",
        )
        self.client.lookup_issues.append(_issue(42, SIG, "lxd"))

    def _exercise_lxd(
        self, *run_ids: int, day: int = 20, workflow_id: int = 1, config_id: int = 1
    ) -> None:
        """Record runs that collected tests in a matrix config (default: failing)."""

        for run_id in run_ids:
            self.conn.execute(
                """
                INSERT INTO runs (
                    run_id, attempt, workflow_id, started_at, completed_at,
                    conclusion, jobs_total, jobs_success, jobs_failure, jobs_skipped,
                    jobs_cancelled, ingest_version
                ) VALUES (?, 1, ?, ?, ?, 'success', 1, 1, 0, 0, 0, 1)
                """,
                (
                    run_id,
                    workflow_id,
                    f"2026-10-{day:02d}T00:00:00Z",
                    f"2026-10-{day:02d}T01:00:00Z",
                ),
            )
            self.conn.execute(
                "INSERT INTO run_tests_collected VALUES (?, 1, ?, 10)",
                (run_id, config_id),
            )

    def _run(self, day: int, **kwargs: Any) -> Any:
        with patch("metrics.issues._utc_now", return_value=_at(day)):
            with self.conn:
                return open_issues(
                    self.conn,
                    self.client,
                    issue_author="ci-bot",
                    **{"dry_run": False, **kwargs},
                )

    def _row(self) -> tuple:
        return tuple(
            self.conn.execute(
                "SELECT issue_state, closed_reason FROM issues WHERE issue_number = 42"
            ).fetchone()
        )

    def test_quiet_issue_with_exposure_is_closed(self) -> None:
        self._exercise_lxd(300, 301, 302)

        result = self._run(31)

        self.assertEqual(result.closed_quiet, (42,))
        self.assertIn(
            (42, {"state": "closed", "state_reason": "completed"}), self.client.edits
        )
        self.assertIn("no occurrence of this signature", self.client.comments[0][1])
        self.assertEqual(self._row(), ("closed", QUIET_CLOSE_REASON))
        # The bridge's own close survives the next sync and is not re-commented.
        again = self._run(32)
        self.assertEqual((again.closed_quiet, again.closed_upstream), ((), ()))
        self.assertEqual(self._row(), ("closed", QUIET_CLOSE_REASON))
        self.assertEqual(len(self.client.comments), 1)

    def test_silence_without_runs_is_not_evidence(self) -> None:
        self._exercise_lxd(300, 301)

        result = self._run(31)

        self.assertEqual(result.closed_quiet, ())
        self.assertIn(
            "issue 42: quiet 14d but only 2 runs of its workflows and configs since",
            result.skipped[0],
        )
        self.assertEqual(self._row(), ("open", None))

    def test_runs_of_other_workflows_or_configs_are_not_evidence(self) -> None:
        # The failing workflow (1) stopped running; another workflow and an
        # unrelated lxd matrix cell kept going.
        self.conn.execute(
            "INSERT INTO workflows (workflow_id, name, file) VALUES (2, 'pr', 'pr.yaml')"
        )
        self.conn.execute(
            """
            INSERT INTO configs (config_id, os, arch, channel, flavor, substrate)
            VALUES (50, 'jammy', 'arm64', '1.36-classic/edge', 'classic', 'lxd')
            """
        )
        self._exercise_lxd(300, 301, 302, workflow_id=2)
        self._exercise_lxd(303, 304, 305, config_id=50)

        result = self._run(31)

        self.assertEqual(result.closed_quiet, ())
        self.assertEqual(self._row(), ("open", None))

    def test_issue_whose_signature_has_no_jobs_is_not_closed(self) -> None:
        # e.g. every job was re-extracted onto a successor signature.
        self.conn.execute("UPDATE jobs SET signature_id = NULL")
        self._exercise_lxd(300, 301, 302)

        result = self._run(31)

        self.assertEqual(result.closed_quiet, ())
        self.assertIn(
            f"issue 42: signature {SIG} has no stored jobs", result.skipped[0]
        )

    def test_recent_issue_is_not_closed_and_zero_days_disables(self) -> None:
        self._exercise_lxd(300, 301, 302)

        self.assertEqual(self._run(20).closed_quiet, ())
        self.assertEqual(self._run(31, quiet_close_days=0).closed_quiet, ())
        self.assertEqual(self._row(), ("open", None))

    def test_dry_run_only_reports_the_close(self) -> None:
        self._exercise_lxd(300, 301, 302)

        result = self._run(31, dry_run=True)

        self.assertIn("would close quiet issue 42", result.dry_run)
        self.assertEqual(self.client.edits, [])
        self.assertEqual(self._row(), ("open", None))

    def test_recurrence_reopens_a_quiet_closed_issue_without_freshness(self) -> None:
        self._exercise_lxd(300, 301, 302)
        self._run(31)
        # The signature fails again in a later run (no provenance: freshness
        # would be unknown, which never reopens a human-closed issue).
        self.conn.execute(
            """
            INSERT INTO runs (
                run_id, attempt, workflow_id, started_at, completed_at, conclusion,
                jobs_total, jobs_success, jobs_failure, jobs_skipped,
                jobs_cancelled, ingest_version
            ) VALUES (400, 1, 1, '2026-11-02T00:00:00Z', '2026-11-02T01:00:00Z',
                      'failure', 1, 0, 1, 0, 0, 1)
            """
        )
        self.conn.execute(
            """
            INSERT INTO jobs (
                job_id, run_id, attempt, config_id, test_id, job_name, html_url,
                signature_id
            ) VALUES (99, 400, 1, 1, 1, 'lxd job', 'https://example.invalid/99', ?)
            """,
            (SIG,),
        )

        result = self._run(33)

        self.assertEqual(result.reopened, (42,))
        self.assertEqual(self.client.edits[-1], (42, {"state": "open"}))
        self.assertIn("recurred after a quiet close", self.client.comments[-1][1])
        self.assertEqual(self._row(), ("open", None))

    def test_closed_issue_without_recurrence_is_left_alone(self) -> None:
        # A human closed it after the last occurrence; candidates count all
        # history, so without a newer failure there is nothing to say.
        self.client.lookup_issues[0] = _closed(
            _issue(42, SIG, "lxd"), "not_planned", "2026-10-13T00:00:00Z"
        )

        result = self._run(40, quiet_close_days=0)

        self.assertIn(f"{SIG}:lxd: no recurrence since close", result.skipped)
        self.assertEqual(self.client.comments, [])


if __name__ == "__main__":
    unittest.main()
