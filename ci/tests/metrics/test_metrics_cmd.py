#
# Copyright 2026 Canonical, Ltd.
#
"""Metrics CLI tests."""

import argparse
import sqlite3
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from cmds.metrics import cmd_audit_secrets, cmd_report  # noqa: E402
from metrics.store import initialise_database  # noqa: E402


class MetricsCommandTest(unittest.TestCase):
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


if __name__ == "__main__":
    unittest.main()
