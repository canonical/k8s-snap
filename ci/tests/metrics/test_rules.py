#
# Copyright 2026 Canonical, Ltd.
#
"""Stage B rule-pack tests."""

import logging
import sqlite3
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.models import SCHEMA_SQL  # noqa: E402
from metrics.rules import (  # noqa: E402
    RuleContext,
    RulePackError,
    classify_signatures,
    load_rules,
)


def _write_rules(text: str) -> Path:
    root = Path(tempfile.mkdtemp())
    path = root / "rules.yaml"
    path.write_text(textwrap.dedent(text), encoding="utf-8")
    return path


def _db(substrates: tuple[str, ...] = ("lxd",)) -> sqlite3.Connection:
    conn = sqlite3.connect(":memory:")
    conn.execute("PRAGMA foreign_keys = ON")
    conn.executescript(SCHEMA_SQL)
    conn.execute(
        "INSERT INTO workflows (workflow_id, name, file) VALUES (1, 'nightly', 'nightly.yaml')"
    )
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
        INSERT INTO signatures (
            signature_id, normalised_pattern, exemplar_excerpt,
            normaliser_version, extractor_version
        ) VALUES (
            '8c531e6ba49b3382',
            'Failed to meet condition',
            'subprocess.CalledProcessError: Failed to meet condition',
            1,
            1
        )
        """
    )
    for index, substrate in enumerate(substrates, start=1):
        conn.execute(
            """
            INSERT INTO configs (config_id, os, arch, channel, flavor, substrate)
            VALUES (?, '', 'amd64', '', '', ?)
            """,
            (index, substrate),
        )
        conn.execute(
            """
            INSERT INTO tests (test_id, nodeid, file) VALUES (?, ?, ?)
            """,
            (
                index,
                f"tests/test_{substrate or 'unknown'}.py::test_case",
                "tests/test.py",
            ),
        )
        conn.execute(
            """
            INSERT INTO jobs (
                job_id, run_id, attempt, config_id, test_id, job_name, html_url,
                failed_step_name, signature_id
            ) VALUES (?, 100, 1, ?, ?, ?, 'https://example.invalid', 'Run test_case', '8c531e6ba49b3382')
            """,
            (index, index, index, f"job {substrate or 'unknown'}"),
        )
    return conn


class RulesTest(unittest.TestCase):
    def test_invalid_failure_class_is_fatal_at_load_time(self) -> None:
        path = _write_rules(
            """
            version: 1
            rules:
              - id: typo
                signature_id: 8c531e6ba49b3382
                failure_class: product_bug
                notes: typo must fail
            """
        )

        with self.assertRaises(RulePackError):
            load_rules(path)

    def test_shadowed_pinned_rule_warns(self) -> None:
        path = _write_rules(
            """
            version: 1
            rules:
              - id: broad
                signature_id: 8c531e6ba49b3382
                failure_class: product.bug
                notes: broad pin
              - id: shadowed
                signature_id: 8c531e6ba49b3382
                failure_class: infra.provisioning
                job:
                  substrate: '^multipass$'
                notes: unreachable behind broad pin
            """
        )

        with self.assertLogs("metrics.rules", level=logging.WARNING) as logs:
            load_rules(path)

        self.assertIn("unreachable", "\n".join(logs.output))

    def test_substrate_guard_leaves_unknown_substrate_unclassified(self) -> None:
        conn = _db(("",))
        pack = load_rules(Path(__file__).resolve().parents[2] / "failure_rules.yaml")

        classify_signatures(conn, pack)
        row = conn.execute(
            "SELECT failure_class FROM v_job_class WHERE job_id = 1"
        ).fetchone()

        self.assertEqual(row, ("unknown",))

    def test_same_signature_splits_by_substrate(self) -> None:
        conn = _db(("multipass", "lxd"))
        pack = load_rules(Path(__file__).resolve().parents[2] / "failure_rules.yaml")

        classify_signatures(conn, pack)
        rows = conn.execute(
            "SELECT job_id, failure_class FROM v_job_class ORDER BY job_id"
        ).fetchall()

        self.assertEqual(rows, [(1, "infra.provisioning"), (2, "product.bug")])

    def test_rule_ordering_keeps_infra_above_product(self) -> None:
        infra_first = load_rules(
            _write_rules(
                """
                version: 1
                rules:
                  - id: infra-disk
                    failure_class: infra.runner
                    excerpt_pattern: 'no space left on device'
                    notes: host disk full wins
                  - id: product-bootstrap
                    failure_class: product.bug
                    excerpt_pattern: 'k8s bootstrap'
                    notes: bootstrap failed
                """
            )
        )
        product_first = load_rules(
            _write_rules(
                """
                version: 1
                rules:
                  - id: product-bootstrap
                    failure_class: product.bug
                    excerpt_pattern: 'k8s bootstrap'
                    notes: bootstrap failed
                  - id: infra-disk
                    failure_class: infra.runner
                    excerpt_pattern: 'no space left on device'
                    notes: host disk full wins
                """
            )
        )
        context = RuleContext(
            signature_id="1111111111111111",
            excerpt="k8s bootstrap failed: no space left on device",
        )

        self.assertEqual(infra_first.match(context).failure_class, "infra.runner")
        self.assertEqual(product_first.match(context).failure_class, "product.bug")

    def test_default_rule_pack_orders_infra_above_product_wrappers(self) -> None:
        pack = load_rules(Path(__file__).resolve().parents[2] / "failure_rules.yaml")
        context = RuleContext(
            signature_id="1111111111111111",
            excerpt="k8s bootstrap failed: no space left on device",
        )

        self.assertEqual(pack.match(context).failure_class, "infra.runner")


if __name__ == "__main__":
    unittest.main()
