#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for the metrics CLI plumbing.

These cover the parts that sit between the modules and are easy to get wrong
without any test noticing -- resolving which stored records a command should
act on, and deciding which failures are still worth an API call.

Both behaviours pinned here were real bugs. Neither raised an error; they
degraded silently into "no data" and "wasted quota" respectively, which is the
failure mode a metrics system can least afford.
"""

import argparse
import gzip
import json

from cmds.metrics import _iter_target_records, _workflow_slug, cmd_audit_secrets
from metrics.gh import DEFAULT_REPO
from metrics.store import MetricsStore


def args(**kwargs):
    return argparse.Namespace(**kwargs)


def seed(root, workflow_slug, period, run_id):
    path = root / "runs" / workflow_slug / period / f"{run_id}-1.json.gz"
    path.parent.mkdir(parents=True, exist_ok=True)
    with gzip.open(path, "wt") as handle:
        json.dump({"run_id": run_id}, handle)
    return path


class TestWorkflowSlug:
    def test_alias_resolves_to_the_on_disk_slug(self):
        """Aliases map to filenames, but records are stored under a slug.

        Leaving the extension on made every filtered command match nothing and
        report it as "no data yet" -- a silent wrong answer, not an error.
        """
        assert _workflow_slug("nightly") == "nightly-test"
        assert _workflow_slug("pr") == "lint_and_integration"

    def test_filename_and_path_forms_also_resolve(self):
        assert _workflow_slug("weekly-test.yaml") == "weekly-test"
        assert _workflow_slug(".github/workflows/weekly-test.yml") == "weekly-test"

    def test_a_bare_slug_is_left_alone(self):
        assert _workflow_slug("nightly-test") == "nightly-test"

    def test_none_means_no_filter(self):
        assert _workflow_slug(None) is None


class TestTargetRecords:
    def test_workflow_filter_selects_only_that_workflow(self, tmp_path):
        store = MetricsStore(root=tmp_path)
        seed(tmp_path, "nightly-test", "2026-09", 1)
        seed(tmp_path, "weekly-test", "2026-09", 2)

        found = _iter_target_records(store, args(run_id=None, workflow="nightly"))
        assert [p.name for p in found] == ["1-1.json.gz"]

    def test_no_filter_selects_everything(self, tmp_path):
        store = MetricsStore(root=tmp_path)
        seed(tmp_path, "nightly-test", "2026-09", 1)
        seed(tmp_path, "weekly-test", "2026-09", 2)

        found = _iter_target_records(store, args(run_id=None, workflow=None))
        assert len(found) == 2

    def test_run_id_filter_wins_over_workflow(self, tmp_path):
        store = MetricsStore(root=tmp_path)
        seed(tmp_path, "nightly-test", "2026-09", 1)
        seed(tmp_path, "weekly-test", "2026-09", 2)

        found = _iter_target_records(store, args(run_id=[2], workflow="nightly"))
        assert [p.name for p in found] == ["2-1.json.gz"]

    def test_since_bounds_the_window_by_period(self, tmp_path, monkeypatch):
        import cmds.metrics as metrics_cli

        store = MetricsStore(root=tmp_path)
        seed(tmp_path, "nightly-test", "2026-06", 1)
        seed(tmp_path, "nightly-test", "2026-09", 2)

        class FakeNow:
            @staticmethod
            def strftime(fmt):
                return "2026-08"

        monkeypatch.setattr(metrics_cli, "_parse_since", lambda value: FakeNow())
        found = _iter_target_records(
            store, args(run_id=None, workflow=None, since="30d")
        )
        assert [p.name for p in found] == ["2-1.json.gz"]


class TestAuditSecretsGate:
    """The audit is the only thing between raw log text and a public branch.

    It runs in CI as a hard gate before anything is committed or uploaded, so
    its failure modes are worth pinning explicitly: a gate that cannot fail is
    indistinguishable from no gate at all.
    """

    def audit(self, root):
        return cmd_audit_secrets(
            args(data_dir=str(root), repo=DEFAULT_REPO, verbose=False)
        )

    def test_empty_data_dir_fails_rather_than_passing_vacuously(self, tmp_path):
        """ "No secrets in zero files" is a misconfigured path, not a pass.

        The workflow points the audit at a path built from the checkout
        layout. If that path were ever wrong the audit would scan nothing,
        report success, and wave through whatever the commit step actually
        committed.
        """
        assert self.audit(tmp_path / "nope") == 1

    def test_secret_in_a_rendered_report_is_caught(self, tmp_path):
        """Reports are committed as markdown, not JSON.

        The scan used to glob '*.json', so every rendered report went to the
        branch unscanned. Nothing renders excerpts into a report today; this
        pins the gate so that staying true is not a precondition for safety.
        """
        report = tmp_path / "reports" / "latest.md"
        report.parent.mkdir(parents=True)
        report.write_text("top signature\nghp_" + "A" * 26 + "\n")
        assert self.audit(tmp_path) == 1

    def test_clean_data_passes(self, tmp_path):
        rollup = tmp_path / "rollups" / "2026-09.json"
        rollup.parent.mkdir(parents=True)
        rollup.write_text('{"period": "2026-09", "m9_unclassified_rate": 3.3}\n')
        assert self.audit(tmp_path) == 0
