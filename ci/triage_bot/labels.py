#
# Copyright 2026 Canonical, Ltd.
#
"""Label-driven state machine configuration.

The bot keeps no database. An issue's triage state *is* the single ``triage:``
label it carries, and transitions are atomic label swaps (remove the old, add
the new). This module owns the label vocabulary and the three sets the router
needs: every triage label, the re-triageable subset (a new comment may restart
triage), and the terminal subset (the bot stops acting).
"""

from __future__ import annotations

from dataclasses import dataclass, fields


@dataclass(frozen=True)
class LabelConfig:
    needs_triage: str = "triage/needs-triage"
    not_actionable: str = "triage/not-actionable"
    needs_reproduction: str = "triage/needs-reproduction"
    skipped: str = "triage/skipped"
    unable_to_reproduce: str = "triage/unable-to-reproduce"
    unable_to_fix: str = "triage/unable-to-fix"
    failed: str = "triage/failed"
    fix_pending: str = "triage/fix-pending"
    fix_rejected: str = "triage/fix-rejected"
    fix_verified: str = "triage/fix-verified"
    # Not triage states: applied alongside a terminal label, never swapped.
    pr_fix_verified: str = "fix-verified"
    docs_needed: str = "docs-change-needed"

    _NON_STATE = ("pr_fix_verified", "docs_needed")

    def all_triage_labels(self) -> list[str]:
        """Every triage-state label (excludes the labels above)."""
        return [
            getattr(self, f.name) for f in fields(self) if f.name not in self._NON_STATE
        ]

    def retriageable_labels(self) -> list[str]:
        """Labels where a fresh non-bot comment may warrant re-triage."""
        return [
            self.needs_triage,
            self.needs_reproduction,
            self.unable_to_reproduce,
            self.unable_to_fix,
            self.failed,
            self.fix_rejected,
        ]

    def terminal_labels(self) -> list[str]:
        """Labels where the bot takes no further action on new comments."""
        return [self.fix_verified, self.not_actionable, self.skipped]


def current_triage_label(issue_labels: list[str], config: LabelConfig) -> str | None:
    """Return the first triage label present on the issue, or ``None``.

    An issue should carry at most one triage label; if several are present
    (manual tampering) the first match in declaration order wins, which keeps
    routing deterministic.
    """
    present = set(issue_labels)
    return next(
        (label for label in config.all_triage_labels() if label in present), None
    )
