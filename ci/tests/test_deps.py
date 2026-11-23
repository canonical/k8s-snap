#
# Copyright 2026 Canonical, Ltd.
#

import argparse
import importlib
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from cmds import deps  # noqa: E402


class StdlibImportTests(unittest.TestCase):
    def test_import_cmds_deps_without_packaging(self):
        blocked = {
            "packaging": None,
            "packaging.utils": None,
            "packaging.version": None,
        }
        with patch.dict(sys.modules, blocked):
            reloaded = importlib.reload(deps)
            parser = argparse.ArgumentParser()
            sub = parser.add_subparsers()
            reloaded.add_deps_cmds(sub)


class StanzaMergeTests(unittest.TestCase):
    def test_merge_replaces_older_version_with_override(self):
        base = (
            "mcp==1.29.0 \\\n"
            "    --hash=sha256:oldhash1 \\\n"
            "    --hash=sha256:oldhash2\n"
            "    # via semgrep\n"
        )
        overrides = (
            "mcp==1.30.0 \\\n"
            "    --hash=sha256:newhash1 \\\n"
            "    --hash=sha256:newhash2\n"
            "    # via -r ci/requirements-semgrep.overrides\n"
        )
        merged = deps._merge_requirement_stanzas(base, overrides)
        self.assertIn("mcp==1.30.0 \\", merged)
        self.assertIn("--hash=sha256:newhash1", merged)
        self.assertNotIn("mcp==1.29.0", merged)
        self.assertNotIn("--hash=sha256:oldhash1", merged)

    def test_merge_adds_missing_transitive_sub_dependency(self):
        base = (
            "semgrep==1.186.0 \\\n"
            "    --hash=sha256:semgrephash\n"
            "    # via -r ci/requirements-semgrep.in\n"
        )
        overrides = (
            "opentelemetry-exporter-otlp-proto-common==1.45.2 \\\n"
            "    --hash=sha256:commonhash\n"
            "    # via opentelemetry-exporter-otlp-proto-http\n"
        )
        merged = deps._merge_requirement_stanzas(base, overrides)
        self.assertIn("opentelemetry-exporter-otlp-proto-common==1.45.2 \\", merged)
        self.assertIn("--hash=sha256:commonhash", merged)
        self.assertIn("semgrep==1.186.0 \\", merged)

    def test_merge_never_downgrades_newer_version(self):
        base = "wcmatch==13.0.0 \\\n    --hash=sha256:newhash\n    # via semgrep\n"
        overrides = (
            "wcmatch==12.0.4 \\\n"
            "    --hash=sha256:overridehash\n"
            "    # via -r ci/requirements-semgrep.overrides\n"
        )
        merged = deps._merge_requirement_stanzas(base, overrides)
        self.assertIn("wcmatch==13.0.0 \\", merged)
        self.assertIn("--hash=sha256:newhash", merged)
        self.assertNotIn("wcmatch==12.0.4", merged)
        self.assertNotIn("--hash=sha256:overridehash", merged)

    def test_merge_preserves_alphabetical_order_and_hashes(self):
        base = (
            "zebra==1.0.0 \\\n"
            "    --hash=sha256:zebrahash\n"
            "    # via tool\n"
            "alpha==1.0.0 \\\n"
            "    --hash=sha256:alphahash\n"
            "    # via tool\n"
        )
        overrides = (
            "middle-pkg==2.0.0 \\\n"
            "    --hash=sha256:midhash1 \\\n"
            "    --hash=sha256:midhash2\n"
            "    # via override\n"
        )
        merged = deps._merge_requirement_stanzas(base, overrides)
        expected_order = ["alpha==1.0.0", "middle-pkg==2.0.0", "zebra==1.0.0"]
        actual_order = [
            line.split(" ")[0] for line in merged.splitlines() if "==" in line
        ]
        self.assertEqual(actual_order, expected_order)
        self.assertIn("--hash=sha256:midhash1", merged)
        self.assertIn("--hash=sha256:midhash2", merged)

    def test_merge_normalizes_pep503_names(self):
        base = "foo_bar==1.0.0 \\\n    --hash=sha256:old\n    # via dep\n"
        overrides = "foo-bar==1.2.0 \\\n    --hash=sha256:new\n    # via override\n"
        merged = deps._merge_requirement_stanzas(base, overrides)
        self.assertIn("foo-bar==1.2.0", merged)
        self.assertNotIn("foo_bar==1.0.0", merged)
        self.assertEqual(len([line for line in merged.splitlines() if "==" in line]), 1)


class StanzaParserEdgeCasesTests(unittest.TestCase):
    def test_parser_handles_markers_extras_and_comments(self):
        content = (
            "# Top level header\n"
            "\n"
            "foo[security]==1.0.0; sys_platform == 'linux' \\\n"
            "    --hash=sha256:foohash\n"
            "    # via direct\n"
            "\n"
            "# The following packages are considered to be unsafe in a requirements file:\n"
            "bar==2.0.0#inline-comment\n"
            "    --hash=sha256:barhash\n"
        )
        stanzas = deps._parse_requirement_stanzas(content)
        self.assertIn("foo", stanzas)
        self.assertIn("bar", stanzas)
        self.assertEqual(str(stanzas["foo"].version), "1.0.0")
        self.assertEqual(str(stanzas["bar"].version), "2.0.0")
        self.assertFalse(any(line.startswith("# Top") for line in stanzas["foo"].lines))
        self.assertFalse(
            any(line.startswith("# The following") for line in stanzas["foo"].lines)
        )


class UpdatePipLocksCommandTests(unittest.TestCase):
    def test_cmd_update_pip_locks_invalid_file(self):
        with tempfile.TemporaryDirectory() as tmp_root:
            args = argparse.Namespace(root=tmp_root, files=["not-a-file.txt"])
            ret = deps.cmd_update_pip_locks(args)
            self.assertEqual(ret, 1)

    def test_cmd_update_pip_locks_rejects_file_outside_root(self):
        with (
            tempfile.TemporaryDirectory() as tmp_root,
            tempfile.TemporaryDirectory() as outside_dir,
        ):
            outside_in = Path(outside_dir) / "requirements-outside.in"
            outside_in.write_text("pkg>=1.0\n")
            args = argparse.Namespace(root=tmp_root, files=[str(outside_in)])
            ret = deps.cmd_update_pip_locks(args)
            self.assertEqual(ret, 1)

    @patch("subprocess.run")
    def test_cmd_update_pip_locks_applies_overrides(self, mock_run):
        with tempfile.TemporaryDirectory() as tmp_root:
            root = Path(tmp_root)
            ci_dir = root / "ci"
            ci_dir.mkdir()
            in_file = ci_dir / "requirements-test.in"
            in_file.write_text("pkg>=1.0\n")
            overrides_file = ci_dir / "requirements-test.overrides"
            overrides_file.write_text("pkg>=1.1\n")
            txt_file = ci_dir / "requirements-test.txt"
            txt_file.write_text("pkg==1.0.0 \\\n    --hash=sha256:old\n")

            def mock_subproc(cmd, **kwargs):
                if cmd[0] == "pip-compile":
                    out_idx = cmd.index("--output-file") + 1
                    target = Path(cmd[out_idx])
                    if cmd[-1].endswith(".overrides"):
                        target.write_text("pkg==1.1.0 \\\n    --hash=sha256:new\n")
                    else:
                        target.write_text("pkg==1.0.0 \\\n    --hash=sha256:old\n")
                return MagicMock(returncode=0, stdout="", stderr="")

            mock_run.side_effect = mock_subproc
            args = argparse.Namespace(root=str(root), files=[str(in_file)])
            ret = deps.cmd_update_pip_locks(args)

            self.assertEqual(ret, 0)
            self.assertIn("pkg==1.1.0", txt_file.read_text())

    @patch("subprocess.run")
    def test_cmd_update_pip_locks_failure_preserves_lockfile(self, mock_run):
        with tempfile.TemporaryDirectory() as tmp_root:
            root = Path(tmp_root)
            ci_dir = root / "ci"
            ci_dir.mkdir()
            in_file = ci_dir / "requirements-test.in"
            in_file.write_text("pkg>=1.0\n")
            overrides_file = ci_dir / "requirements-test.overrides"
            overrides_file.write_text("pkg>=1.1\n")
            txt_file = ci_dir / "requirements-test.txt"
            txt_file.write_text("original-content\n")

            def mock_subproc(cmd, **kwargs):
                if cmd[0] == "pip-compile":
                    if cmd[-1].endswith(".overrides"):
                        return MagicMock(
                            returncode=1, stdout="", stderr="compile error"
                        )
                    out_idx = cmd.index("--output-file") + 1
                    Path(cmd[out_idx]).write_text(
                        "pkg==1.0.0 \\\n    --hash=sha256:old\n"
                    )
                return MagicMock(returncode=0, stdout="", stderr="")

            mock_run.side_effect = mock_subproc
            args = argparse.Namespace(root=str(root), files=[str(in_file)])
            ret = deps.cmd_update_pip_locks(args)

            self.assertEqual(ret, 1)
            self.assertEqual(txt_file.read_text(), "original-content\n")


if __name__ == "__main__":
    unittest.main()
