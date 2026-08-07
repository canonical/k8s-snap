#
# Copyright 2026 Canonical, Ltd.
#
"""The ``report.md`` scratchpad shared across pipeline stages.

Each pipeline stage (reproduce -> verify -> reproducer -> diagnose -> fix) reads
the report written by the previous stage and appends its own findings. The file
is the only state carried between stages, so it is the single source of truth
the comment generator later renders. Kept deliberately simple: a working
directory per issue plus append-only section writes.
"""

from __future__ import annotations

from pathlib import Path


class Report:
    """Append-only ``report.md`` inside a per-issue triage directory."""

    def __init__(self, triage_dir: str | Path):
        self.dir = Path(triage_dir)
        self.path = self.dir / "report.md"

    def start(self, issue_number: int, title: str, body: str) -> "Report":
        self.dir.mkdir(parents=True, exist_ok=True)
        header = (
            f"# Triage report for #{issue_number}\n\n"
            f"## Issue\n\n**{title}**\n\n{body}\n"
        )
        self.path.write_text(header, encoding="utf-8")
        return self

    def append(self, section: str, text: str) -> None:
        with self.path.open("a", encoding="utf-8") as fh:
            fh.write(f"\n## {section}\n\n{text.rstrip()}\n")

    def read(self) -> str:
        return self.path.read_text(encoding="utf-8") if self.path.exists() else ""
