#
# Copyright 2026 Canonical, Ltd.
#
"""Metrics CLI tests."""

import argparse
import io
import json
import sqlite3
import subprocess
import sys
import tempfile
import unittest
from contextlib import redirect_stdout
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from cmds.metrics import cmd_report  # noqa: E402
from cmds.metrics import cmd_audit_secrets, cmd_digest, cmd_rebuild  # noqa: E402
from metrics.store import initialise_database  # noqa: E402


class MetricsCommandTest(unittest.TestCase):
    def test_cli_loads_without_requests_installed(self) -> None:
        # The e2e workflow runs `k8s-ci.py e2e ...` with a bare python3; loading
        # the metrics commands must not require third-party packages.
        script = Path(__file__).resolve().parents[2] / "k8s-ci.py"
        code = "\n".join(
            [
                "import runpy, sys",
                "sys.modules['requests'] = None",
                "sys.argv = ['k8s-ci.py', 'e2e', '--help']",
                f"runpy.run_path({str(script)!r}, run_name='__main__')",
            ]
        )

        result = subprocess.run(  # noqa: S603
            [sys.executable, "-c", code], capture_output=True, text=True, check=False
        )

        self.assertEqual(result.returncode, 0, result.stderr)

    def test_rebuild_without_dump_fails_unless_init_if_missing(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            db, sql = Path(tmp) / "metrics.db", Path(tmp) / "metrics.sql"
            args = argparse.Namespace(db=db, sql=sql, init_if_missing=False)
            with self.assertRaises(FileNotFoundError):
                cmd_rebuild(args)

            args.init_if_missing = True
            with redirect_stdout(io.StringIO()):
                self.assertEqual(cmd_rebuild(args), 0)

            conn = sqlite3.connect(db)
            self.assertEqual(
                conn.execute("SELECT value FROM schema_meta").fetchone(), ("1",)
            )
            self.assertEqual(conn.execute("SELECT COUNT(*) FROM runs").fetchone(), (0,))

    def test_audit_secrets_exits_nonzero_on_file_finding(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "excerpt.txt"
            path.write_text("token=abc123def456\n", encoding="utf-8")
            args = argparse.Namespace(paths=[path], db=Path(tmp) / "missing.db")

            self.assertEqual(cmd_audit_secrets(args), 1)

    def test_audit_secrets_scans_database_excerpts(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            db = Path(tmp) / "metrics.db"
            conn = sqlite3.connect(db)
            conn.execute("CREATE TABLE job_excerpts (job_id INTEGER, excerpt TEXT)")
            conn.execute(
                "INSERT INTO job_excerpts (job_id, excerpt) VALUES (1, ?)",
                ("sudo k8s join <redacted> failed",),
            )
            conn.commit()
            args = argparse.Namespace(paths=[], db=db)

            self.assertEqual(cmd_audit_secrets(args), 0)

    def test_report_writes_to_explicit_output_directories(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            db = root / "metrics.db"
            conn = sqlite3.connect(db)
            initialise_database(conn)
            conn.close()
            reports_dir = root / "ci-metrics-checkout" / "reports"
            signatures_dir = root / "ci-metrics-checkout" / "signatures"
            args = argparse.Namespace(
                db=db,
                reports_dir=reports_dir,
                signatures_dir=signatures_dir,
            )

            self.assertEqual(cmd_report(args), 0)

            self.assertTrue((reports_dir / "latest.md").exists())
            self.assertTrue((signatures_dir / "catalogue.md").exists())

    def test_digest_prints_mattermost_payload_with_report_link(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            db = Path(tmp) / "metrics.db"
            conn = sqlite3.connect(db)
            initialise_database(conn)
            conn.execute(
                "INSERT INTO workflows (workflow_id, name, file) VALUES (1, 'Nightly', 'nightly-test.yaml')"
            )
            conn.execute(
                """
                INSERT INTO runs (
                    run_id, attempt, workflow_id, event, completed_at, started_at,
                    conclusion, jobs_total, jobs_success, jobs_failure, jobs_skipped,
                    jobs_cancelled, runner_minutes_total, runner_minutes_failed,
                    ingest_version
                ) VALUES (100, 1, 1, 'schedule', '2026-10-12T00:10:00Z',
                          '2026-10-12T00:00:00Z', 'failure', 10, 9, 1, 0, 0, 10.0, 1.0, 1)
                """
            )
            conn.commit()
            conn.close()
            args = argparse.Namespace(
                db=db,
                repo="canonical/k8s-snap",
                report_ref="ci-metrics",
                report_path="ci/metrics-data/reports/latest.md",
            )

            stdout = io.StringIO()
            with redirect_stdout(stdout):
                exit_code = cmd_digest(args)

            self.assertEqual(exit_code, 0)
            payload = json.loads(stdout.getvalue())
            self.assertIn("text", payload)
            self.assertIn("Attribution and delta", payload["text"])
            self.assertIn(
                "https://github.com/canonical/k8s-snap/blob/ci-metrics/"
                "ci/metrics-data/reports/latest.md",
                payload["text"],
            )


if __name__ == "__main__":
    unittest.main()
