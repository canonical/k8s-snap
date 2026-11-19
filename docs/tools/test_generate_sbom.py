#
# Copyright 2026 Canonical, Ltd.
#
import importlib.util
import sys
import types
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

import yaml


def _load_generate_sbom():
    repo_root = Path(__file__).resolve().parents[2]
    hack_dir = repo_root / "build-scripts" / "hack"
    sys.path.insert(0, str(hack_dir))
    injected_tenacity = False
    if "tenacity" not in sys.modules:
        fake_tenacity = types.ModuleType("tenacity")
        fake_tenacity.retry = lambda *a, **kw: (lambda fn: fn)
        fake_tenacity.before_sleep_log = MagicMock()
        fake_tenacity.retry_if_exception_type = MagicMock()
        fake_tenacity.stop_after_attempt = MagicMock()
        fake_tenacity.wait_exponential = MagicMock()
        fake_tenacity.wait_fixed = MagicMock()
        sys.modules["tenacity"] = fake_tenacity
        injected_tenacity = True
    try:
        spec = importlib.util.spec_from_file_location(
            "generate_sbom", hack_dir / "generate-sbom.py"
        )
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        return module
    finally:
        sys.path.pop(0)
        if injected_tenacity:
            sys.modules.pop("tenacity", None)
            sys.modules.pop("util", None)


class GenerateSbomTests(unittest.TestCase):
    def test_snapcraft_go_components_include_all_bom_parts(self):
        repo_root = Path(__file__).resolve().parents[2]
        snapcraft_yaml = yaml.safe_load(
            (repo_root / "snap" / "snapcraft.yaml").read_text()
        )
        bom_after = set(snapcraft_yaml["parts"]["bom"]["after"])
        expected_go_components = {
            part
            for part in bom_after
            if (repo_root / "build-scripts" / "components" / part / "version").exists()
        }

        module = _load_generate_sbom()
        self.assertEqual(set(module.SNAPCRAFT_GO_COMPONENTS), expected_go_components)

        manifest = {"snap": {"external": {}}}
        extra_files = {}
        fake_dir = Path("/tmp/fake-go-repo")
        with (
            patch.object(module.util, "git_repo") as mock_git_repo,
            patch.object(module.util, "parse_output", return_value="deadbeef"),
            patch.object(
                module.util,
                "read_file",
                side_effect=lambda p: (
                    p.read_text().strip()
                    if Path(p).exists()
                    else "module example.com/fake\n"
                ),
            ),
        ):
            mock_git_repo.return_value.__enter__.return_value = fake_dir
            module.go_components(manifest, extra_files)

        self.assertIn("etcd", manifest["snap"]["external"])
        self.assertIn("etcd/go.mod", extra_files)


if __name__ == "__main__":
    unittest.main()
