# Copyright 2026 Canonical, Ltd.
# See LICENSE file for licensing details.

from __future__ import annotations

from pathlib import Path

from patch_notices import workbook


def _triage_result() -> list[dict]:
    return [
        {
            "sha": "abcdef1234567890abcdef1234567890abcdef12",
            "title": "[bad](javascript:alert(1)) <b>title</b> | table",
            "pr_number": 12,
            "pr_url": "https://github.com/canonical/k8s-snap/pull/12",
            "triage": {
                "action": "include",
                "category": "Bug Fix",
                "summary": "Fix [thing](javascript:alert(1)) <script>x</script> | pipe",
                "components": None,
                "reason": None,
                "limited_context": True,
                "group_hint": None,
            },
        },
        {
            "sha": "1234567890abcdef1234567890abcdef12345678",
            "title": "Discard <img src=x onerror=alert(1)>",
            "pr_number": 13,
            "html_url": "https://github.com/canonical/k8s-snap/commit/1234567890abcdef1234567890abcdef12345678",
            "triage": {
                "action": "discard",
                "category": None,
                "summary": None,
                "components": None,
                "reason": "Only updates [CI](javascript:alert(1)) <!-- hidden -->",
                "limited_context": False,
                "group_hint": None,
            },
        },
    ]


def test_clean_entry_lines_escape_ai_summary() -> None:
    text = "\n".join(workbook._clean_entry_lines(_triage_result()))

    assert "<script>" not in text
    assert "[thing](javascript:alert(1))" not in text
    assert r"\[thing\]\(javascript:alert\(1\)\)" in text
    assert r"\| pipe" in text


def test_write_escapes_workbook_fields(tmp_path: Path) -> None:
    output = tmp_path / "review.md"
    workbook.write(_triage_result(), str(output), channel_key="snap:1.35-classic/stable")
    text = output.read_text()

    assert "<!-- sha:abcdef1234567890abcdef1234567890abcdef12 -->" in text
    assert "[bad](javascript:alert(1))" not in text
    assert "<img" not in text
    assert r"\[bad\]\(javascript:alert\(1\)\)" in text
    assert r"Only updates \[CI\]\(javascript:alert\(1\)\)" in text


def test_insert_patch_notice_escapes_release_note_entry(tmp_path: Path) -> None:
    notes = tmp_path / "release.md"
    notes.write_text("# 1.35\n\n## Patch notices\n\nJul 1, 2026\n\n- Existing\n")

    workbook.insert_patch_notice(str(notes), _triage_result(), "Aug 19, 2026", "snap")
    text = notes.read_text()

    assert "Aug 19, 2026" in text
    assert "<script>" not in text
    assert "[thing](javascript:alert(1))" not in text
    assert r"\[thing\]\(javascript:alert\(1\)\)" in text


def test_build_pr_body_escapes_summary_titles_and_reasons() -> None:
    summary = workbook.build_track_summary(
        _triage_result(), "snap:1.35-classic/stable", "Aug 19, 2026"
    )
    text = workbook.build_pr_body([summary])

    assert "<script>" not in text
    assert "<img" not in text
    assert "[bad](javascript:alert(1))" not in text
    assert "[CI](javascript:alert(1))" not in text
    assert r"\[bad\]\(javascript:alert\(1\)\)" in text
    assert r"\[CI\]\(javascript:alert\(1\)\)" in text
    assert r"\| table" in text


def test_build_pr_body_tolerates_malformed_summary_shas() -> None:
    text = workbook.build_pr_body(
        [
            {
                "track": "snap:1.35-classic/stable",
                "status": "all-discarded",
                "date": "Aug 20, 2026",
                "included": [],
                "discarded": [
                    {
                        "sha": "",
                        "pr_number": None,
                        "title": "bad summary",
                        "reason": "bad sha",
                    }
                ],
                "limited_context": [
                    {
                        "sha": "not-a-sha",
                        "pr_number": 1,
                        "title": "bad limited context",
                    }
                ],
            }
        ]
    )

    assert "`unknown` (PR #1) — bad limited context" in text
    assert "`unknown` — bad summary" in text
