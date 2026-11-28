#
# Copyright 2026 Canonical, Ltd.
#
"""Ingest tests for completed-run handling and attempt coverage."""

import sqlite3
import sys
import unittest
from pathlib import Path
from typing import Any, Dict, List

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.ingest import (  # noqa: E402
    InFlightRunError,
    build_run_records,
    initialise_database,
    parse_job_name,
    store_run_record,
)


def _run(**overrides: Any) -> Dict[str, Any]:
    run = {
        "id": 100,
        "run_attempt": 2,
        "name": "Nightly",
        "path": ".github/workflows/nightly-test.yaml@refs/heads/main",
        "event": "schedule",
        "head_branch": "main",
        "head_sha": "deadbeef",
        "pull_requests": [],
        "created_at": "2026-10-12T00:00:00Z",
        "run_started_at": "2026-10-12T00:00:00Z",
        "updated_at": "2026-10-12T00:10:00Z",
        "conclusion": "failure",
    }
    run.update(overrides)
    return run


def _job(job_id: int, attempt: int, conclusion: str) -> Dict[str, Any]:
    return {
        "id": job_id,
        "run_id": 100,
        "run_attempt": attempt,
        "name": (
            "Integration (noble, amd64, 1.37-classic/edge, classic) / "
            "tests/test_smoke.py::test_smoke"
        ),
        "html_url": f"https://example.invalid/{job_id}",
        "conclusion": conclusion,
        "labels": ["self-hosted", "amd64"],
        "started_at": "2026-10-12T00:00:00Z",
        "completed_at": "2026-10-12T00:01:00Z",
        "steps": [
            {
                "name": "Run test",
                "number": 1,
                "conclusion": conclusion,
                "started_at": "2026-10-12T00:00:00Z",
                "completed_at": "2026-10-12T00:01:00Z",
            }
        ],
    }


class FakeClient:
    def __init__(self) -> None:
        self.calls: List[int] = []

    def list_run_jobs(self, run_id: int, attempt: int) -> List[Dict[str, Any]]:
        del run_id
        self.calls.append(attempt)
        if attempt == 1:
            return [_job(1, 1, "failure")]
        return [_job(2, 2, "success")]


class IngestTest(unittest.TestCase):
    def test_parse_job_name_recovers_dimensions_by_content(self) -> None:
        parsed = parse_job_name(
            "Integration (noble, amd64, 1.37-classic/edge, classic) / "
            "tests/test_smoke.py::test_smoke"
        )

        self.assertEqual(parsed.os, "noble")
        self.assertEqual(parsed.arch, "amd64")
        self.assertEqual(parsed.channel, "1.37-classic/edge")
        self.assertEqual(parsed.flavor, "classic")
        self.assertEqual(parsed.substrate, "multipass")
        self.assertEqual(parsed.test_nodeid, "tests/test_smoke.py::test_smoke")

    def test_build_run_records_fetches_every_attempt_and_positive_retry_evidence(
        self,
    ) -> None:
        client = FakeClient()

        records = build_run_records(client, _run())

        self.assertEqual([record.attempt for record in records], [1, 2])
        self.assertEqual(client.calls, [1, 2, 2])
        self.assertEqual(records[0].jobs_failure, 1)
        self.assertEqual(records[1].jobs_success, 1)
        self.assertEqual(records[1].retry_outcomes[0].conclusion, "success")

    def test_in_flight_run_is_refused_before_storage(self) -> None:
        client = FakeClient()

        with self.assertRaises(InFlightRunError):
            build_run_records(client, _run(conclusion=None))

    def test_store_run_record_writes_failed_jobs(self) -> None:
        client = FakeClient()
        records = build_run_records(client, _run(run_attempt=1))
        conn = sqlite3.connect(":memory:")
        conn.execute("PRAGMA foreign_keys = ON")
        initialise_database(conn)

        store_run_record(conn, records[0])

        self.assertEqual(conn.execute("SELECT COUNT(*) FROM runs").fetchone(), (1,))
        self.assertEqual(conn.execute("SELECT COUNT(*) FROM jobs").fetchone(), (1,))
        self.assertEqual(
            conn.execute("SELECT COUNT(*) FROM job_steps").fetchone(), (1,)
        )


if __name__ == "__main__":
    unittest.main()
