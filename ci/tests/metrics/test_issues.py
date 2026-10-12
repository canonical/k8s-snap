#
# Copyright 2026 Canonical, Ltd.
#
"""Signature-to-issue bridge safety tests."""

from __future__ import annotations

import sqlite3
import sys
import tempfile
import unittest
from pathlib import Path
from typing import Any, Dict, List, Sequence

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.issues import open_issues  # noqa: E402
from metrics.models import SCHEMA_SQL  # noqa: E402


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


class FakeIssueClient:
    def __init__(self) -> None:
        self.sweep_issues: List[Dict[str, Any]] = []
        self.lookup_issues: List[Dict[str, Any]] = []
        self.created: List[Dict[str, Any]] = []
        self.edits: List[tuple[int, Dict[str, Any]]] = []
        self.comments: List[tuple[int, str]] = []
        self._next = 1000

    def current_user(self) -> str:
        return "ci-bot"

    def list_issues(
        self, labels: Sequence[str], state: str = "open"
    ) -> List[Dict[str, Any]]:
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
        for issue in self.created:
            if issue["number"] == number:
                return dict(issue)
        raise AssertionError(f"unknown issue {number}")

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


if __name__ == "__main__":
    unittest.main()
