#
# Copyright 2026 Canonical, Ltd.
#
"""Stage A metadata-router tests."""

import sqlite3
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.models import SCHEMA_SQL  # noqa: E402
from metrics.router import RouterJob, apply_router, route  # noqa: E402


def _db() -> sqlite3.Connection:
    conn = sqlite3.connect(":memory:")
    conn.execute("PRAGMA foreign_keys = ON")
    conn.executescript(SCHEMA_SQL)
    conn.execute(
        "INSERT INTO workflows (workflow_id, name, file) VALUES (1, 'nightly', 'nightly.yaml')"
    )
    conn.execute("INSERT INTO tests (test_id, nodeid) VALUES (0, '')")
    conn.execute(
        """
        INSERT INTO runs (
            run_id, attempt, workflow_id, completed_at, conclusion, jobs_total,
            jobs_success, jobs_failure, jobs_skipped, jobs_cancelled,
            runner_minutes_total, runner_minutes_failed, ingest_version
        ) VALUES (100, 1, 1, '2026-10-12T00:00:00Z', 'failure', 1, 0, 1, 0, 0, 1.0, 1.0, 1)
        """
    )
    conn.execute(
        """
        INSERT INTO configs (config_id, os, arch, channel, flavor, substrate)
        VALUES (1, '', 'amd64', '', '', '')
        """
    )
    conn.execute(
        """
        INSERT INTO configs (config_id, os, arch, channel, flavor, substrate)
        VALUES (2, '', 'amd64', '1.37-classic/edge', '', '')
        """
    )
    return conn


def _insert_job(
    conn: sqlite3.Connection,
    job_id: int,
    failed_step_name: str | None,
    log_available: int,
    config_id: int = 1,
) -> None:
    conn.execute(
        """
        INSERT INTO jobs (
            job_id, run_id, attempt, config_id, test_id, job_name, html_url,
            failed_step_name, log_available
        ) VALUES (?, 100, 1, ?, 0, ?, 'https://example.invalid', ?, ?)
        """,
        (job_id, config_id, f"job {job_id}", failed_step_name, log_available),
    )


class RouterTest(unittest.TestCase):
    def test_runner_lost_uses_no_failed_step_and_missing_log(self) -> None:
        verdict = route(
            RouterJob(
                job_id=1,
                failed_step_name=None,
                log_available=0,
                duration_s=3120.0,
            )
        )

        self.assertFalse(verdict.defer)
        self.assertEqual(verdict.failure_class, "infra.runner")
        self.assertEqual(verdict.router_rule_id, "infra.runner.lost")

    def test_test_steps_defer_to_stage_b(self) -> None:
        verdict = route(
            RouterJob(
                job_id=1,
                failed_step_name="Run test_version_upgrades",
                log_available=1,
            )
        )

        self.assertTrue(verdict.defer)

    def test_runner_lost_classification_uses_lookup_table(self) -> None:
        conn = _db()
        _insert_job(conn, 1, None, 0)

        apply_router(conn)
        row = conn.execute(
            "SELECT failure_class, subclass, classified_by FROM v_job_class WHERE job_id = 1"
        ).fetchone()

        self.assertEqual(row, ("infra.runner", "runner_lost", "router"))

    def test_step_rule_lxd_setup_preserves_full_dotted_class(self) -> None:
        verdict = route(
            RouterJob(job_id=1, failed_step_name="Setup LXD", log_available=1)
        )
        self.assertEqual(verdict.failure_class, "infra.provisioning")

        conn = _db()
        _insert_job(conn, 1, "Setup LXD", 1)

        apply_router(conn)
        row = conn.execute(
            "SELECT failure_class, subclass, classified_by FROM v_job_class WHERE job_id = 1"
        ).fetchone()

        self.assertEqual(row, ("infra.provisioning", "lxd_setup", "router"))

    def test_step_rule_snap_download_preserves_full_dotted_class(self) -> None:
        verdict = route(
            RouterJob(
                job_id=1,
                failed_step_name="Download k8s-snap",
                log_available=1,
                channel="1.37-classic/edge",
            )
        )
        self.assertEqual(verdict.failure_class, "external.dependency")

        conn = _db()
        _insert_job(conn, 1, "Download k8s-snap", 1, config_id=2)

        apply_router(conn)
        row = conn.execute(
            "SELECT failure_class, subclass, classified_by FROM v_job_class WHERE job_id = 1"
        ).fetchone()

        self.assertEqual(row, ("external.dependency", "snap_store", "router"))


if __name__ == "__main__":
    unittest.main()
