#
# Copyright 2026 Canonical, Ltd.
#
"""Persistence for the metrics pipeline.

Storage is deliberately split in two, because the two kinds of data have very
different volume and retention needs:

* **Committed** (``rollups/``, ``signatures/``, ``reports/``) -- small,
  append-mostly, kept forever. Roughly 1-3 MB/year, which is negligible
  against a 14 MB repository.
* **Artifact-only** (``runs/``) -- per-run job failure records including 4 KB
  evidence excerpts, ~100 KB per run. Retained for 90 days as a GitHub Actions
  artifact and never committed. At ~240 MB/year these would dominate every
  clone of the repository for very little added value.

The consequence, documented rather than hidden: ``metrics reclassify`` can only
replay the last 90 days. That matches GitHub's own job-log retention, so
nothing is meaningfully lost, and rollups preserve the long-term trend
regardless. Rollups older than the window are frozen, not retroactively
rewritten, and carry the ``taxonomy_version``/``ruleset_version`` that produced
them.

Data lives on a long-lived ``ci-metrics`` branch, *not* an orphan branch and
not ``main``:

* ``main`` is protected, so CI cannot push to it directly.
* Routing writes through a pull request would be circular -- a commit touching
  ``ci/`` triggers the full build-snap and e2e matrix, so the metrics system
  would be blocked from recording data by the very CI redness it exists to
  measure.
* Branching from ``main`` (rather than orphaning) keeps shared history, so
  ``git log``/``diff``/cherry-pick behave normally and the branch deltas
  against ``main`` instead of storing a second full tree.
"""

import gzip
import json
import logging
import os
import subprocess  # nosec B404 - git invocations are fixed argv lists, never shell
import tempfile
from pathlib import Path
from typing import Any, Dict, List, Optional

LOG = logging.getLogger(__name__)

METRICS_BRANCH = "ci-metrics"
DATA_ROOT = "ci/metrics-data"

ROLLUPS_DIR = "rollups"
SIGNATURES_DIR = "signatures"
REPORTS_DIR = "reports"
RUNS_DIR = "runs"

CATALOGUE_FILE = "catalogue.json"


def _run_git(args: List[str], cwd: Path) -> str:
    """Run a git command with a fixed argument vector (never via a shell)."""
    result = subprocess.run(  # nosec B603 - fixed argv, no shell, trusted args
        ["git", *args],
        cwd=str(cwd),
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise RuntimeError(f"git {' '.join(args)} failed: {result.stderr.strip()}")
    return result.stdout.strip()


def _atomic_write(path: Path, data: bytes) -> None:
    """Write a file atomically so an interrupted job cannot corrupt the store."""
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, tmp_name = tempfile.mkstemp(dir=str(path.parent), prefix=".tmp-")
    try:
        with os.fdopen(fd, "wb") as handle:
            handle.write(data)
        os.replace(tmp_name, path)
    except BaseException:
        if os.path.exists(tmp_name):
            os.unlink(tmp_name)
        raise


class MetricsStore:
    """Filesystem view of the metrics data tree.

    The store is a plain directory so it is trivially testable; git
    synchronisation is a separate, optional concern handled by
    :meth:`sync_pull` and :meth:`commit_and_push`.
    """

    def __init__(self, root: Path, repo_root: Optional[Path] = None) -> None:
        self.root = Path(root)
        self.repo_root = Path(repo_root) if repo_root else None

    # -- paths ------------------------------------------------------------

    def rollup_path(self, month: str) -> Path:
        """``month`` is ``YYYY-MM``."""
        return self.root / ROLLUPS_DIR / f"{month}.json"

    def catalogue_path(self) -> Path:
        return self.root / SIGNATURES_DIR / CATALOGUE_FILE

    def report_path(self, name: str) -> Path:
        return self.root / REPORTS_DIR / f"{name}.md"

    def run_path(
        self, workflow_slug: str, month: str, run_id: int, attempt: int
    ) -> Path:
        return (
            self.root / RUNS_DIR / workflow_slug / month / f"{run_id}-{attempt}.json.gz"
        )

    # -- run records (artifact-only, not committed) ------------------------

    def write_run(
        self,
        workflow_slug: str,
        month: str,
        run_id: int,
        attempt: int,
        record: Dict[str, Any],
    ) -> Path:
        path = self.run_path(workflow_slug, month, run_id, attempt)
        payload = json.dumps(record, indent=1, sort_keys=True).encode("utf-8")
        _atomic_write(path, gzip.compress(payload))
        return path

    def read_run(
        self, workflow_slug: str, month: str, run_id: int, attempt: int
    ) -> Optional[Dict[str, Any]]:
        return self.read_run_path(self.run_path(workflow_slug, month, run_id, attempt))

    def read_run_path(self, path: Path) -> Optional[Dict[str, Any]]:
        """Read a run record by path, for callers that located it via ``iter_runs``."""
        if not path.exists():
            return None
        with gzip.open(path, "rt", encoding="utf-8") as handle:
            return json.load(handle)

    def has_run(
        self, workflow_slug: str, month: str, run_id: int, attempt: int
    ) -> bool:
        """Used by ingest to skip already-recorded runs, making backfill resumable."""
        return self.run_path(workflow_slug, month, run_id, attempt).exists()

    def iter_runs(self, workflow_slug: Optional[str] = None) -> List[Path]:
        base = self.root / RUNS_DIR
        if workflow_slug:
            base = base / workflow_slug
        if not base.exists():
            return []
        return sorted(base.rglob("*.json.gz"))

    # -- committed data ---------------------------------------------------

    def read_json(self, path: Path, default: Any = None) -> Any:
        if not path.exists():
            return default
        with path.open("r", encoding="utf-8") as handle:
            return json.load(handle)

    def write_json(self, path: Path, payload: Any) -> Path:
        data = json.dumps(payload, indent=2, sort_keys=True).encode("utf-8")
        _atomic_write(path, data + b"\n")
        return path

    def read_catalogue(self) -> Dict[str, Any]:
        return self.read_json(self.catalogue_path(), default={}) or {}

    def write_catalogue(self, catalogue: Dict[str, Any]) -> Path:
        return self.write_json(self.catalogue_path(), catalogue)

    def read_rollup(self, month: str) -> Dict[str, Any]:
        return self.read_json(self.rollup_path(month), default={}) or {}

    def write_rollup(self, month: str, payload: Dict[str, Any]) -> Path:
        return self.write_json(self.rollup_path(month), payload)

    def write_report(self, name: str, markdown: str) -> Path:
        path = self.report_path(name)
        _atomic_write(path, markdown.encode("utf-8"))
        return path

    # -- git synchronisation ----------------------------------------------

    def sync_pull(self) -> None:
        """Fast-forward the local checkout of the metrics branch."""
        if not self.repo_root:
            return
        _run_git(["fetch", "origin", METRICS_BRANCH], cwd=self.repo_root)
        _run_git(["checkout", METRICS_BRANCH], cwd=self.repo_root)
        _run_git(["reset", "--hard", f"origin/{METRICS_BRANCH}"], cwd=self.repo_root)

    def commit_and_push(
        self, message: str, paths: List[Path], retries: int = 5
    ) -> bool:
        """Commit and push, rebasing on rejection.

        Concurrent writers are also serialised by a workflow-level concurrency
        group; this retry loop is the second line of defence for the case where
        a scheduled reconciliation and a ``workflow_run`` trigger overlap.

        Commits are signed off (DCO) as required by repository policy.
        """
        if not self.repo_root:
            LOG.info("No repo root configured; skipping commit")
            return False

        rel_paths = [str(p.relative_to(self.repo_root)) for p in paths]
        if not rel_paths:
            return False

        for attempt in range(retries):
            _run_git(["add", *rel_paths], cwd=self.repo_root)
            status = _run_git(
                ["status", "--porcelain", "--", *rel_paths], cwd=self.repo_root
            )
            if not status:
                LOG.info("No changes to commit")
                return False

            _run_git(["commit", "--signoff", "-m", message], cwd=self.repo_root)
            try:
                _run_git(
                    ["push", "origin", f"HEAD:{METRICS_BRANCH}"], cwd=self.repo_root
                )
                return True
            except RuntimeError as exc:
                LOG.warning("Push rejected (attempt %s): %s", attempt + 1, exc)
                _run_git(["fetch", "origin", METRICS_BRANCH], cwd=self.repo_root)
                _run_git(["rebase", f"origin/{METRICS_BRANCH}"], cwd=self.repo_root)

        raise RuntimeError(
            f"Failed to push to {METRICS_BRANCH} after {retries} attempts"
        )
