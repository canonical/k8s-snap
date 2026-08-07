#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for the pipeline orchestrator.

Two contracts live here. First the ordering: no product code is touched until an
end-to-end test exists and has been observed to fail, so a stage that cannot
produce a red test must stop before diagnose and fix ever run. Second the PR
seam: ``_open_pr`` pushes the branch the skills committed and opens a draft PR,
degrading gracefully (reuse an existing PR, skip a branch that was never
committed, swallow a GitHub error) and opening a test-only PR when the fix
itself failed.
"""

from __future__ import annotations

import subprocess

from triage_bot.context import ActionContext
from triage_bot.github import GitHubError
from triage_bot.handlers.base import IssueContext, Runtime
from triage_bot.pipeline import _cleanup, _open_pr, run_pipeline
from triage_bot.schema import (
    DiagnoseResult,
    FixResult,
    ReproduceResult,
    ReproducerResult,
    VerifyResult,
)
from triage_bot.tests.doubles import FakeGitHub

ISSUE = 42
BRANCH = f"triage/fix-{ISSUE}"
TEST_PATH = "tests/integration/tests/test_dns.py"


class _Rt:
    """Minimal runtime: ``_open_pr`` only reaches ``rt.gh`` and the worktree."""

    def __init__(self, gh):
        self.gh = gh

    def worktree(self, number: int):
        return f"/repo/.triage/issue-{number}/checkout"


def _issue():
    return IssueContext(number=ISSUE, title="dns broke", body="x")


def _fix(fixed=True):
    return FixResult(
        fixed=fixed, commit_message="fix: resolve dns crashloop" if fixed else None
    )


def _reproducer(red=True):
    return ReproducerResult(
        test_path=TEST_PATH,
        test_selector="test_coredns_spread",
        fails_before_fix=red,
        failure_output="AssertionError: both replicas on one node",
    )


# --- stage ordering ---------------------------------------------------------


def _pipeline_runtime(tmp_path, **ctx_kwargs):
    ctx = ActionContext(workdir_root=str(tmp_path), **ctx_kwargs)
    return Runtime(ctx=ctx, gh=FakeGitHub(), pipeline=run_pipeline)


def _stub_stages(monkeypatch, results):
    """Record the steps run, returning a canned result for each.

    The real worktree is stubbed: these tests pin ordering, and checking out a
    second copy of the repository per test would dominate the suite. Cleanup
    is stubbed too -- otherwise every test would shell out for real to
    ``hack/cluster-up.sh --destroy``, which on a bare CI runner tries to
    ``sudo snap install lxd``. ``_cleanup`` has its own dedicated tests below,
    with ``subprocess.run`` mocked instead of the whole function.
    """
    ran: list[str] = []

    def fake_run_skill(*, step, **_):
        ran.append(step)
        return results[step]

    monkeypatch.setattr("triage_bot.pipeline.run_skill", fake_run_skill)
    monkeypatch.setattr("triage_bot.pipeline.ensure_worktree", lambda path, _: path)
    monkeypatch.setattr("triage_bot.pipeline._cleanup", lambda checkout: None)
    return ran


def _all_green():
    return {
        "reproduce": ReproduceResult(reproducible=True, evidence="observed restart"),
        "verify": VerifyResult(verdict="bug", confidence="high"),
        "reproducer": _reproducer(),
        "diagnose": DiagnoseResult(confidence="high"),
        "fix": _fix(),
    }


def test_stages_run_test_first_then_diagnose_and_fix(tmp_path, monkeypatch):
    ran = _stub_stages(monkeypatch, _all_green())

    result = run_pipeline(_pipeline_runtime(tmp_path), _issue())

    assert ran == ["reproduce", "verify", "reproducer", "diagnose", "fix"]
    assert result.completed_stage == "fix"
    assert result.fixed
    assert result.test_path == TEST_PATH


def test_stages_log_a_one_line_narrative_at_info_and_nothing_below(
    tmp_path, monkeypatch, caplog
):
    # The ask this pins: a human watching the default log level sees
    # "[stage] outcome" once per stage, not the per-tool-call trace (that
    # lives at DEBUG inside RunLogger, exercised separately in test_runlog).
    _stub_stages(monkeypatch, _all_green())

    with caplog.at_level("INFO", logger="triage_bot"):
        run_pipeline(_pipeline_runtime(tmp_path), _issue())

    lines = [r.message for r in caplog.records]
    assert "[pipeline] issue #42: starting" in lines
    assert "[reproduce] running" in lines
    assert "[reproduce] reproducible=True" in lines
    assert "[verify] verdict=bug confidence=high" in lines
    assert any(line.startswith("[reproducer] test=") for line in lines)
    assert "[diagnose] confidence=high" in lines
    assert "[fix] fixed=True" in lines
    assert any(line.startswith("[pipeline] issue #42: done") for line in lines)


def test_no_code_is_touched_without_a_failing_test(tmp_path, monkeypatch):
    # The reproducer step could not make the bug fail as a test. Diagnose and
    # fix must not run: there would be nothing to prove a fix against.
    results = _all_green() | {"reproducer": _reproducer(red=False)}
    ran = _stub_stages(monkeypatch, results)

    result = run_pipeline(_pipeline_runtime(tmp_path), _issue())

    assert ran == ["reproduce", "verify", "reproducer"]
    assert result.completed_stage == "reproducer"
    assert not result.fixed


def test_unreproducible_issue_stops_before_verify(tmp_path, monkeypatch):
    results = _all_green() | {"reproduce": ReproduceResult(reproducible=False)}
    ran = _stub_stages(monkeypatch, results)

    result = run_pipeline(_pipeline_runtime(tmp_path), _issue())

    assert ran == ["reproduce"]
    assert not result.reproducible


def test_intended_behaviour_stops_before_writing_a_test(tmp_path, monkeypatch):
    results = _all_green() | {"verify": VerifyResult(verdict="intended-behavior")}
    ran = _stub_stages(monkeypatch, results)

    result = run_pipeline(_pipeline_runtime(tmp_path), _issue())

    assert ran == ["reproduce", "verify"]
    assert result.verdict == "intended-behavior"


def test_no_pr_is_opened_unless_auto_pr_is_set(tmp_path, monkeypatch):
    _stub_stages(monkeypatch, _all_green())
    rt = _pipeline_runtime(tmp_path)

    assert run_pipeline(rt, _issue()).pr_url is None
    assert rt.gh.pulls_created == []


# --- the PR seam ------------------------------------------------------------


def test_open_pr_pushes_committed_branch_and_opens_draft():
    gh = FakeGitHub(local_branches=[BRANCH])

    url = _open_pr(_Rt(gh), _issue(), _fix(), _reproducer())

    assert BRANCH in gh.pushed_branches
    assert len(gh.pulls_created) == 1
    assert gh.pulls_created[0]["head"] == BRANCH
    assert url == gh.pulls_created[0]["html_url"]


def test_failed_fix_still_opens_a_pr_carrying_the_test():
    # The whole point of committing the test separately: a maintainer gets a
    # runnable reproducer even though the bot could not fix the bug.
    gh = FakeGitHub(local_branches=[BRANCH])

    url = _open_pr(_Rt(gh), _issue(), _fix(fixed=False), _reproducer())

    assert url is not None
    pr = gh.pulls_created[0]
    assert pr["title"].startswith("test:")
    # It must not claim to close an issue it did not fix.
    assert "Closes #" not in pr["body"]
    assert f"Refs #{ISSUE}" in pr["body"]


def test_open_pr_updates_branch_then_reuses_existing_pr():
    gh = FakeGitHub(local_branches=[BRANCH])
    existing = gh.create_pull_request(
        head=BRANCH, base="main", title="fix", body="prior run"
    )

    url = _open_pr(_Rt(gh), _issue(), _fix(), _reproducer())

    # A re-run force-pushes the new commit first (so the open PR is never
    # stale), then reuses the existing PR rather than opening a duplicate.
    assert gh.pushed_branches == [BRANCH]
    assert url == existing["html_url"]
    assert len(gh.pulls_created) == 1


def test_open_pr_skips_uncommitted_branch():
    # The skills left no local branch: no PR, no crash.
    gh = FakeGitHub(local_branches=[])

    url = _open_pr(_Rt(gh), _issue(), _fix(), _reproducer())

    assert url is None
    assert gh.pulls_created == []


def test_open_pr_swallows_push_error():
    class _RaisingGitHub(FakeGitHub):
        def push_branch(self, branch: str, cwd: str) -> bool:
            raise GitHubError("push rejected")

    gh = _RaisingGitHub(local_branches=[BRANCH])

    # A GitHub hiccup must degrade to "no PR", never propagate and get the fix
    # misreported as an error by the handler's generic except.
    assert _open_pr(_Rt(gh), _issue(), _fix(), _reproducer()) is None
    assert gh.pulls_created == []


# --- cluster cleanup -------------------------------------------------------
#
# ``subprocess.run`` is mocked here rather than ``_cleanup`` itself, so these
# exercise the real primary-tree lookup and log lines without ever shelling
# out to real git/bash/lxc (which, on a bare CI runner, would try to
# ``sudo snap install lxd``).


def _fake_primary_with_script(tmp_path):
    primary = tmp_path / "primary"
    script = primary / "hack" / "cluster-up.sh"
    script.parent.mkdir(parents=True)
    script.write_text("#!/bin/bash\n", encoding="utf-8")
    checkout = tmp_path / "issue-1" / "checkout"
    checkout.mkdir(parents=True)
    return checkout, primary, script


def test_cleanup_destroys_via_the_primary_checkouts_script(
    tmp_path, monkeypatch, caplog
):
    checkout, primary, script = _fake_primary_with_script(tmp_path)
    calls = []

    def fake_run(args, **kwargs):
        calls.append(args)
        if args[0] == "git":
            return subprocess.CompletedProcess(args, 0, stdout=f"worktree {primary}\n")
        return subprocess.CompletedProcess(args, 0, stdout="")

    monkeypatch.setattr("subprocess.run", fake_run)

    with caplog.at_level("INFO", logger="triage_bot"):
        _cleanup(checkout)

    assert calls[-1] == ["bash", str(script), "--destroy"]
    messages = [r.message for r in caplog.records]
    assert "[cleanup] destroying cluster" in messages
    assert "[cleanup] done" in messages


def test_cleanup_warns_and_returns_when_script_is_missing(
    tmp_path, monkeypatch, caplog
):
    checkout = tmp_path / "checkout"
    checkout.mkdir()
    monkeypatch.setattr(
        "subprocess.run",
        lambda args, **kwargs: subprocess.CompletedProcess(args, 1, stdout=""),
    )
    monkeypatch.setattr("triage_bot.pipeline.repo_root", lambda: tmp_path / "nowhere")

    with caplog.at_level("INFO", logger="triage_bot"):
        _cleanup(checkout)

    assert any("cluster-up.sh not found" in r.message for r in caplog.records)


def test_cleanup_never_raises_even_if_destroy_fails(tmp_path, monkeypatch, caplog):
    # Best-effort by contract: a cleanup hiccup must never be mistaken for
    # the pipeline's own verdict.
    checkout, primary, _ = _fake_primary_with_script(tmp_path)

    def fake_run(args, **kwargs):
        if args[0] == "git":
            return subprocess.CompletedProcess(args, 0, stdout=f"worktree {primary}\n")
        raise subprocess.TimeoutExpired(cmd=args, timeout=120)

    monkeypatch.setattr("subprocess.run", fake_run)

    with caplog.at_level("INFO", logger="triage_bot"):
        _cleanup(checkout)  # must not raise

    assert any("[cleanup] failed" in r.message for r in caplog.records)
