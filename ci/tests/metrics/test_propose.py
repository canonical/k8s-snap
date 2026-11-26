#
# Copyright 2026 Canonical, Ltd.
#
"""Rule proposer tests: the model proposes, deterministic checks decide."""

import json
import sqlite3
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.models import SCHEMA_SQL  # noqa: E402
from metrics.propose import (  # noqa: E402
    ProposalRejected,
    insert_rules,
    log_evidence,
    propose_rules,
    render_report,
    sanitise_rule,
    unknown_targets,
)
from metrics.rules import classify_signatures, load_rules  # noqa: E402

PACK = """\
##
## Test pack.
##

version: 4

rules:
  # --- external.dependency ---

  - id: external-snap-store
    failure_class: external.dependency
    subclass: snap_store
    excerpt_pattern: 'unable to contact snap store'
    examples:
      - 'error: unable to contact snap store'
    notes: The store was unreachable.

  # --- product.bug ---

  - id: product-node-never-ready
    failure_class: product.bug
    subclass: node_not_ready
    excerpt_pattern: 'in wait_until_k8s_ready$'
    examples:
      - 'File "util.py", line 4, in wait_until_k8s_ready'
    notes: A node never became Ready.
"""
SUBUNIT = "ModuleNotFoundError: No module named 'subunit'"
NODEID = "tests/test_a.py::test_a"
JOB_NAME = f"Integration (ubuntu:24.04, amd64, latest/edge) / {NODEID}"
GOOD_RULE = {
    "id": "test-missing-subunit",
    "failure_class": "test.bug",
    "subclass": "missing_dependency",
    "excerpt_pattern": "No module named 'subunit'",
    "examples": [SUBUNIT],
    "notes": "The test environment lacks a Python dependency.",
}


def _write_pack(text: str = PACK) -> Path:
    path = Path(tempfile.mkdtemp()) / "failure_rules.yaml"
    path.write_text(text, encoding="utf-8")
    return path


def _db(pack: Path) -> sqlite3.Connection:
    conn = sqlite3.connect(":memory:")
    conn.executescript(SCHEMA_SQL)
    conn.execute(
        "INSERT INTO workflows (workflow_id, name, file) VALUES (1, 'Nightly', 'nightly-test.yaml')"
    )
    conn.execute("""
        INSERT INTO runs (
            run_id, attempt, workflow_id, completed_at, conclusion, jobs_total,
            jobs_success, jobs_failure, jobs_skipped, jobs_cancelled,
            runner_minutes_total, runner_minutes_failed, ingest_version
        ) VALUES (100, 1, 1, '2026-11-20T00:00:00Z', 'failure', 9, 4, 5, 0, 0, 1, 1, 1)
        """)
    conn.execute(
        "INSERT INTO configs (config_id, os, arch, channel, flavor, substrate) "
        "VALUES (1, 'noble', 'amd64', 'latest/edge', 'classic', 'lxd')"
    )
    conn.execute(
        "INSERT INTO tests (test_id, nodeid, file) VALUES (1, ?, 'tests/test_a.py')",
        (NODEID,),
    )
    signatures = (
        ("aaaaaaaaaaaaaaaa", SUBUNIT, 3),
        ("bbbbbbbbbbbbbbbb", "stderr: error: unable to contact snap store", 1),
        ("cccccccccccccccc", "ImportError: No module named 'subunit' (vendored)", 1),
    )
    job_id = 0
    for signature_id, excerpt, jobs in signatures:
        conn.execute(
            """
            INSERT INTO signatures (
                signature_id, normalised_pattern, exemplar_excerpt,
                exemplar_job_id, normaliser_version, extractor_version
            ) VALUES (?, 'p', ?, ?, 4, 3)
            """,
            (signature_id, excerpt, job_id + 1),
        )
        for _ in range(jobs):
            job_id += 1
            conn.execute(
                """
                INSERT INTO jobs (
                    job_id, run_id, attempt, config_id, test_id, job_name,
                    html_url, failed_step_name, signature_id
                ) VALUES (?, 100, 1, 1, 1, ?, ?, 'Run test_a', ?)
                """,
                (
                    job_id,
                    JOB_NAME,
                    f"https://example.invalid/job/{job_id}",
                    signature_id,
                ),
            )
    classify_signatures(conn, load_rules(pack))
    return conn


class ScriptedModel:
    """Replies per target signature, in order; records what it was sent."""

    def __init__(self, replies: Dict[str, List[Any]]) -> None:
        self.replies = {key: list(values) for key, values in replies.items()}
        self.calls: List[List[Dict[str, str]]] = []

    def __call__(self, messages: List[Dict[str, str]]) -> str:
        self.calls.append(messages)
        target = next(key for key in self.replies if key in messages[1]["content"])
        reply = self.replies[target].pop(0)
        return reply if isinstance(reply, str) else json.dumps(reply)


def _rule(**changes: Any) -> Dict[str, Any]:
    return {"decision": "rule", "reason": "evidence", "rule": {**GOOD_RULE, **changes}}


class FakeGitHub:
    """Precision client: one successful run of the same test, and job logs."""

    def __init__(self, green_log: str = "all good\n", failed_log: str = "") -> None:
        self.green_log = green_log
        self.failed_log = failed_log

    def list_run_jobs(self, run_id: int, attempt: int) -> List[Dict[str, Any]]:
        return [
            {"id": 900 + n, "name": JOB_NAME, "conclusion": "success"} for n in range(4)
        ]

    def download_job_log(self, job_id: int) -> Tuple[int, Optional[str]]:
        return 200, self.green_log if job_id >= 900 else self.failed_log


class UnknownTargetsTest(unittest.TestCase):
    def test_lists_unclassified_signatures_most_jobs_first(self) -> None:
        targets = unknown_targets(_db(_write_pack()))

        self.assertEqual(
            [(t.signature_id, t.jobs) for t in targets],
            [("aaaaaaaaaaaaaaaa", 3), ("cccccccccccccccc", 1)],
        )
        self.assertEqual(targets[0].test_nodeids, (NODEID,))
        self.assertEqual(targets[0].substrates, ("lxd",))


class ProposeRulesTest(unittest.TestCase):
    def test_accepted_rule_lands_in_its_tier_and_covers_related_signatures(
        self,
    ) -> None:
        pack = _write_pack()
        conn = _db(pack)
        model = ScriptedModel({"aaaaaaaaaaaaaaaa": [_rule()]})

        proposals, text = propose_rules(conn, pack, model, None)

        # One accepted rule covers both subunit signatures, so the second is
        # never sent to the model.
        self.assertEqual(len(model.calls), 1)
        (proposal,) = proposals
        self.assertEqual(proposal.decision, "accepted")
        self.assertEqual((proposal.covered_signatures, proposal.covered_jobs), (2, 4))
        self.assertEqual(proposal.precision, "unchecked (offline)")
        self.assertIn("version: 5\n\nrules:", text)
        self.assertTrue(
            text.rstrip().endswith(
                "notes: The test environment lacks a Python dependency."
            )
        )
        new_pack = load_rules(_write_pack(text))
        self.assertEqual(new_pack.rules[-1].id, "test-missing-subunit")

    def test_rejection_is_fed_back_and_a_fixed_rule_is_accepted(self) -> None:
        pack = _write_pack()
        model = ScriptedModel(
            {
                "aaaaaaaaaaaaaaaa": [
                    _rule(examples=["ModuleNotFoundError: invented"]),
                    _rule(),
                ]
            }
        )

        proposals, _ = propose_rules(_db(pack), pack, model, None, limit=1)

        self.assertEqual(
            (proposals[0].decision, proposals[0].attempts), ("accepted", 2)
        )
        feedback = model.calls[1][-1]["content"]
        self.assertIn("Rejected: examples[0] is not copied from the evidence", feedback)

    def test_unusable_reply_is_retried_with_feedback(self) -> None:
        pack = _write_pack()
        model = ScriptedModel(
            {
                "aaaaaaaaaaaaaaaa": [
                    '{"decision": "rule", "rule": {"id": "test-',
                    _rule(),
                ]
            }
        )

        proposals, _ = propose_rules(_db(pack), pack, model, None, limit=1)

        self.assertEqual(
            (proposals[0].decision, proposals[0].attempts), ("accepted", 2)
        )
        self.assertIn(
            "Unusable reply: reply is not valid JSON", model.calls[1][-1]["content"]
        )

    def test_rules_that_move_owned_signatures_or_miss_the_target_are_rejected(
        self,
    ) -> None:
        cases = {
            "classified by rule 'infra-broken-env'": _rule(
                id="infra-broken-env",
                failure_class="infra.runner",
                excerpt_pattern="subunit|snap store",
            ),
            "would reclassify signatures other rules own: bbbbbbbbbbbbbbbb": _rule(
                id="infra-broken-env",
                failure_class="infra.runner",
                excerpt_pattern="subunit|^stderr: .*snap store",
            ),
            "does not match the target": _rule(
                excerpt_pattern="No module named 'subunit'",
                job={"substrate": "^multipass$"},
                examples=[{"excerpt": SUBUNIT, "job": {"substrate": "multipass"}}],
            ),
        }
        for expected, reply in cases.items():
            with self.subTest(expected=expected):
                pack = _write_pack()
                model = ScriptedModel({"aaaaaaaaaaaaaaaa": [reply]})

                proposals, text = propose_rules(
                    _db(pack), pack, model, None, limit=1, attempts=1
                )

                self.assertEqual(proposals[0].decision, "rejected")
                self.assertIn(expected, proposals[0].reason)
                self.assertEqual(text, PACK)

    def test_rule_matching_successful_jobs_is_rejected_as_imprecise(self) -> None:
        pack = _write_pack()
        github = FakeGitHub(
            green_log="ImportError: No module named 'subunit', using fallback\n",
            failed_log=f"=== FAILURES ===\n____ test_a ____\n{SUBUNIT}\n",
        )
        model = ScriptedModel({"aaaaaaaaaaaaaaaa": [_rule()]})

        proposals, _ = propose_rules(
            _db(pack), pack, model, github, limit=1, attempts=1
        )

        self.assertEqual(proposals[0].decision, "rejected")
        self.assertIn("matches 4/4 successful jobs", proposals[0].reason)

    def test_precise_rule_reports_its_green_hit_rate(self) -> None:
        pack = _write_pack()
        model = ScriptedModel({"aaaaaaaaaaaaaaaa": [_rule()]})

        proposals, _ = propose_rules(_db(pack), pack, model, FakeGitHub(), limit=1)

        self.assertEqual(proposals[0].precision, "0/4 green hits")

    def test_declines_and_model_errors_change_nothing(self) -> None:
        pack = _write_pack()
        model = ScriptedModel(
            {
                "aaaaaaaaaaaaaaaa": [
                    {"decision": "unknown", "reason": "no cause shown"}
                ],
                "cccccccccccccccc": ["this is not JSON", "still not JSON"],
            }
        )

        proposals, text = propose_rules(_db(pack), pack, model, None)

        self.assertEqual(
            [(p.decision, p.reason) for p in proposals],
            [
                ("unknown", "no cause shown"),
                ("error", "reply contains no JSON object: 'still not JSON'"),
            ],
        )
        self.assertEqual(text, PACK)
        self.assertIn(
            "0 of 2 unclassified signatures got a rule", render_report(proposals)
        )


class SanitiseRuleTest(unittest.TestCase):
    def test_unsafe_or_malformed_rules_are_rejected(self) -> None:
        pack = load_rules(_write_pack())
        cases = {
            "unsupported rule keys": {"signature_id": "aaaaaaaaaaaaaaaa"},
            "already exists": {"id": "external-snap-store"},
            "must start with 'test-'": {"id": "product-missing-subunit"},
            "nested quantifier": {"excerpt_pattern": "(subunit+)+"},
            "single-line": {"excerpt_pattern": "subunit\\nfoo"},
            "not a valid regex": {"excerpt_pattern": "subunit("},
            "not allowed": {"failure_class": "unknown"},
            "notes are required": {"notes": " "},
            "job keys": {"job": {"runner": "x"}},
        }
        for expected, change in cases.items():
            with self.subTest(expected=expected):
                with self.assertRaisesRegex(ProposalRejected, expected):
                    sanitise_rule({**GOOD_RULE, **change}, pack, SUBUNIT)

    def test_clean_rule_keeps_only_known_fields(self) -> None:
        pack = load_rules(_write_pack())

        rule = sanitise_rule({**GOOD_RULE, "confidence": 0.8}, pack, SUBUNIT)

        self.assertEqual(
            list(rule),
            [
                "id",
                "failure_class",
                "subclass",
                "confidence",
                "excerpt_pattern",
                "examples",
                "notes",
            ],
        )


class InsertRulesTest(unittest.TestCase):
    def test_rules_go_to_the_end_of_their_tier_and_version_bumps_once(self) -> None:
        external = {
            "id": "external-pypi",
            "failure_class": "external.dependency",
            "subclass": "pypi",
            "excerpt_pattern": "pypi",
            "examples": ["pypi timeout"],
            "notes": "PyPI.",
        }

        text = insert_rules(PACK, [external, GOOD_RULE])

        ids = [rule.id for rule in load_rules(_write_pack(text)).rules]
        self.assertEqual(
            ids,
            [
                "external-snap-store",
                "external-pypi",
                "product-node-never-ready",
                "test-missing-subunit",
            ],
        )
        self.assertIn("version: 5\n", text)
        # The new external rule sits above the next tier's banner comment.
        self.assertLess(text.index("external-pypi"), text.index("# --- product.bug"))


class LogEvidenceTest(unittest.TestCase):
    def test_keeps_the_failing_test_and_scrubs_secrets(self) -> None:
        log = textwrap.dedent("""\
            ##[group]Run pytest
            =================================== FAILURES ===================================
            ____ test_a ____
            token=ghp_abcdefghijklmnopqrstuvwxyz0123456789
            ModuleNotFoundError: No module named 'subunit'
            ____ test_b ____
            unrelated failure
            =========================== short test summary info ============================
            ##[error]Process completed with exit code 1.
            """)

        evidence = log_evidence(log)

        self.assertIn(SUBUNIT, evidence)
        self.assertNotIn("unrelated failure", evidence)
        self.assertNotIn("ghp_abcdefghijklmnopqrstuvwxyz0123456789", evidence)


if __name__ == "__main__":
    unittest.main()
