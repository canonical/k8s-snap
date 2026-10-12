#
# Copyright 2026 Canonical, Ltd.
#
"""Derived rollup and report tests."""

import sqlite3
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.report import render_latest_markdown  # noqa: E402
from metrics.rollup import derive_rollup, derive_rollups  # noqa: E402
from metrics.store import initialise_database  # noqa: E402


def _db() -> sqlite3.Connection:
    conn = sqlite3.connect(":memory:")
    conn.execute("PRAGMA foreign_keys = ON")
    initialise_database(conn)
    conn.execute(
        """
        INSERT INTO workflows (workflow_id, name, file)
        VALUES (1, 'nightly-test', 'nightly-test.yaml'),
               (2, 'lint-and-integration', 'lint_and_integration.yaml')
        """
    )
    conn.execute(
        """
        INSERT INTO configs (config_id, os, arch, channel, flavor, substrate)
        VALUES (1, 'noble', 'amd64', '1.37-classic/edge', 'classic', 'lxd')
        """
    )
    conn.execute(
        "INSERT INTO tests (test_id, nodeid, file) VALUES (1, 'tests/test_a.py::test_a', 'tests/test_a.py')"
    )
    return conn


def _insert_run(
    conn: sqlite3.Connection,
    run_id: int,
    *,
    workflow_id: int = 1,
    event: str = "schedule",
    conclusion: str = "failure",
    jobs_total: int = 10,
    jobs_success: int = 9,
    jobs_failure: int = 1,
    started_at: str = "2026-10-12T00:00:00Z",
) -> None:
    conn.execute(
        """
        INSERT INTO runs (
            run_id, attempt, workflow_id, event, completed_at, started_at,
            conclusion, jobs_total, jobs_success, jobs_failure, jobs_skipped,
            jobs_cancelled, runner_minutes_total, runner_minutes_failed,
            ingest_version
        ) VALUES (?, 1, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 10.0, 1.0, 1)
        """,
        (
            run_id,
            workflow_id,
            event,
            started_at.replace("00:00:00", "00:10:00"),
            started_at,
            conclusion,
            jobs_total,
            jobs_success,
            jobs_failure,
        ),
    )
    conn.execute(
        """
        INSERT INTO run_tests_collected (run_id, attempt, config_id, count)
        VALUES (?, 1, 1, 42)
        """,
        (run_id,),
    )


def _insert_signature(
    conn: sqlite3.Connection,
    signature_id: str,
    job_id: int,
    *,
    failure_class: str = "product.bug",
    subclass: str = "service",
    rule_id: str = "product.service",
) -> None:
    conn.execute(
        """
        INSERT INTO signatures (
            signature_id, normalised_pattern, exemplar_excerpt, exemplar_job_id,
            normaliser_version, extractor_version
        ) VALUES (?, ?, ?, ?, 1, 1)
        """,
        (signature_id, f"pattern {signature_id}", "exemplar", job_id),
    )
    conn.execute(
        """
        INSERT INTO signature_classifications (
            signature_id, substrate, ruleset_version, failure_class, subclass,
            rule_id, confidence, classified_by, classified_at
        ) VALUES (?, '*', 1, ?, ?, ?, 1.0, 'rules', '2026-10-12T00:00:00Z')
        """,
        (signature_id, failure_class, subclass, rule_id),
    )


def _insert_failed_job(
    conn: sqlite3.Connection,
    run_id: int,
    job_id: int,
    signature_id: str,
) -> None:
    conn.execute(
        """
        INSERT INTO jobs (
            job_id, run_id, attempt, config_id, test_id, job_name, html_url,
            started_at, completed_at, duration_s, failed_step_name,
            failed_step_number, log_available, signature_id
        ) VALUES (?, ?, 1, 1, 1, ?, ?, '2026-10-12T00:01:00Z',
                  '2026-10-12T00:02:00Z', 60.0, 'Run pytest', 1, 1, ?)
        """,
        (
            job_id,
            run_id,
            f"job {job_id}",
            f"https://example.invalid/{job_id}",
            signature_id,
        ),
    )


class RollupReportTest(unittest.TestCase):
    def test_report_renders_zero_denominator_rates_as_na(self) -> None:
        conn = _db()
        _insert_run(
            conn,
            100,
            event="schedule",
            conclusion="success",
            jobs_total=1,
            jobs_success=1,
            jobs_failure=0,
        )

        report = render_latest_markdown(derive_rollups(conn))

        self.assertIn("PR first-pass", report)
        self.assertIn(
            "| 2026-10 | 1 | 0 | 0.0% (n=1 jobs) | 100.0% (n=1 runs) | n/a | n/a | n/a |",
            report,
        )

    def test_report_includes_m11_blocked_jobs_in_failure_pressure(self) -> None:
        conn = _db()
        _insert_run(conn, 100, jobs_total=10, jobs_success=9, jobs_failure=1)
        _insert_signature(conn, "1111111111111111", 1001)
        _insert_failed_job(conn, 100, 1001, "1111111111111111")
        conn.execute(
            """
            INSERT INTO run_blocked_jobs (
                run_id, attempt, prepare_job_name, jobs_not_run
            ) VALUES (100, 1, 'Prepare Environment', 10)
            """
        )

        rollup = derive_rollup(conn, "2026-10")
        report = render_latest_markdown([rollup])

        self.assertEqual(rollup["m11_failure_pressure_rate"], 55.0)
        self.assertIn(
            "- Failure pressure including blocked jobs: 55.0% (n=20 jobs plus blocked)",
            report,
        )
        self.assertIn(
            "- M11 jobs blocked by failed `Prepare Environment`: 10 across 1 prepare job(s).",
            report,
        )

    def test_report_is_deterministic_for_tied_signatures_inserted_in_different_orders(
        self,
    ) -> None:
        def build(signature_order: list[str]) -> str:
            conn = _db()
            _insert_run(conn, 100)
            for signature_id in signature_order:
                job_id = 1001 if signature_id.startswith("1") else 1002
                _insert_signature(conn, signature_id, job_id)
                _insert_failed_job(conn, 100, job_id, signature_id)
            return render_latest_markdown([derive_rollup(conn, "2026-10")])

        forward = build(["1111111111111111", "2222222222222222"])
        reverse = build(["2222222222222222", "1111111111111111"])

        self.assertEqual(forward, reverse)
        self.assertLess(
            forward.index("`1111111111111111`"),
            forward.index("`2222222222222222`"),
        )


if __name__ == "__main__":
    unittest.main()
