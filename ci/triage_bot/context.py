#
# Copyright 2026 Canonical, Ltd.
#
"""Shared context threaded through every handler."""

from __future__ import annotations

from dataclasses import dataclass, field

from .labels import LabelConfig
from .llm import DEFAULT_MODEL

# Verification/retriage model spec; defaults to DEFAULT_MODEL.
DEFAULT_VERIFICATION_MODEL = DEFAULT_MODEL

# Bot logins ignored to avoid self-trigger loops.
ALWAYS_IGNORED_BOTS = ("github-actions[bot]",)


@dataclass
class ActionContext:
    repo: str = "canonical/k8s-snap"
    dry_run: bool = False
    triage_skill_dir: str = ".agents/skills/triage"
    # Open a draft PR when the run produces a branch.
    auto_pr: bool = False
    triage_model: str = DEFAULT_MODEL
    verification_model: str = DEFAULT_VERIFICATION_MODEL
    labels: LabelConfig = field(default_factory=LabelConfig)
    bot_logins: tuple[str, ...] = ALWAYS_IGNORED_BOTS
    jsonl_path: str | None = None
    # Maximum triage failures before stopping automated retriage.
    max_triage_failures: int = 3
    # Root directory for per-issue scratch space (report.md, checkouts).
    workdir_root: str = ".triage"
    # Maintainer team for mentions (without leading '@').
    maintainer_team: str = "canonical/kubernetes"

    def with_bot_logins(self, extra: list[str]) -> "ActionContext":
        merged = list(ALWAYS_IGNORED_BOTS)
        for login in extra:
            login = login.strip()
            if login and login not in merged:
                merged.append(login)
        self.bot_logins = tuple(merged)
        return self
