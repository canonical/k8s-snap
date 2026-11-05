#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for the rule precision check."""

import sqlite3
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.ingest import ingest_run  # noqa: E402
from metrics.precision import check_rules, render_markdown  # noqa: E402
from metrics.rules import load_rules  # noqa: E402
from metrics.store import initialise_database  # noqa: E402

RUN_ID = 200
CELL = "Integration (noble, amd64, 1.37-classic/edge, classic)"

FAILED_LOG = """
##[group]Run sudo apt-get install -y tox
E: Unable to locate package tox
##[error]Failed after 5 attempts
"""
# A green job that retried through the same apt error before succeeding.
HEALED_LOG = """
##[group]Run sudo apt-get install -y tox
E: Unable to locate package tox
Attempt 1/5 failed, retrying in 5s...
Setting up tox (4.11.3-1) ...
"""
CLEAN_LOG = "##[group]Run pytest\n1 passed in 3.2s\n"


def _job(job_id: int, test: str, conclusion: str) -> Dict[str, Any]:
    return {
        "id": job_id,
        "run_id": RUN_ID,
        "run_attempt": 1,
        "name": f"{CELL} / {test}",
        "html_url": f"https://example.invalid/{job_id}",
        "conclusion": conclusion,
        "labels": ["self-hosted"],
        "started_at": "2026-10-12T00:00:00Z",
        "completed_at": "2026-10-12T00:05:00Z",
        "steps": (
            [
                {
                    "name": "Install tox",
                    "number": 3,
                    "conclusion": "failure",
                    "started_at": "2026-10-12T00:00:00Z",
                    "completed_at": "2026-10-12T00:01:00Z",
                }
            ]
            if conclusion == "failure"
            else []
        ),
    }


class FakeClient:
    def __init__(self) -> None:
        self.jobs = [
            _job(1, "tests/test_a.py::test_a", "failure"),
            _job(2, "tests/test_a.py::test_a", "success"),
            _job(3, "tests/test_b.py::test_b", "success"),
        ]
        self.logs = {1: FAILED_LOG, 2: HEALED_LOG, 3: CLEAN_LOG}
        self.downloaded: List[int] = []

    def get_run(self, run_id: int) -> Dict[str, Any]:
        return {
            "id": run_id,
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

    def list_run_jobs(self, run_id: int, attempt: int) -> List[Dict[str, Any]]:
        del run_id, attempt
        return self.jobs

    def download_job_log(self, job_id: int) -> Tuple[int, Optional[str]]:
        self.downloaded.append(job_id)
        return 200, self.logs[job_id]


RULES = """
version: 1
rules:
  - id: noisy-apt
    failure_class: external.dependency
    excerpt_pattern: 'Unable to locate package'
    notes: Matches text a healed green job also prints.
  - id: never-seen
    failure_class: product.bug
    excerpt_pattern: 'kube-proxy should be active'
    notes: Has matched nothing yet.
  - id: pinned
    signature_id: 00000000000000aa
    failure_class: product.bug
    notes: Pinned to an exact signature.
"""


class RulePrecisionTest(unittest.TestCase):
    def setUp(self) -> None:
        self.client = FakeClient()
        self.conn = sqlite3.connect(":memory:")
        self.conn.execute("PRAGMA foreign_keys = ON")
        initialise_database(self.conn)
        ingest_run(self.client, self.conn, RUN_ID)
        self.client.downloaded.clear()
        tmp = tempfile.TemporaryDirectory()
        self.addCleanup(tmp.cleanup)
        path = Path(tmp.name) / "rules.yaml"
        path.write_text(textwrap.dedent(RULES), encoding="utf-8")
        self.pack = load_rules(path)

    def _by_id(self, client: Optional[FakeClient]) -> Dict[str, Any]:
        return {
            report.rule_id: report
            for report in check_rules(self.conn, self.pack, client, sample=5)
        }

    def test_rule_matching_healed_green_jobs_is_imprecise(self) -> None:
        reports = self._by_id(self.client)

        noisy = reports["noisy-apt"]
        self.assertEqual((noisy.signatures, noisy.jobs), (1, 1))
        # Only the green run of the *same* test is the baseline for a used rule.
        self.assertEqual((noisy.success_hits, noisy.success_checked), (1, 1))
        self.assertEqual(noisy.verdict, "imprecise")
        self.assertEqual(noisy.examples, ("E: Unable to locate package tox",))

    def test_unused_rule_is_checked_against_a_general_green_sample(self) -> None:
        reports = self._by_id(self.client)

        unused = reports["never-seen"]
        self.assertEqual(unused.jobs, 0)
        self.assertEqual(unused.success_checked, 2)
        self.assertEqual(unused.verdict, "ok")
        self.assertEqual(reports["pinned"].verdict, "pinned")
        # Each green log is downloaded once and shared across rules.
        self.assertEqual(sorted(self.client.downloaded), [2, 3])

    def test_offline_reports_coverage_without_fetching(self) -> None:
        reports = self._by_id(None)

        self.assertEqual(reports["noisy-apt"].jobs, 1)
        self.assertEqual(reports["noisy-apt"].verdict, "unchecked")
        self.assertEqual(self.client.downloaded, [])
        rendered = render_markdown(list(reports.values()), 0.1)
        self.assertIn("| `never-seen` | 0 | 0 | - | unchecked, unused |", rendered)


if __name__ == "__main__":
    unittest.main()
