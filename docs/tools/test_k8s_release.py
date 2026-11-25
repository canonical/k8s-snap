#
# Copyright 2026 Canonical, Ltd.
#
import shutil
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

from packaging.version import Version

REPO_ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(REPO_ROOT / "build-scripts"))
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


class PinReleaseBranchVersionsTests(unittest.TestCase):
    def test_pins_real_updater_to_release_minor_series(self):
        with tempfile.TemporaryDirectory() as tmp:
            updater = Path(tmp, "build-scripts/hack/update-component-versions.py")
            helm = Path(tmp, "build-scripts/components/helm/version")
            updater.parent.mkdir(parents=True)
            helm.parent.mkdir(parents=True)
            shutil.copy(
                REPO_ROOT / "build-scripts/hack/update-component-versions.py", updater
            )
            helm.write_text("v4.3.2\n")

            k8s_release.pin_release_branch_versions(tmp, Version("1.38.0"))

            content = updater.read_text()
            self.assertIn(
                'KUBERNETES_VERSION_MARKER = "https://dl.k8s.io/release/stable-1.38.txt"',
                content,
            )
            self.assertIn(
                'HELM_BRANCH, HELM_RELEASE_SEMVER = "main", Version("4.3")', content
            )


if __name__ == "__main__":
    unittest.main()
