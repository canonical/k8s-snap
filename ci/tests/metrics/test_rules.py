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


SUBSTRATE_SPLIT = """
version: 1
rules:
  - id: multipass-never-ready
    failure_class: infra.provisioning
    excerpt_pattern: 'Failed to meet condition'
    job:
      substrate: '^multipass$'
    notes: the same text means provisioning on Multipass
  - id: lxd-never-ready
    failure_class: product.bug
    excerpt_pattern: 'Failed to meet condition'
    job:
      substrate: '^lxd$'
    notes: and a product failure on LXD
"""
DEFAULT_PACK = Path(__file__).resolve().parents[2] / "failure_rules.yaml"


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
        pack = load_rules(_write_rules(SUBSTRATE_SPLIT))

        classify_signatures(conn, pack)
        row = conn.execute(
            "SELECT failure_class FROM v_job_class WHERE job_id = 1"
        ).fetchone()

        self.assertEqual(row, ("unknown",))

    def test_same_signature_splits_by_substrate(self) -> None:
        conn = _db(("multipass", "lxd"))
        pack = load_rules(_write_rules(SUBSTRATE_SPLIT))

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
        pack = load_rules(DEFAULT_PACK)
        context = RuleContext(
            signature_id="1111111111111111",
            excerpt="k8s bootstrap failed: no space left on device",
        )

        self.assertEqual(pack.match(context).failure_class, "infra.runner")

    def test_example_must_be_classified_by_its_own_rule(self) -> None:
        typo = _write_rules(
            """
            version: 1
            rules:
              - id: snap-store
                failure_class: external.dependency
                excerpt_pattern: 'unable to contact snap stroe'
                examples: ['stderr: error: unable to contact snap store']
                notes: typo in the pattern
            """
        )
        shadowed = _write_rules(
            """
            version: 1
            rules:
              - id: broad
                failure_class: product.bug
                excerpt_pattern: 'snap refresh'
                notes: too broad, listed first
              - id: snap-store
                failure_class: external.dependency
                excerpt_pattern: 'unable to contact snap store'
                examples:
                  - "Command 'snap refresh k8s' failed\\nstderr: error: unable to contact snap store"
                notes: never reached
            """
        )

        with self.assertRaisesRegex(RulePackError, "classified by no rule"):
            load_rules(typo)
        with self.assertRaisesRegex(RulePackError, "classified by rule 'broad'"):
            load_rules(shadowed)

    def test_examples_can_carry_job_metadata(self) -> None:
        path = _write_rules(
            SUBSTRATE_SPLIT.replace(
                "    notes: and a product failure on LXD",
                "    notes: and a product failure on LXD\n"
                "    examples:\n"
                "      - excerpt: 'AssertionError: Failed to meet condition'\n"
                "        job: {substrate: lxd}",
            )
        )

        pack = load_rules(path)

        self.assertEqual(pack.rules[1].examples[0].substrate, "lxd")

    def test_narrowed_rule_unclassifies_what_it_used_to_match(self) -> None:
        conn = _db(("lxd",))
        classify_signatures(conn, load_rules(_write_rules(SUBSTRATE_SPLIT)))
        narrowed = SUBSTRATE_SPLIT.replace(
            "'Failed to meet condition'\n    job:\n      substrate: '^lxd$'",
            "'Failed to meet condition: upgrade'\n    job:\n      substrate: '^lxd$'",
        )

        classify_signatures(conn, load_rules(_write_rules(narrowed)))

        self.assertEqual(
            conn.execute("SELECT failure_class FROM v_job_class").fetchall(),
            [("unknown",)],
        )

    def test_new_ruleset_supersedes_older_rows_but_keeps_them(self) -> None:
        conn = _db(("lxd",))
        classify_signatures(conn, load_rules(_write_rules(SUBSTRATE_SPLIT)))
        conn.execute(
            """
            INSERT INTO signature_classifications (
                signature_id, substrate, ruleset_version, failure_class,
                rule_id, confidence, classified_by, classified_at
            ) VALUES ('8c531e6ba49b3382', 'lxd', 1, 'test.bug', NULL, 1, 'human', '')
            """
        )
        empty_v2 = _write_rules(
            """
            version: 2
            rules:
              - id: unrelated
                failure_class: infra.runner
                excerpt_pattern: 'no space left on device'
                notes: matches nothing stored
            """
        )

        classify_signatures(conn, load_rules(empty_v2))

        self.assertEqual(
            conn.execute("SELECT failure_class FROM v_job_class").fetchall(),
            [("test.bug",)],
        )
        conn.execute(
            "DELETE FROM signature_classifications WHERE classified_by='human'"
        )
        self.assertEqual(
            conn.execute("SELECT failure_class FROM v_job_class").fetchall(),
            [("unknown",)],
        )
        self.assertEqual(
            conn.execute(
                "SELECT ruleset_version, rule_id FROM signature_classifications"
            ).fetchall(),
            [(1, "lxd-never-ready")],
        )

    def test_default_pack_follows_the_authoring_contract(self) -> None:
        pack = load_rules(DEFAULT_PACK)
        tiers = [
            "infra.runner",
            "infra.provisioning",
            "external.dependency",
            "ci.config",
            "product.bug",
            "test.bug",
        ]

        for rule in pack.rules:
            with self.subTest(rule=rule.id):
                self.assertTrue(rule.examples, "every rule needs a real example")
                self.assertFalse(
                    rule.signature_ids,
                    "signature IDs change with the extractor; match the excerpt",
                )
                self.assertTrue(rule.subclass, "every rule needs a subclass")
        order = [tiers.index(rule.failure_class) for rule in pack.rules]
        self.assertEqual(order, sorted(order), "rules must be ordered by tier")


if __name__ == "__main__":
    unittest.main()
