#
# Copyright 2026 Canonical, Ltd.
#
"""Shared context threaded through every handler.

Mirrors the action-owned/project-owned split of the triage bot: this dataclass
holds the action-owned configuration (repo, models, label set, skill location)
while the project-owned behaviour lives in the skill markdown under
``triage_skill_dir``. Secrets are never stored here; the GitHub boundary reads
``GH_TOKEN`` from the environment and the LLM factory reads its provider key
from the environment.
"""

from __future__ import annotations

from dataclasses import dataclass, field

from .labels import LabelConfig
from .llm import DEFAULT_MODEL

# Verification/retriage is a cheaper decision than the full pipeline, so it
# defaults to the same provider but is overridable independently.
DEFAULT_VERIFICATION_MODEL = DEFAULT_MODEL

# github-actions[bot] is always ignored so the bot never triggers on its own
# comments; additional bot logins are merged in from configuration.
ALWAYS_IGNORED_BOTS = ("github-actions[bot]",)


@dataclass
class ActionContext:
    repo: str = "canonical/k8s-snap"
    dry_run: bool = False
    triage_skill_dir: str = ".agents/skills/triage"
    # Opens a draft PR when the run produces a branch, whether or not the fix
    # succeeded (a proven-red reproducer test is worth landing on its own).
    auto_pr: bool = False
    triage_model: str = DEFAULT_MODEL
    verification_model: str = DEFAULT_VERIFICATION_MODEL
    labels: LabelConfig = field(default_factory=LabelConfig)
    bot_logins: tuple[str, ...] = ALWAYS_IGNORED_BOTS
    jsonl_path: str | None = None
    # After this many consecutive failed triage attempts the bot stops
    # re-triaging an issue on new comments (withastro's MAX_TRIAGE_FAILURES).
    max_triage_failures: int = 3
    # Root directory for per-issue scratch space (report.md, checkouts).
    workdir_root: str = ".triage"
    # GitHub team to mention in the maintainer section of enhancement proposals.
    # Written without the leading ``@``; leave empty to skip the mention.
    maintainer_team: str = "canonical/kubernetes"

    def with_bot_logins(self, extra: list[str]) -> "ActionContext":
        merged = list(ALWAYS_IGNORED_BOTS)
        for login in extra:
            login = login.strip()
            if login and login not in merged:
                merged.append(login)
        self.bot_logins = tuple(merged)
        return self
