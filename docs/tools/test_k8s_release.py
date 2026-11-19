#
# Copyright 2026 Canonical, Ltd.
#
import sys
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "build-scripts"))
import k8s_release  # noqa: E402


class K8sReleaseExecTests(unittest.TestCase):
    def test_exec_runs_when_imported_as_module(self):
        proc = MagicMock(stdout="  origin/main\n", stderr="")
        with patch("k8s_release.subprocess.run", return_value=proc) as mock_run:
            stdout, stderr = k8s_release._exec(
                ["git", "branch", "-r"], capture_output=True
            )

        self.assertEqual(stdout, "  origin/main\n")
        self.assertEqual(stderr, "")
        mock_run.assert_called_once()


if __name__ == "__main__":
    unittest.main()
