#
# Copyright 2026 Canonical, Ltd.
#
"""SQLite schema constraints from the CI failure classification plan."""

import sqlite3
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.models import SCHEMA_SQL  # noqa: E402


def _db() -> sqlite3.Connection:
    conn = sqlite3.connect(":memory:")
    conn.execute("PRAGMA foreign_keys = ON")
    conn.executescript(SCHEMA_SQL)
    conn.execute(
        "INSERT INTO workflows (workflow_id, name, file) VALUES (1, 'nightly', 'nightly-test.yaml')"
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
    return conn


class SchemaTest(unittest.TestCase):
    def test_configs_unique_deduplicates_empty_channel_cells(self) -> None:
        conn = _db()
        values = ("noble", "amd64", "", "classic", "multipass")
        conn.execute(
            "INSERT INTO configs (os, arch, channel, flavor, substrate) VALUES (?, ?, ?, ?, ?)",
            values,
        )

        with self.assertRaises(sqlite3.IntegrityError):
            conn.execute(
                "INSERT INTO configs (os, arch, channel, flavor, substrate) VALUES (?, ?, ?, ?, ?)",
                values,
            )

    def test_jobs_check_accepts_one_or_no_tested_source(self) -> None:
        conn = _db()
        conn.execute(
            "INSERT INTO configs (config_id, os, arch, channel, flavor, substrate) VALUES (1, '', '', '', '', '')"
        )

        def insert_job(job_id: int, revision: str, artifact: str) -> None:
            conn.execute(
                """
                INSERT INTO jobs (
                    job_id, run_id, attempt, config_id, test_id, job_name, html_url,
                    tested_revision, tested_artifact
                ) VALUES (?, 100, 1, 1, 0, 'job', 'https://example.invalid', ?, ?)
                """,
                (job_id, revision, artifact),
            )

        insert_job(1, "123", "")
        insert_job(2, "", "k8s-snap")
        insert_job(3, "", "")
        with self.assertRaises(sqlite3.IntegrityError):
            insert_job(4, "123", "k8s-snap")

    def test_snap_provenance_resolves_revision_to_built_sha(self) -> None:
        conn = _db()
        conn.execute(
            """
            INSERT INTO snap_provenance (channel, arch, revision, built_sha, observed_at)
            VALUES ('1.37-classic/edge', 'amd64', 42, 'abc123', '2026-10-12T00:00:00Z')
            """
        )
        row = conn.execute(
            """
            SELECT built_sha FROM snap_provenance
            WHERE channel = ? AND arch = ? AND revision = ?
            """,
            ("1.37-classic/edge", "amd64", 42),
        ).fetchone()
        self.assertEqual(row, ("abc123",))

    def test_open_issue_cannot_lack_issue_number(self) -> None:
        conn = _db()
        conn.execute(
            """
            INSERT INTO signatures (
                signature_id, normalised_pattern, exemplar_excerpt,
                normaliser_version, extractor_version
            ) VALUES ('sig1', 'pattern', 'excerpt', 1, 1)
            """
        )
        with self.assertRaises(sqlite3.IntegrityError):
            conn.execute(
                """
                INSERT INTO issues (
                    signature_id, idempotency_key, issue_state, claimed_at
                ) VALUES ('sig1', 'sig1:*', 'open', '2026-10-12T00:00:00Z')
                """
            )

    def test_claiming_index_prevents_second_live_row_in_one_database(self) -> None:
        conn = _db()
        conn.execute(
            """
            INSERT INTO signatures (
                signature_id, normalised_pattern, exemplar_excerpt,
                normaliser_version, extractor_version
            ) VALUES ('sig1', 'pattern', 'excerpt', 1, 1)
            """
        )
        conn.execute(
            """
            INSERT INTO issues (
                signature_id, substrate, idempotency_key, issue_state, claimed_at
            ) VALUES ('sig1', 'lxd', 'one', 'claiming', '2026-10-12T00:00:00Z')
            """
        )
        with self.assertRaises(sqlite3.IntegrityError):
            conn.execute(
                """
                INSERT INTO issues (
                    signature_id, substrate, idempotency_key, issue_state, claimed_at
                ) VALUES ('sig1', 'lxd', 'two', 'claiming', '2026-10-12T00:00:00Z')
                """
            )


if __name__ == "__main__":
    unittest.main()
