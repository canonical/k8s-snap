#
# Copyright 2026 Canonical, Ltd.
#
"""Guards on skill loading.

These pin the path contract: the CLI is invoked from ``ci/`` (locally and via
the workflow's ``working-directory``) while the skills live at the checkout
root, so a relative skill dir must never resolve against the process cwd.
"""

from __future__ import annotations

import subprocess

import pytest

from triage_bot import skills
from triage_bot.context import ActionContext


def _seed(root):
    skill = root / ".agents" / "skills" / "triage"
    skill.mkdir(parents=True)
    (skill / "SKILL.md").write_text("BASE", encoding="utf-8")
    (skill / "reproduce.md").write_text("STEP", encoding="utf-8")
    return skill


def test_relative_skill_dir_resolves_from_repo_root_not_cwd(tmp_path, monkeypatch):
    checkout = tmp_path / "checkout"
    _seed(checkout)
    monkeypatch.setattr(skills, "repo_root", lambda: checkout)
    # Process cwd is deliberately not the checkout root, mirroring ``ci/``.
    monkeypatch.chdir(tmp_path)

    text = skills.load_skill(".agents/skills/triage", "reproduce")

    assert "BASE" in text
    assert "STEP" in text


def test_absolute_skill_dir_is_used_as_is(tmp_path, monkeypatch):
    skill = _seed(tmp_path / "elsewhere")
    monkeypatch.setattr(skills, "repo_root", lambda: tmp_path / "unused")

    assert "BASE" in skills.load_skill(skill)


def test_default_skill_dir_loads_from_any_cwd(tmp_path, monkeypatch):
    # The shipped default must resolve for real: this is the regression that
    # made every trusted run fail with SkillError when invoked from ``ci/``.
    monkeypatch.chdir(tmp_path)

    assert skills.load_skill(ActionContext().triage_skill_dir, "reproduce")


def test_missing_skill_reports_the_resolved_path(tmp_path, monkeypatch):
    monkeypatch.setattr(skills, "repo_root", lambda: tmp_path)

    with pytest.raises(skills.SkillError, match=str(tmp_path)):
        skills.load_skill("nope")


def test_missing_step_file_raises_instead_of_dropping_instructions(tmp_path):
    # A missing SKILL.md + step pairing must be loud: silently running the
    # agent on the generic SKILL.md alone, without the step's instructions,
    # is a confusing, hard-to-diagnose degraded run.
    skill = tmp_path / "triage"
    skill.mkdir()
    (skill / "SKILL.md").write_text("BASE", encoding="utf-8")

    with pytest.raises(skills.SkillError, match="reproduce.md"):
        skills.load_skill(skill, "reproduce")


# --- shared skill composition ("> Uses:") ---


def _skill(root, name, skill_md="BASE", **steps):
    """A ``<root>/<name>`` skill directory: SKILL.md plus any given steps.

    ``steps`` maps step name to file text, e.g. ``reproduce="STEP"``.
    """
    skill = root / name
    skill.mkdir(parents=True)
    (skill / "SKILL.md").write_text(skill_md, encoding="utf-8")
    for step, text in steps.items():
        (skill / f"{step}.md").write_text(text, encoding="utf-8")
    return skill


def test_step_uses_directive_inlines_the_shared_skill(tmp_path):
    _skill(tmp_path, "local-cluster", "CLUSTER TEXT")
    skill = _skill(tmp_path, "triage", reproduce="> Uses: `local-cluster`\n\nSTEP")

    text = skills.load_skill(skill, "reproduce")

    assert "CLUSTER TEXT" in text


def test_base_skill_uses_directive_inlines_the_shared_skill(tmp_path):
    _skill(tmp_path, "local-cluster", "CLUSTER TEXT")
    skill = _skill(
        tmp_path, "triage", "> Uses: `local-cluster`\n\nBASE", reproduce="STEP"
    )

    text = skills.load_skill(skill, "reproduce")

    assert "CLUSTER TEXT" in text


def test_skill_named_by_both_base_and_step_is_inlined_once(tmp_path):
    _skill(tmp_path, "local-cluster", "CLUSTER TEXT")
    skill = _skill(
        tmp_path,
        "triage",
        "> Uses: `local-cluster`\n\nBASE",
        reproduce="> Uses: `local-cluster`\n\nSTEP",
    )

    text = skills.load_skill(skill, "reproduce")

    assert text.count("CLUSTER TEXT") == 1


def test_assembly_order_is_base_then_shared_then_step(tmp_path):
    _skill(tmp_path, "local-cluster", "CLUSTER TEXT")
    skill = _skill(tmp_path, "triage", reproduce="> Uses: `local-cluster`\n\nSTEP")

    text = skills.load_skill(skill, "reproduce")

    assert text.index("BASE") < text.index("CLUSTER TEXT") < text.index("STEP")


def test_missing_shared_skill_raises_naming_it(tmp_path):
    skill = _skill(tmp_path, "triage", reproduce="> Uses: `nope`\n\nSTEP")

    with pytest.raises(skills.SkillError, match="nope"):
        skills.load_skill(skill, "reproduce")


def test_invalid_shared_skill_name_raises(tmp_path):
    # Anything outside [a-z0-9][a-z0-9-]* must fail loudly rather than
    # resolve to nothing: a relative escape out of the skills tree, a name
    # whose case disagrees with the directory it has to match, and an empty
    # name (backticks with nothing between them).
    for index, bad_name in enumerate(("../evil", "Foo", "")):
        skill = _skill(
            tmp_path / str(index), "triage", reproduce=f"> Uses: `{bad_name}`\n\nSTEP"
        )

        with pytest.raises(skills.SkillError):
            skills.load_skill(skill, "reproduce")


def test_no_uses_directive_matches_todays_output_byte_for_byte(tmp_path):
    skill = _skill(tmp_path, "triage", reproduce="STEP")

    assert skills.load_skill(skill, "reproduce") == "BASE\n\n---\n\nSTEP"


def test_one_directive_naming_two_skills_inlines_both_in_order(tmp_path):
    # Two names in one directive.
    _skill(tmp_path, "local-cluster", "CLUSTER TEXT")
    _skill(tmp_path, "inspection-report", "REPORT TEXT")
    skill = _skill(
        tmp_path,
        "triage",
        reproduce="> Uses: `local-cluster`, `inspection-report`\n\nSTEP",
    )

    text = skills.load_skill(skill, "reproduce")

    assert text.index("CLUSTER TEXT") < text.index("REPORT TEXT") < text.index("STEP")


def test_uses_directive_inside_a_shared_skill_is_refused(tmp_path):
    # Nested Uses directives fail loudly.
    _skill(tmp_path, "inspection-report", "REPORT TEXT")
    _skill(tmp_path, "local-cluster", "> Uses: `inspection-report`\n\nCLUSTER TEXT")
    skill = _skill(tmp_path, "triage", reproduce="> Uses: `local-cluster`\n\nSTEP")

    with pytest.raises(skills.SkillError, match="nested composition"):
        skills.load_skill(skill, "reproduce")


def test_agent_shell_can_commit_with_the_supplied_identity(tmp_path):
    # The scratch HOME hides the runner's ~/.gitconfig, so without an injected
    # identity ``git commit`` fails and the fix step degrades to "no PR".
    repo = tmp_path / "repo"
    repo.mkdir()
    env = skills._safe_env(tmp_path / "home")

    def git(*args):
        return subprocess.run(
            ("git",) + args, cwd=repo, env=env, capture_output=True, text=True
        )

    git("init", "-q")
    (repo / "f").write_text("x", encoding="utf-8")
    git("add", "f")
    committed = git("commit", "-qm", "fix: something")

    assert committed.returncode == 0, committed.stderr
    assert "K8s builder bot" in git("log", "-1", "--format=%an <%ae>").stdout


def test_agent_shell_carries_no_secrets(tmp_path):
    env = skills._safe_env(tmp_path / "home")

    assert not [k for k in env if k.endswith(("_TOKEN", "_KEY"))]


def test_agent_shell_can_locate_the_session_bus_for_snap_tracking(
    tmp_path, monkeypatch
):
    # Every `snap run` (`snapcraft`, `lxc`) fails immediately with "cannot
    # run snap applications on this system" unless it can reach the session
    # D-Bus bus PAM/logind set up for this uid -- a bare subprocess child
    # does not inherit that pointer the way an interactive login shell does.
    # See https://forum.snapcraft.io/t/46210.
    monkeypatch.delenv("XDG_RUNTIME_DIR", raising=False)
    monkeypatch.delenv("DBUS_SESSION_BUS_ADDRESS", raising=False)
    monkeypatch.setattr(skills.os, "getuid", lambda: 424242)

    env = skills._safe_env(tmp_path / "home")

    assert env["XDG_RUNTIME_DIR"] == "/run/user/424242"
    assert env["DBUS_SESSION_BUS_ADDRESS"] == "unix:path=/run/user/424242/bus"


def test_agent_shell_prefers_the_orchestrators_own_session_bus(tmp_path, monkeypatch):
    # A runner where these are already set to something non-default must not
    # be overridden by the generic /run/user/<uid> guess.
    monkeypatch.setenv("XDG_RUNTIME_DIR", "/custom/runtime")
    monkeypatch.setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/custom/runtime/bus")

    env = skills._safe_env(tmp_path / "home")

    assert env["XDG_RUNTIME_DIR"] == "/custom/runtime"
    assert env["DBUS_SESSION_BUS_ADDRESS"] == "unix:path=/custom/runtime/bus"


def test_agent_shell_builds_tox_envs_off_the_checkout(tmp_path):
    # virtualenv symlinks the interpreter, which a multipass/sshfs checkout
    # mount rejects, so `.tox` must never be created inside the tree.
    env = skills._safe_env(tmp_path / "home")

    assert not env["TOX_WORK_DIR"].startswith(str(skills.repo_root()))


def test_agent_shell_defaults_to_the_scripts_own_prefix(tmp_path):
    env = skills._safe_env(tmp_path / "home")

    assert env["CLUSTER_PREFIX"] == skills.DEFAULT_CLUSTER_PREFIX


def test_agent_shell_uses_the_given_cluster_prefix(tmp_path):
    env = skills._safe_env(tmp_path / "home", "k8s-triage-42")

    assert env["CLUSTER_PREFIX"] == "k8s-triage-42"


def test_agent_shell_never_inherits_stdin(tmp_path):
    # A command that reads stdin must hit EOF immediately: inheriting a
    # terminal would let ``sudo``/``git`` prompts block for the full ceiling.
    shell = skills._make_shell_tool(tmp_path, tmp_path)

    out = shell.invoke({"command": "read -r line && echo PROMPTED || echo EOF"})

    assert "EOF" in out
    assert "PROMPTED" not in out


def test_agent_shell_timeout_returns_the_standard_exit_format(tmp_path, monkeypatch):
    # Every other return starts with "exit=<code>"; a timeout must too, or
    # callers (the agent included) need a separate path just to notice one.
    monkeypatch.setattr(skills, "_SHELL_TIMEOUT", 0.1)
    shell = skills._make_shell_tool(tmp_path, tmp_path)

    out = shell.invoke({"command": "sleep 5"})

    assert out.startswith("exit=124\n")
    assert "timed out" in out


def test_agent_shell_marks_truncated_output(tmp_path):
    # A silent cut would hide the real failure from both the agent and the
    # run log, which only ever see what this tool returns.
    shell = skills._make_shell_tool(tmp_path, tmp_path)

    out = shell.invoke({"command": "printf 'x%.0s' {1..9000}"})

    assert "chars omitted" in out
    assert len(out) < 9000


def test_agent_shell_leaves_short_output_untouched(tmp_path):
    shell = skills._make_shell_tool(tmp_path, tmp_path)

    out = shell.invoke({"command": "echo hello"})

    assert out == "exit=0\nhello\n"


# --- worktree isolation ---


def _tiny_repo(tmp_path):
    """A throwaway repository standing in for the checkout."""
    root = tmp_path / "primary"
    root.mkdir()

    def git(*args):
        subprocess.run(("git", "-C", str(root), *args), check=True, capture_output=True)

    git("init", "-q", "-b", "main")
    (root / "tracked.txt").write_text("committed\n", encoding="utf-8")
    git("add", "tracked.txt")
    git("-c", "user.name=t", "-c", "user.email=t@e", "commit", "-qm", "init")
    return root, git


def test_worktree_cannot_disturb_the_primary_tree(tmp_path, monkeypatch):
    # The failure this prevents: an agent finding the primary tree dirty and
    # reverting work it did not own to get a clean `git status`.
    root, _ = _tiny_repo(tmp_path)
    (root / "tracked.txt").write_text("uncommitted work\n", encoding="utf-8")
    monkeypatch.setattr(skills, "repo_root", lambda: root)

    tree = skills.ensure_worktree(tmp_path / "scratch" / "checkout", "triage/fix-1")

    # The agent sees the committed state; the maintainer's edit is untouched
    # and invisible to it.
    assert (tree / "tracked.txt").read_text(encoding="utf-8") == "committed\n"
    assert (root / "tracked.txt").read_text(encoding="utf-8") == "uncommitted work\n"


def test_worktree_is_reused_and_stale_ones_reclaimed(tmp_path, monkeypatch):
    root, _ = _tiny_repo(tmp_path)
    monkeypatch.setattr(skills, "repo_root", lambda: root)
    first = skills.ensure_worktree(tmp_path / "a" / "checkout", "triage/fix-1")

    assert skills.ensure_worktree(first, "triage/fix-1") == first

    # A crashed run left the branch checked out elsewhere; retrying at a new
    # path must reclaim it rather than fail forever.
    second = skills.ensure_worktree(tmp_path / "b" / "checkout", "triage/fix-1")
    assert (second / "tracked.txt").exists()


def test_worktree_refuses_to_reuse_a_directory_on_the_wrong_branch(
    tmp_path, monkeypatch
):
    # A directory with a .git could be stale, manually tampered with, or a
    # leftover from a future bug that reuses a path across branches. Working
    # in it anyway would commit the agent's changes onto the wrong branch,
    # and a later push of the *expected* branch name would push stale
    # content instead of the run's actual work.
    root, git = _tiny_repo(tmp_path)
    monkeypatch.setattr(skills, "repo_root", lambda: root)
    path = tmp_path / "a" / "checkout"
    skills.ensure_worktree(path, "triage/fix-1")
    subprocess.run(
        ("git", "-C", str(path), "switch", "-c", "some-other-branch"),
        check=True,
        capture_output=True,
    )

    with pytest.raises(skills.SkillError, match="some-other-branch"):
        skills.ensure_worktree(path, "triage/fix-1")


def test_worktree_keeps_commits_an_earlier_run_made(tmp_path, monkeypatch):
    # The reproducer test is committed on the branch before the fix stage runs.
    # Re-creating the worktree (a retry, a crash) must not reset that away.
    root, _ = _tiny_repo(tmp_path)
    monkeypatch.setattr(skills, "repo_root", lambda: root)
    first = skills.ensure_worktree(tmp_path / "a" / "checkout", "triage/fix-1")
    (first / "test_repro.py").write_text("assert False\n", encoding="utf-8")
    subprocess.run(
        ("git", "-C", str(first), "add", "test_repro.py"),
        check=True,
        capture_output=True,
    )
    subprocess.run(
        (
            "git",
            "-C",
            str(first),
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@e",
            "commit",
            "-qm",
            "test: reproduce",
        ),
        check=True,
        capture_output=True,
    )

    second = skills.ensure_worktree(tmp_path / "b" / "checkout", "triage/fix-1")

    assert (second / "test_repro.py").exists()


def test_worktree_refuses_to_steal_a_branch_from_the_primary_tree(
    tmp_path, monkeypatch
):
    root, git = _tiny_repo(tmp_path)
    git("switch", "-q", "-c", "triage/fix-1")
    monkeypatch.setattr(skills, "repo_root", lambda: root)

    with pytest.raises(skills.SkillError, match="primary tree"):
        skills.ensure_worktree(tmp_path / "scratch" / "checkout", "triage/fix-1")


def test_commit_sha_reads_the_worktrees_own_head(tmp_path, monkeypatch):
    root, git = _tiny_repo(tmp_path)
    monkeypatch.setattr(skills, "repo_root", lambda: root)
    tree = skills.ensure_worktree(tmp_path / "scratch" / "checkout", "triage/fix-1")

    assert skills.commit_sha(tree) == skills.commit_sha(root)


def test_commit_sha_advances_after_a_new_commit(tmp_path, monkeypatch):
    root, git = _tiny_repo(tmp_path)
    monkeypatch.setattr(skills, "repo_root", lambda: root)
    tree = skills.ensure_worktree(tmp_path / "scratch" / "checkout", "triage/fix-1")
    before = skills.commit_sha(tree)

    (tree / "new_file.txt").write_text("x\n", encoding="utf-8")
    subprocess.run(("git", "-C", str(tree), "add", "new_file.txt"), check=True)
    subprocess.run(
        (
            "git",
            "-C",
            str(tree),
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@e",
            "commit",
            "-qm",
            "a real commit",
        ),
        check=True,
    )

    after = skills.commit_sha(tree)
    assert after != before
    assert len(after) == 40  # a real, full git SHA -- not a truncated/empty read
