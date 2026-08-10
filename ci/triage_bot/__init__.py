#
# Copyright 2026 Canonical, Ltd.
#
"""Autonomous issue triage bot for canonical/k8s-snap.

A label-driven finite state machine: an issue's single ``triage/`` label is its
state, GitHub events drive transitions, and a pure :mod:`~triage_bot.router`
maps ``(event, current label)`` to one of the action handlers. The handlers own
the GitHub side effects and invoke project-owned *skills* (markdown under
``.agents/skills/triage``) to reproduce, diagnose, verify, and fix issues,
threading a ``report.md`` scratchpad between stages.

The deterministic labeler (:mod:`~triage_bot.triage_core`) and the skill
pipeline (:mod:`~triage_bot.pipeline`) are separated so the whole FSM runs
offline in tests with a fake GitHub client and a canned pipeline.
"""
