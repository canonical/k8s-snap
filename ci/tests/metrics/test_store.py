#
# Copyright 2026 Canonical, Ltd.
#
"""Deterministic SQLite store tests."""

import sqlite3
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.ingest import JobRecord, RunRecord, StepRecord  # noqa: E402
from metrics.store import (  # noqa: E402
    dump_database,
    dump_database_bytes,
    initialise_database,
    rebuild_database,
    store_run_record,
)


def _record(run_id: int, workflow: str, job_id: int, arch: str) -> RunRecord:
    return RunRecord(
        run_id=run_id,
        attempt=1,
        workflow_name=workflow,
        workflow_file=f".github/workflows/{workflow.lower()}.yaml",
        event="schedule",
        head_branch="main",
        head_sha=f"{run_id:040x}"[-40:],
        pr_number=None,
        created_at="2026-10-12T00:00:00Z",
        started_at="2026-10-12T00:00:00Z",
        completed_at="2026-10-12T00:10:00Z",
        conclusion="failure",
        jobs_total=1,
        jobs_success=0,
        jobs_failure=1,
        jobs_skipped=0,
        jobs_cancelled=0,
        runner_minutes_total=1.0,
        runner_minutes_failed=1.0,
        failed_jobs=[
            JobRecord(
                job_id=job_id,
                run_id=run_id,
                attempt=1,
                config_key=("noble", arch, "1.37-classic/edge", "classic", "lxd"),
                test_nodeid=f"tests/test_{arch}.py::test_smoke",
                test_file=f"tests/test_{arch}.py",
                runner_labels=("self-hosted", arch),
                self_hosted=True,
                job_name=f"Integration {arch}",
                caller_job="Integration",
                html_url=f"https://example.invalid/{job_id}",
                started_at="2026-10-12T00:00:00Z",
                completed_at="2026-10-12T00:01:00Z",
                duration_s=60.0,
                failed_step_name="Run test",
                failed_step_number=1,
                log_available=1,
                steps=(
                    StepRecord(
                        name="Run test",
                        number=1,
                        conclusion="failure",
                        duration_s=60.0,
                    ),
                ),
            )
        ],
        tests_collected={("noble", arch, "1.37-classic/edge", "classic", "lxd"): 1},
    )


def _conn() -> sqlite3.Connection:
    conn = sqlite3.connect(":memory:")
    conn.execute("PRAGMA foreign_keys = ON")
    initialise_database(conn)
    return conn


def _dump_for(records: list[RunRecord]) -> bytes:
    conn = _conn()
    with conn:
        for record in records:
            store_run_record(conn, record)
    return dump_database_bytes(conn)


class StoreTest(unittest.TestCase):
    def test_store_run_record_is_idempotent(self) -> None:
        conn = _conn()
        record = _record(100, "Nightly", 1001, "amd64")

        with conn:
            store_run_record(conn, record)
            first = dump_database_bytes(conn)
            store_run_record(conn, record)
            second = dump_database_bytes(conn)

        self.assertEqual(first, second)
        self.assertEqual(conn.execute("SELECT COUNT(*) FROM runs").fetchone(), (1,))
        self.assertEqual(conn.execute("SELECT COUNT(*) FROM jobs").fetchone(), (1,))
        self.assertEqual(
            conn.execute("SELECT COUNT(*) FROM job_steps").fetchone(), (1,)
        )

    def test_dump_round_trips_and_excludes_ephemeral_tables(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            db_path = root / "metrics.db"
            rebuilt_path = root / "rebuilt.db"
            dump_path = root / "metrics.sql"
            conn = sqlite3.connect(db_path)
            conn.execute("PRAGMA foreign_keys = ON")
            initialise_database(conn)
            with conn:
                store_run_record(conn, _record(100, "Nightly", 1001, "amd64"))
                conn.execute(
                    """
                    INSERT INTO signatures (
                        signature_id, normalised_pattern, exemplar_excerpt,
                        exemplar_job_id, normaliser_version, extractor_version
                    ) VALUES (?, ?, ?, ?, ?, ?)
                    """,
                    ("sig-amd64", "pattern", "durable exemplar", 1001, 1, 1),
                )
                conn.execute(
                    "UPDATE jobs SET signature_id = ? WHERE job_id = ?",
                    ("sig-amd64", 1001),
                )
                conn.execute(
                    """
                    INSERT INTO job_excerpts
                    (job_id, excerpt, source, extractor_version)
                    VALUES (?, ?, ?, ?)
                    """,
                    (1001, "ephemeral excerpt", "errors-block", 1),
                )

            first = dump_database(conn, dump_path)
            self.assertNotIn(b"job_excerpts", first)
            self.assertNotIn(b"job_steps", first)
            self.assertNotIn(b"ephemeral excerpt", first)
            self.assertIn(b"durable exemplar", first)

            rebuilt = rebuild_database(rebuilt_path, dump_path)
            second = dump_database_bytes(rebuilt)

        self.assertEqual(first, second)

    def test_dump_is_independent_of_insertion_order(self) -> None:
        records = [
            _record(100, "Nightly", 1001, "amd64"),
            _record(200, "Weekly", 2001, "arm64"),
        ]

        forward = _dump_for(records)
        reverse = _dump_for(list(reversed(records)))

        self.assertEqual(forward, reverse)


if __name__ == "__main__":
    unittest.main()
