#
# Copyright 2026 Canonical, Ltd.
#
"""Skill runner: execute a project-owned markdown skill as an agent.

This is the bridge between action-owned orchestration (the handlers) and
project-owned behaviour (the ``.agents/skills/triage`` markdown). A skill is a
directory with a ``SKILL.md`` plus per-step files (``reproduce.md`` etc.). The
runner builds a tool-using agent whose system prompt is the skill text, lets it
run shell commands in the triage working directory, and coerces the final
answer into a typed Pydantic model via structured output.

Kept import-light: LangGraph/LangChain are imported lazily inside ``run_skill``
so the base CLI and the pure FSM/router stay dependency-free.
"""

from __future__ import annotations

import json
import os
import subprocess
import uuid
from pathlib import Path
from typing import Optional, Type, TypeVar

from pydantic import BaseModel

from .llm import DEFAULT_MODEL

T = TypeVar("T", bound=BaseModel)

# Shell commands the agent runs can be slow, so the per-command ceiling is
# generous but bounded to avoid a hung CI job. It must exceed the slowest
# command the skill mandates -- a cold ``snapcraft --use-lxd``, documented in
# SKILL.md as "tens of minutes" -- otherwise the build is killed mid-flight and
# the primary reproduce path can never succeed.
_SHELL_TIMEOUT = 3600

# The agent and the run log both only see what this tool returns, so a
# silent cut would hide the real failure from both. Keep the tail (the most
# recent, usually most relevant output) but mark it explicitly when trimmed.
_SHELL_OUTPUT_CHARS = 8000

# LangGraph's default recursion limit (25) allows only ~12 tool calls, far
# fewer than a reproduce or fix step needs (expand tarball, build, bootstrap a
# cluster, drive `k8s`, run an e2e test). Exhausting it aborts the step, so the
# cap is raised to bound runaway loops rather than to bound useful work.
_RECURSION_LIMIT = 120


class SkillError(RuntimeError):
    """Raised when a skill cannot be loaded or produces no structured result."""


def load_skill(skill_dir: str | Path, step: Optional[str] = None) -> str:
    """Return the skill text: ``SKILL.md`` plus the step file when given."""
    # A relative skill_dir is anchored to the checkout root, not the process
    # cwd: the CLI runs from ``ci/`` (locally and via the workflow's
    # ``working-directory``), while the skills live at the repo root. An
    # absolute dir (e.g. a test's tmp_path) is used as-is.
    root = Path(skill_dir)
    if not root.is_absolute():
        root = repo_root() / root
    skill_md = root / "SKILL.md"
    if not skill_md.exists():
        raise SkillError(f"skill not found: {skill_md}")
    parts = [skill_md.read_text(encoding="utf-8")]
    if step:
        step_md = root / f"{step}.md"
        if not step_md.exists():
            raise SkillError(f"skill step not found: {step_md}")
        parts.append(step_md.read_text(encoding="utf-8"))
    return "\n\n---\n\n".join(parts)


# The agent's shell is driven by reporter-controlled text. The measures below
# strip the obvious in-process/on-disk secret paths as DEFENSE IN DEPTH -- they
# raise the bar but are NOT the containment boundary:
#   * env: pass only a minimal, secret-free allowlist -- no *_TOKEN/*_KEY
#     (GH_TOKEN, GOOGLE_API_KEY, ...) in the shell's own environ;
#   * HOME: point at a throwaway per-issue dir so `~/.config/gh/hosts.yml`,
#     ~/.netrc, ~/.kube/config resolve to empties, not the runner's real ones;
#   * shell: use ``bash -c`` (NOT ``-lc``) so no /etc/profile or ~/.bash_profile
#     is sourced -- a login shell could re-export the very secrets stripped here.
# A same-uid child can still read the orchestrator's env via `/proc/<ppid>/environ`
# (GOOGLE_API_KEY must stay in the parent for the LLM client), and the pipeline's
# passwordless sudo reaches root-owned creds regardless of HOME. Real containment
# therefore REQUIRES running the pipeline job on an isolated, ephemeral, throwaway
# self-hosted runner (or under bwrap/nsjail with a private /proc and no host-HOME
# bind, as a dedicated low-privilege uid without passwordless sudo). See the
# rollout notes in the plan; the workflow's pipeline job documents this too.
_ENV_PASSTHROUGH = ("PATH", "LANG", "LC_ALL", "TERM", "TZ")


def repo_root() -> Path:
    """The k8s-snap checkout root.

    ``ci/triage_bot/skills.py`` -> two parents up is the repository root, which
    is where the agent edits source and builds the snap. Derived from the module
    path (not the process cwd) so it is stable however the CLI is invoked.
    """
    return Path(__file__).resolve().parents[2]


# The scratch HOME hides the runner's ~/.gitconfig, so the agent has no commit
# identity and ``git commit`` in the fix step would fail with "Committer
# identity unknown" -- silently degrading every fix to "no branch, no PR".
# Supply the project's bot identity via env (not ``git config``, which would
# mutate the checkout). Matches ``build-scripts/build-component.sh``.
_GIT_IDENTITY = {
    "GIT_AUTHOR_NAME": "K8s builder bot",
    "GIT_AUTHOR_EMAIL": "k8s-bot@canonical.com",
    "GIT_COMMITTER_NAME": "K8s builder bot",
    "GIT_COMMITTER_EMAIL": "k8s-bot@canonical.com",
}

# virtualenv builds an env by symlinking the interpreter, which some checkout
# mounts refuse outright (a multipass/sshfs share answers "Operation not
# supported"), so `tox -e integration` can never create `.tox/` inside the
# tree. Point it at local disk: the agent gets a working test runner without
# having to discover this, and the env is reused across pipeline steps.
_TOX_WORK_DIR = os.environ.get("TRIAGE_TOX_WORK_DIR", "/tmp/triage-tox")


# Matches hack/cluster-up.sh's own ``PREFIX="${CLUSTER_PREFIX:-k8s-triage}"``
# default. Kept as one named constant (rather than duplicating the literal)
# since pipeline.py's cleanup must destroy the same prefix the agent used.
DEFAULT_CLUSTER_PREFIX = "k8s-triage"


def _safe_env(home: Path, cluster_prefix: str = DEFAULT_CLUSTER_PREFIX) -> dict:
    """A minimal, secret-free environment for the agent shell.

    ``CLUSTER_PREFIX`` scopes ``hack/cluster-up.sh`` to this issue: concurrent
    pipeline runs (different issues, same self-hosted runner pool) would
    otherwise share the script's default prefix and collide on container
    names, or have one run's cleanup destroy another's cluster.
    """
    env = {k: os.environ[k] for k in _ENV_PASSTHROUGH if k in os.environ}
    env.setdefault(
        "PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    )
    env["HOME"] = str(home)
    env.update(_GIT_IDENTITY)
    env["TOX_WORK_DIR"] = _TOX_WORK_DIR
    env["CLUSTER_PREFIX"] = cluster_prefix
    return env


def ensure_worktree(path: Path, branch: str) -> Path:
    """Create (or reuse) an isolated checkout for the agent.

    The bot must never work in the primary checkout. A live tree carries
    unrelated work in progress, and an agent that finds it dirty is tempted to
    "clean" it -- one run reverted a maintainer's uncommitted files to get a
    tidy `git status`. A worktree shares the object store but has its own
    branch, index and files, so the agent physically cannot touch the primary
    tree, and its branch is ready to commit on without a checkout step.
    """
    if (path / ".git").exists():
        return path
    root = repo_root()
    _git(root, "worktree", "prune")
    _release_branch(root, branch)
    path.parent.mkdir(parents=True, exist_ok=True)
    # Reuse the branch when it already exists: an earlier run may have
    # committed the reproducer test on it, and `-B` would reset that away.
    if (
        _git(
            root, "rev-parse", "--verify", "--quiet", f"refs/heads/{branch}"
        ).returncode
        == 0
    ):
        add = _git(root, "worktree", "add", str(path), branch)
    else:
        add = _git(root, "worktree", "add", "-b", branch, str(path), "HEAD")
    if add.returncode != 0:
        raise SkillError(f"could not create worktree at {path}: {add.stderr.strip()}")
    return path


def _git(cwd: Path, *args: str) -> subprocess.CompletedProcess:
    return subprocess.run(
        ["git", "-C", str(cwd), *args], capture_output=True, text=True
    )


def _release_branch(root: Path, branch: str) -> None:
    """Free ``branch`` if a leftover worktree still holds it.

    Git allows a branch in only one worktree, so a crashed earlier run would
    otherwise block every retry. ``triage/fix-*`` is the bot's own namespace,
    so reclaiming a *linked* worktree that holds it is safe. The primary
    checkout is never touched: if a human has the branch out there, say so
    instead of yanking the tree from under them.
    """
    trees: list[list] = []
    for line in _git(root, "worktree", "list", "--porcelain").stdout.splitlines():
        if line.startswith("worktree "):
            trees.append([line.removeprefix("worktree "), None])
        elif line.startswith("branch ") and trees:
            trees[-1][1] = line.removeprefix("branch ")
    for index, (tree_path, tree_branch) in enumerate(trees):
        if tree_branch != f"refs/heads/{branch}":
            continue
        # The first entry git reports is always the primary checkout.
        if index == 0:
            raise SkillError(
                f"branch {branch} is checked out in the primary tree "
                f"({tree_path}). Switch it away before triaging this issue."
            )
        _git(root, "worktree", "remove", "--force", tree_path)
        return


def _make_shell_tool(
    workdir: Path, cwd: Path, cluster_prefix: str = DEFAULT_CLUSTER_PREFIX
):
    from langchain_core.tools import tool

    # The agent works in its own worktree; the per-issue dir is scratch --
    # reuse it as an isolated HOME so credential stores under the runner
    # user's real HOME stay out of reach.
    home = workdir / "home"
    home.mkdir(parents=True, exist_ok=True)
    safe_env = _safe_env(home, cluster_prefix)

    @tool
    def shell(command: str) -> str:
        """Run a bash command in the k8s-snap checkout and return its combined
        stdout/stderr and exit code."""
        try:
            proc = subprocess.run(
                ["bash", "-c", command],
                cwd=str(cwd),
                capture_output=True,
                text=True,
                timeout=_SHELL_TIMEOUT,
                env=safe_env,
                # Never inherit a terminal: a command that prompts (``sudo``
                # without a cached credential, ``git`` asking for a password)
                # would otherwise block for the full ceiling instead of failing.
                stdin=subprocess.DEVNULL,
            )
        except subprocess.TimeoutExpired:
            return f"[timed out after {_SHELL_TIMEOUT}s]"
        out = (proc.stdout or "") + (proc.stderr or "")
        if len(out) > _SHELL_OUTPUT_CHARS:
            omitted = len(out) - _SHELL_OUTPUT_CHARS
            out = f"(+{omitted} chars omitted)...\n{out[-_SHELL_OUTPUT_CHARS:]}"
        return f"exit={proc.returncode}\n{out}"

    return shell


def run_skill(
    *,
    skill_dir: str | Path,
    step: str,
    instructions: str,
    result_model: Type[T],
    workdir: str | Path,
    cwd: str | Path,
    model_spec: str = DEFAULT_MODEL,
    extra_context: str = "",
    run_id: Optional[str] = None,
    jsonl_path: Optional[str] = None,
    cluster_prefix: str = DEFAULT_CLUSTER_PREFIX,
) -> T:
    """Run one skill step and return its structured result.

    ``instructions`` scopes the step (e.g. "run only reproduce"); ``workdir`` is
    the per-issue scratch directory (report, agent HOME); ``cwd`` is the
    worktree the shell runs in; ``result_model`` is the Pydantic schema the
    agent must satisfy. When ``run_id`` is given the stage is logged (redacted)
    via :class:`~triage_bot.runlog.RunLogger`, to ``jsonl_path`` as well when
    set. ``cluster_prefix`` scopes the agent's ``hack/cluster-up.sh`` calls to
    this issue; the caller must destroy that same prefix in cleanup.
    """
    from langgraph.prebuilt import create_react_agent

    from .llm import make_llm
    from .runlog import CredentialRedactor, RunLogger, jsonl_sink

    system = load_skill(skill_dir, step)
    agent = create_react_agent(
        make_llm(model_spec),
        tools=[_make_shell_tool(Path(workdir), Path(cwd), cluster_prefix)],
        prompt=system,
    )
    user = instructions if not extra_context else f"{instructions}\n\n{extra_context}"

    run_id = run_id or uuid.uuid4().hex[:12]
    redactor = CredentialRedactor(
        {k: v for k, v in os.environ.items() if k.endswith(("_TOKEN", "_KEY"))}
    )
    logger = RunLogger(
        run_id, redactor, sink=jsonl_sink(jsonl_path) if jsonl_path else None
    )
    result = agent.invoke(
        {"messages": [{"role": "user", "content": user}]},
        config={
            "callbacks": [logger],
            "run_name": f"skill-{step}",
            "recursion_limit": _RECURSION_LIMIT,
        },
    )
    return _structure(
        result["messages"], step=step, result_model=result_model, model_spec=model_spec
    )


def _structure(messages, *, step: str, result_model: Type[T], model_spec: str) -> T:
    """Derive the step's structured result from the agent's final answer.

    ``create_react_agent(response_format=...)`` cannot do this: its
    ``generate_structured_response`` node replays the transcript, which always
    ends on a model turn, and Gemini rejects those outright ("Requests ending
    with a model turn are not supported"). Re-asking with the answer as a user
    turn is provider-neutral and mirrors the classifier in ``triage_core``.
    """
    from .llm import make_llm

    answer = next(
        (
            str(m.content)
            for m in reversed(messages)
            if getattr(m, "type", "") == "ai" and m.content
        ),
        "",
    )
    if not answer:
        raise SkillError(f"skill step {step!r} produced no answer to structure")
    structured = (
        make_llm(model_spec)
        .with_structured_output(result_model)
        .invoke(
            f"A triage agent ran the {step!r} step and reported:\n\n{answer}\n\n"
            "Return that report as the structured result."
        )
    )
    if not isinstance(structured, result_model):
        raise SkillError(f"skill step {step!r} returned no structured result")
    return structured


def render_report_json(model: BaseModel) -> str:
    """Compact JSON of a result model, for logging into ``report.md``."""
    return json.dumps(model.model_dump(), indent=2, default=str)
