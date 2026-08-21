#
# Copyright 2026 Canonical, Ltd.
#
"""Invariants of the label configuration and current-label resolution."""

from __future__ import annotations

from triage_bot.labels import LabelConfig, current_triage_label

LABELS = LabelConfig()


def test_terminal_and_retriageable_are_disjoint():
    assert set(LABELS.terminal_labels()).isdisjoint(LABELS.retriageable_labels())


def test_all_triage_labels_excludes_pr_only_label():
    assert LABELS.pr_fix_verified not in LABELS.all_triage_labels()


def test_fix_pending_is_neither_terminal_nor_retriageable():
    # It is a distinct waiting state routed straight to verify-fix.
    assert LABELS.fix_pending not in LABELS.terminal_labels()
    assert LABELS.fix_pending not in LABELS.retriageable_labels()


def test_current_triage_label_picks_the_triage_label():
    labels = ["kind/bug", "area/dns", LABELS.needs_reproduction]
    assert current_triage_label(labels, LABELS) == LABELS.needs_reproduction


def test_current_triage_label_none_when_absent():
    assert current_triage_label(["kind/bug"], LABELS) is None


def test_current_triage_label_declaration_order_wins():
    # Two triage labels present (manual tampering): first in declaration order.
    labels = [LABELS.skipped, LABELS.needs_triage]
    assert current_triage_label(labels, LABELS) == LABELS.needs_triage
