# Copyright 2026 Canonical, Ltd.
# See LICENSE file for licensing details.

from __future__ import annotations

import pytest

from patch_notices import sanitize


def test_markdown_text_escapes_markup_and_strips_html() -> None:
    text = sanitize.markdown_text(
        "[click](javascript:alert(1)) <script>alert(1)</script> | table\n# heading"
    )

    assert text == r"\[click\]\(javascript:alert\(1\)\) alert\(1\) \| table \# heading"


def test_markdown_text_strips_html_comments() -> None:
    assert sanitize.markdown_text("safe <!-- hidden --> text") == "safe text"


def test_sha_rejects_non_hex_content() -> None:
    with pytest.raises(ValueError):
        sanitize.sha("abcdef1 --><script>")


def test_github_url_rejects_non_github_targets() -> None:
    assert sanitize.github_url("javascript:alert(1)") == ""
    assert sanitize.github_url("https://example.com/canonical/k8s-snap/pull/1") == ""
    assert sanitize.github_url("https://github.com/canonical/k8s-snap/pull/1")


def test_category_uses_manual_review_fallback() -> None:
    assert sanitize.category("Bug Fix") == "Bug Fix"
    assert sanitize.category("<script>Bug Fix</script>") == "Review Required"