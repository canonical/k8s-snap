#
# Copyright 2026 Canonical, Ltd.
#
"""Label-driven state machine configuration."""

from __future__ import annotations

from dataclasses import dataclass, fields


@dataclass(frozen=True)
class LabelConfig:
    needs_triage: str = "triage/needs-triage"
    needs_human: str = "triage/needs-human"
    needs_reproduction: str = "triage/needs-reproduction"
    needs_manual_review: str = "triage/needs-manual-review"
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
        """All triage-state labels."""
        return [
            getattr(self, f.name) for f in fields(self) if f.name not in self._NON_STATE
        ]

    def retriageable_labels(self) -> list[str]:
        """Labels eligible for retriage on fresh comments."""
        return [
            self.needs_triage,
            self.needs_reproduction,
            self.unable_to_reproduce,
            self.unable_to_fix,
            self.failed,
            self.fix_rejected,
        ]

    def terminal_labels(self) -> list[str]:
        """Labels where the bot takes no further action."""
        return [
            self.fix_verified,
            self.needs_human,
            self.skipped,
            self.needs_manual_review,
        ]


def current_triage_label(issue_labels: list[str], config: LabelConfig) -> str | None:
    """Return the first triage label present on the issue, or ``None``."""
    present = set(issue_labels)
    return next(
        (label for label in config.all_triage_labels() if label in present), None
    )
