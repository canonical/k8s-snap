#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for log fetching and fingerprinting during ingest."""

import argparse
import io
import json
import sqlite3
import sys
import tempfile
import unittest
from contextlib import redirect_stdout
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple, Union
from unittest import mock

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from cmds.metrics import cmd_fetch_logs  # noqa: E402
from metrics.gh import GitHubError  # noqa: E402
from metrics.ingest import build_run_records, ingest_run  # noqa: E402
from metrics.logs import fetch_and_fingerprint, pending_jobs, stale_jobs  # noqa: E402
from metrics.router import apply_router  # noqa: E402
from metrics.store import store_run_record  # noqa: E402
from metrics.store import initialise_database, open_database  # noqa: E402

RUN_ID = 100
RECENT = (datetime.now(timezone.utc) - timedelta(hours=1)).strftime(
    "%Y-%m-%dT%H:%M:%SZ"
)


def _failure_log(ip: str, line: int) -> str:
    return "\n".join(
        [
            "2026-10-12T00:00:01.0000000Z ##[group]Run tests",
            "2026-10-12T00:00:02.0000000Z ==================== FAILURES ====================",
            "2026-10-12T00:00:02.0000000Z ____________ test_smoke ____________",
            '  File "/home/runner/work/k8s-snap/k8s-snap/tests/util.py", '
            f"line {line}, in run",
            "AssertionError: Service kube-proxy should be active, but it is "
            f"inactive on {ip}",
            "==================== short test summary info ====================",
            "##[error]Process completed with exit code 1.",
        ]
    )


def _run(**overrides: Any) -> Dict[str, Any]:
    run = {
        "id": RUN_ID,
        "run_attempt": 1,
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


def _job(
    job_id: int,
    test: str,
    step: Optional[str] = "Run test",
    completed_at: str = RECENT,
) -> Dict[str, Any]:
    steps = []
    if step is not None:
        steps.append(
            {
                "name": step,
                "number": 1,
                "conclusion": "failure",
                "started_at": "2026-10-12T00:00:00Z",
                "completed_at": "2026-10-12T00:01:00Z",
            }
        )
    return {
        "id": job_id,
        "run_id": RUN_ID,
        "run_attempt": 1,
        "name": f"Integration (noble, amd64, 1.37-classic/edge, classic) / {test}",
        "html_url": f"https://example.invalid/{job_id}",
        "conclusion": "failure",
        "labels": ["self-hosted", "amd64"],
        "started_at": completed_at,
        "completed_at": completed_at,
        "steps": steps,
    }


LogSpec = Union[str, int, None, Exception]


class FakeClient:
    """``logs`` maps job ID to log text, a missing-log HTTP status, or an error."""

    def __init__(self, jobs: List[Dict[str, Any]], logs: Dict[int, LogSpec]) -> None:
        self.jobs = jobs
        self.logs = logs
        self.downloaded: List[int] = []

    def get_run(self, run_id: int) -> Dict[str, Any]:
        return _run(id=run_id)

    def list_run_jobs(self, run_id: int, attempt: int) -> List[Dict[str, Any]]:
        del run_id, attempt
        return self.jobs

    def download_job_log(self, job_id: int) -> Tuple[int, Optional[str]]:
        self.downloaded.append(job_id)
        log = self.logs.get(job_id)
        if isinstance(log, Exception):
            raise log
        if isinstance(log, str):
            return 200, log
        return log or 404, None


def _db() -> sqlite3.Connection:
    conn = sqlite3.connect(":memory:")
    conn.execute("PRAGMA foreign_keys = ON")
    initialise_database(conn)
    return conn


def _job_row(conn: sqlite3.Connection, job_id: int) -> tuple:
    return conn.execute(
        "SELECT log_available, signature_id, router_rule_id FROM jobs WHERE job_id = ?",
        (job_id,),
    ).fetchone()


class FetchAndFingerprintTest(unittest.TestCase):
    def test_ingest_run_fingerprints_deferred_jobs_and_collapses_duplicates(
        self,
    ) -> None:
        client = FakeClient(
            [
                _job(1, "tests/test_a.py::test_a"),
                _job(2, "tests/test_b.py::test_b"),
            ],
            {1: _failure_log("10.1.2.3", 412), 2: _failure_log("10.9.9.9", 77)},
        )
        conn = _db()

        ingest_run(client, conn, RUN_ID)

        self.assertEqual(client.downloaded, [1, 2])
        rows = conn.execute(
            "SELECT signature_id, exemplar_job_id, exemplar_excerpt FROM signatures"
        ).fetchall()
        self.assertEqual(len(rows), 1)
        signature_id, exemplar_job_id, excerpt = rows[0]
        self.assertEqual(len(signature_id), 16)
        self.assertEqual(exemplar_job_id, 1)
        self.assertIn("kube-proxy should be active", excerpt)
        self.assertEqual(_job_row(conn, 1), (1, signature_id, None))
        self.assertEqual(_job_row(conn, 2), (1, signature_id, None))
        self.assertEqual(
            conn.execute("SELECT COUNT(*) FROM job_excerpts").fetchone(), (2,)
        )

    def test_ingest_run_can_skip_log_fetching(self) -> None:
        client = FakeClient([_job(1, "tests/test_a.py::test_a")], {})
        conn = _db()

        ingest_run(client, conn, RUN_ID, fetch_logs=False)

        self.assertEqual(client.downloaded, [])
        self.assertEqual([job.job_id for job in pending_jobs(conn)], [1])

    def test_reingest_does_not_refetch_or_reset_log_state(self) -> None:
        client = FakeClient(
            [_job(1, "tests/test_a.py::test_a")], {1: _failure_log("10.1.2.3", 1)}
        )
        conn = _db()

        ingest_run(client, conn, RUN_ID)
        signature = _job_row(conn, 1)[1]
        ingest_run(client, conn, RUN_ID)

        self.assertEqual(client.downloaded, [1])
        self.assertEqual(_job_row(conn, 1), (1, signature, None))

    def test_missing_log_marks_job_and_lets_router_flag_lost_runner(self) -> None:
        client = FakeClient([_job(1, "tests/test_a.py::test_a", step=None)], {1: None})
        conn = _db()

        ingest_run(client, conn, RUN_ID)

        self.assertEqual(_job_row(conn, 1), (0, None, None))
        self.assertEqual(pending_jobs(conn), [])
        apply_router(conn)
        self.assertEqual(_job_row(conn, 1), (0, None, "infra.runner.lost"))

    def test_expired_or_old_missing_log_is_not_runner_lost_evidence(self) -> None:
        old = "2026-01-01T00:00:00Z"
        client = FakeClient(
            [
                _job(1, "tests/test_a.py::test_a", step=None, completed_at=old),
                _job(2, "tests/test_b.py::test_b", step=None),
            ],
            {1: 404, 2: 410},
        )
        conn = _db()

        ingest_run(client, conn, RUN_ID)

        self.assertEqual(_job_row(conn, 1), (2, None, None))
        self.assertEqual(_job_row(conn, 2), (2, None, None))
        self.assertEqual(pending_jobs(conn), [])
        apply_router(conn)
        self.assertEqual(_job_row(conn, 1), (2, None, None))
        self.assertEqual(_job_row(conn, 2), (2, None, None))

    def test_metadata_routed_jobs_are_not_fetched(self) -> None:
        client = FakeClient([_job(1, "tests/test_a.py::test_a", "Set up job")], {})
        conn = _db()

        ingest_run(client, conn, RUN_ID)

        self.assertEqual(client.downloaded, [])
        self.assertEqual(_job_row(conn, 1), (None, None, None))

    def test_fetch_error_leaves_job_pending_for_next_pass(self) -> None:
        client = FakeClient(
            [_job(1, "tests/test_a.py::test_a"), _job(2, "tests/test_b.py::test_b")],
            {1: GitHubError("HTTP 500"), 2: _failure_log("10.1.2.3", 1)},
        )
        conn = _db()
        for record in build_run_records(client, _run()):
            store_run_record(conn, record)

        summary = fetch_and_fingerprint(client, conn, pending_jobs(conn))

        self.assertEqual(summary["errors"], 1)
        self.assertEqual(summary["fingerprinted"], 1)
        self.assertEqual([job.job_id for job in pending_jobs(conn)], [1])

    def test_stored_excerpt_is_scrubbed(self) -> None:
        log = _failure_log("10.1.2.3", 1).replace(
            "AssertionError: Service",
            "AssertionError: token=abc123def456ghi789 Service",
        )
        client = FakeClient([_job(1, "tests/test_a.py::test_a")], {1: log})
        conn = _db()

        ingest_run(client, conn, RUN_ID)

        stored = " ".join(
            conn.execute(
                "SELECT normalised_pattern || exemplar_excerpt FROM signatures"
            ).fetchone()
            + conn.execute("SELECT excerpt FROM job_excerpts").fetchone()
        )
        self.assertNotIn("abc123def456ghi789", stored)

    def _age_signature(self, conn: sqlite3.Connection, job_id: int) -> str:
        """Pretend ``job_id`` was fingerprinted by extractor/normaliser v1."""

        old = "00000000000000aa"
        conn.execute(
            """
            INSERT INTO signatures (
                signature_id, normalised_pattern, exemplar_excerpt, exemplar_job_id,
                normaliser_version, extractor_version
            ) VALUES (?, 'old', 'old', ?, 1, 1)
            """,
            (old, job_id),
        )
        conn.execute("UPDATE jobs SET signature_id = ? WHERE job_id = ?", (old, job_id))
        conn.execute(
            "UPDATE job_excerpts SET extractor_version = 1 WHERE job_id = ?",
            (job_id,),
        )
        conn.commit()
        return old

    def test_reextract_moves_old_version_jobs_to_current_signatures(self) -> None:
        client = FakeClient(
            [_job(1, "tests/test_a.py::test_a"), _job(2, "tests/test_b.py::test_b")],
            {1: _failure_log("10.1.2.3", 1), 2: _failure_log("10.1.2.4", 2)},
        )
        conn = _db()
        ingest_run(client, conn, RUN_ID)
        current = _job_row(conn, 2)[1]
        self._age_signature(conn, 1)

        stale = stale_jobs(conn)
        self.assertEqual([job.job_id for job in stale], [1])
        self.assertEqual(pending_jobs(conn), [])
        summary = fetch_and_fingerprint(client, conn, stale)

        self.assertEqual(summary["reextracted"], 1)
        self.assertEqual(_job_row(conn, 1), (1, current, None))
        self.assertEqual(stale_jobs(conn), [])

    def test_reextract_converges_when_new_version_reproduces_the_same_id(
        self,
    ) -> None:
        client = FakeClient(
            [_job(1, "tests/test_a.py::test_a"), _job(2, "tests/test_b.py::test_b")],
            {1: _failure_log("10.1.2.3", 1), 2: _failure_log("10.1.2.4", 2)},
        )
        conn = _db()
        ingest_run(client, conn, RUN_ID)
        # Job 1 was signed by v1, whose text happened to normalise identically,
        # so both jobs share one signature row that still says version 1.
        conn.execute("UPDATE signatures SET normaliser_version = 1")
        conn.execute("UPDATE job_excerpts SET extractor_version = 1 WHERE job_id = 1")
        conn.commit()

        fetch_and_fingerprint(client, conn, stale_jobs(conn))

        self.assertEqual(stale_jobs(conn), [])

    def test_reextract_keeps_signature_when_log_has_expired(self) -> None:
        client = FakeClient(
            [_job(1, "tests/test_a.py::test_a")], {1: _failure_log("10.1.2.3", 1)}
        )
        conn = _db()
        ingest_run(client, conn, RUN_ID)
        old = self._age_signature(conn, 1)
        client.logs[1] = 404

        summary = fetch_and_fingerprint(client, conn, stale_jobs(conn))

        self.assertEqual(summary["log_gone"], 1)
        self.assertEqual(summary["log_missing"], 0)
        # Recent 404 on a job we once had a log for is expiry, not a lost runner.
        self.assertEqual(_job_row(conn, 1), (2, old, None))
        self.assertEqual(stale_jobs(conn), [])


class FetchLogsCommandTest(unittest.TestCase):
    def test_fetch_logs_drains_pending_jobs_and_dumps(self) -> None:
        client = FakeClient(
            [_job(1, "tests/test_a.py::test_a")], {1: _failure_log("10.1.2.3", 1)}
        )
        with tempfile.TemporaryDirectory() as tmp:
            db, sql = Path(tmp) / "metrics.db", Path(tmp) / "metrics.sql"
            conn = open_database(db)
            with conn:
                for record in build_run_records(client, _run()):
                    store_run_record(conn, record)
            conn.close()
            args = argparse.Namespace(
                db=db,
                repo="o/r",
                token="t",
                run_id=None,
                limit=None,
                reextract=False,
                json=True,
                sql=sql,
                no_dump=False,
            )
            out = io.StringIO()
            with mock.patch("cmds.metrics.GitHubClient", return_value=client):
                with redirect_stdout(out):
                    self.assertEqual(cmd_fetch_logs(args), 0)

            summary = json.loads(out.getvalue())
            self.assertEqual(summary["pending"], 1)
            self.assertEqual(summary["fingerprinted"], 1)
            self.assertIn("INSERT INTO signatures", sql.read_text(encoding="utf-8"))


if __name__ == "__main__":
    unittest.main()
