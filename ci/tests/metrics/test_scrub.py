#
# Copyright 2026 Canonical, Ltd.
#
"""Secret scrubber tests."""

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.scrub import REDACTED, find_secrets, scrub  # noqa: E402


class ScrubTest(unittest.TestCase):
    def test_join_token_is_excised_mid_line_and_command_survives(self) -> None:
        token = "AbCdEfGhIjKlMnOpQrStUvWxYz0123456789-_AbCdEfGhIjKlMnOpQrStUvWxYz012345"
        line = f"sudo k8s join {token} --worker --name node-a"

        safe = scrub(line)

        self.assertEqual(safe, f"sudo k8s join {REDACTED} --worker --name node-a")
        self.assertEqual(find_secrets(safe), ())

    def test_assignment_secret_is_redacted_in_place(self) -> None:
        safe = scrub("running command with token=abc123def456 next")

        self.assertEqual(safe, f"running command with token={REDACTED} next")
        self.assertEqual(find_secrets(safe), ())

    def test_audit_reports_unscrubbed_secret(self) -> None:
        findings = find_secrets("Authorization: Bearer abcdefghijklmnop")

        self.assertEqual(len(findings), 1)
        self.assertEqual(findings[0].line_number, 1)
        self.assertEqual(findings[0].rule_name, "authorization-header")


if __name__ == "__main__":
    unittest.main()
