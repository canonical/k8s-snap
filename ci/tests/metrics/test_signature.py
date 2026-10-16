#
# Copyright 2026 Canonical, Ltd.
#
"""Signature extraction and normalisation tests."""

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.signature import error_lines, extract, fingerprint, normalise  # noqa: E402


class SignatureTest(unittest.TestCase):
    def test_failures_block_extracts_innermost_exception(self) -> None:
        log = """
=== FAILURES ===
Traceback (most recent call last):
  File "/home/runner/work/k8s-snap/k8s-snap/tests/test_service.py", line 40, in wait
    assert_service()
  File "/home/runner/work/k8s-snap/k8s-snap/tests/test_service.py", line 41, in assert_service
    assert False
AssertionError: Service kube-proxy should be active, but it is inactive

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/home/runner/work/k8s-snap/k8s-snap/tests/test_service.py", line 50, in test_service
    retryer(fn)
tenacity.RetryError: RetryError[<Future at 0xabc state=finished raised AssertionError>]
=== short test summary info ===
FAILED tests/test_service.py::test_service - tenacity.RetryError
"""

        excerpt = extract(log)

        self.assertEqual(excerpt.source, "failures-block")
        self.assertIn(
            "AssertionError: Service kube-proxy should be active", excerpt.text
        )
        self.assertNotIn("tenacity.RetryError", excerpt.text)
        # Two root-cause frames plus the test frame that reached the helper.
        self.assertEqual(excerpt.text.count('File "'), 3)
        self.assertTrue(excerpt.text.splitlines()[0].endswith("in test_service"))

    def test_helper_root_cause_is_anchored_to_the_calling_test(self) -> None:
        def log(test: str) -> str:
            return """
=== FAILURES ===
____ TEST ____
Traceback (most recent call last):
  File "/home/runner/work/k8s-snap/k8s-snap/tests/test_util/util.py", line 133, in _exec
    assert self._condition(resp), "Failed to meet condition"
AssertionError: Failed to meet condition

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/home/runner/work/k8s-snap/k8s-snap/tests/test_upgrade.py", line 454, in TEST
    ).exec(
tenacity.RetryError: RetryError[<Future at 0xff6c4b66c830 state=finished>]
____ test_other ____
Traceback (most recent call last):
  File "/home/runner/work/k8s-snap/k8s-snap/tests/test_other.py", line 9, in test_other
    boom()
ValueError: unrelated
""".replace(
                "TEST", test
            )

        first = fingerprint(log("test_upgrade_inplace"))
        second = fingerprint(log("test_downgrade_rollback"))

        self.assertIn("in test_upgrade_inplace", first.excerpt)
        self.assertNotIn("test_other", first.excerpt)
        self.assertNotEqual(first.signature_id, second.signature_id)

    def test_post_failure_steps_do_not_supply_the_excerpt(self) -> None:
        log = """
##[group]Run make linkcheck
make: *** [Makefile:106: linkcheck] Error 1
Command 'make -f Makefile linkcheck' returned non-zero exit status 2.
##[error]Process completed with exit code 1.
##[group]Run actions/upload-artifact@v4
##[error]No files were found with the provided path
Post job cleanup.
[command]/usr/bin/git version
git version 2.55.0
"""

        excerpt = extract(log)

        self.assertEqual(excerpt.source, "diagnostic-line")
        self.assertEqual(
            excerpt.text,
            "Command 'make -f Makefile linkcheck' returned non-zero exit status 2.",
        )

    def test_retry_wrapper_verdict_falls_through_to_the_tool_error(self) -> None:
        log = """
##[group]Run for attempt in $(seq 1 "${RETRY_MAX_ATTEMPTS}"); do
Reading package lists...
E: Unable to locate package tox
Attempt 1/5 failed, retrying in 5s...
E: Unable to locate package tox
##[error]Failed after 5 attempts
##[error]Process completed with exit code 1.
"""

        excerpt = extract(log)

        self.assertEqual(excerpt.source, "diagnostic-line")
        self.assertEqual(excerpt.text, "E: Unable to locate package tox")

    def test_error_lines_lists_error_shaped_lines_of_a_whole_log(self) -> None:
        log = """
2026-10-15T15:14:45.1Z Attempt 1/40 failed. Error: Failed to meet condition
2026-10-15T15:14:46.1Z AssertionError: Failed to meet condition
2026-10-15T15:14:47.1Z ##[error]Process completed with exit code 1.
2026-10-15T15:14:48.1Z Error: node not ready
"""

        lines = error_lines(log)

        self.assertIn("AssertionError: Failed to meet condition", lines)
        self.assertIn("Error: node not ready", lines)
        self.assertNotIn("##[error]Process completed with exit code 1.", lines)

    def test_errors_block_precedes_summary_for_fixture_failures(self) -> None:
        log = """
=== ERRORS ===
ERROR at setup of test_smoke
Traceback (most recent call last):
  File "/home/runner/work/k8s-snap/k8s-snap/tests/conftest.py", line 30, in cluster
    bootstrap()
AssertionError: fixture bootstrap failed
=== short test summary info ===
ERROR tests/test_a.py::test_smoke - tenacity.RetryError
ERROR tests/test_b.py::test_smoke - tenacity.RetryError
"""

        excerpt = extract(log)

        self.assertEqual(excerpt.source, "errors-block")
        self.assertIn("AssertionError: fixture bootstrap failed", excerpt.text)
        self.assertNotIn("tests/test_a.py", excerpt.text)

    def test_summary_precedes_informative_error_marker(self) -> None:
        log = """
=== short test summary info ===
FAILED tests/test_smoke.py::test_smoke - AssertionError: bad
##[error]real error
"""

        excerpt = extract(log)

        self.assertEqual(excerpt.source, "summary-line")
        self.assertIn("FAILED tests/test_smoke.py::test_smoke", excerpt.text)

    def test_useless_exit_code_error_marker_falls_through_to_tail(self) -> None:
        log = "context line\n##[error]Process completed with exit code 1."

        excerpt = extract(log)

        self.assertEqual(excerpt.source, "log-tail")
        self.assertIn("context line", excerpt.text)

    def test_normalise_collapses_only_named_variation_and_preserves_exit_code_digits(
        self,
    ) -> None:
        first = normalise(
            "k8s-integration-4711-a3f9 failed from 10.1.2.3 at "
            "2026-09-30T12:08:26Z in /home/runner/work/k8s-snap/k8s-snap/file.py "
            "line 412 after 3 restarts with exit code 1"
        )
        second = normalise(
            "k8s-integration-9999-deadbeef failed from 10.9.8.7 at "
            "2026-10-01T12:08:26Z in /home/ubuntu/actions-runner/_work/k8s-snap/k8s-snap/file.py "
            "line 900 after 10 restarts with exit code 2"
        )

        self.assertIn("exit code 1", first)
        self.assertIn("exit code 2", second)
        self.assertNotEqual(first, second)
        self.assertEqual(
            first.replace("exit code 1", "exit code <x>"),
            second.replace("exit code 2", "exit code <x>"),
        )

    def test_normalise_keeps_the_repo_relative_file_name(self) -> None:
        hosted = normalise(
            '  File "/home/runner/work/k8s-snap/k8s-snap/tests/.tox/lib/python3.12/'
            'site-packages/tenacity/__init__.py", line 382, in __call__'
        )
        self_hosted = normalise(
            '  File "/home/ubuntu/actions-runner/_work/k8s-snap/k8s-snap/tests/.tox/'
            'lib/python3.10/site-packages/tenacity/__init__.py", line 9, in __call__'
        )

        self.assertEqual(hosted, self_hosted)
        self.assertEqual(
            hosted,
            'File "<workspace>/tests/.tox/lib/python3.<n>/site-packages/tenacity/'
            '__init__.py", line <n>, in __call__',
        )

    def test_fingerprint_scrubs_before_hash_and_store(self) -> None:
        token = "abcdef.0123456789abcdef"
        result = fingerprint(
            "=== short test summary info ===\n"
            f"FAILED tests/test_join.py::test_join - sudo k8s join {token} failed"
        )

        self.assertNotIn(token, result.excerpt)
        self.assertNotIn(token, result.normalised)
        self.assertEqual(result.source, "summary-line")


if __name__ == "__main__":
    unittest.main()
