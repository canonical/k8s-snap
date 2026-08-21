# Copyright 2026 Canonical, Ltd.
# See LICENSE file for licensing details.

"""Sanitize untrusted text before rendering Markdown."""

from __future__ import annotations

import re
from typing import Any
from urllib.parse import urlparse

MARKDOWN_TEXT_LIMIT = 240
COMPONENT_TEXT_LIMIT = 120
LABEL_TEXT_LIMIT = 80

_CONTROL_RE = re.compile(r"[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]")
_HTML_COMMENT_RE = re.compile(r"<!--.*?-->", re.DOTALL)
_HTML_TAG_RE = re.compile(r"</?[^>]+>")
_WHITESPACE_RE = re.compile(r"\s+")
_SHA_RE = re.compile(r"^[0-9a-f]{7,40}$", re.IGNORECASE)
_MD_SPECIAL_RE = re.compile(r"([\\`*_{}\[\]()#!|<>])")
_CATEGORY_VALUES = {
    "Major Feature",
    "Security",
    "Deprecation",
    "Bug Fix",
    "Performance",
    "Component Bump",
    "Documentation",
}


def markdown_text(value: Any, *, limit: int = MARKDOWN_TEXT_LIMIT) -> str:
    """Return *value* as escaped, single-line Markdown plain text."""
    text = _single_line(value, limit=limit)
    return _MD_SPECIAL_RE.sub(r"\\\1", text)


def component_text(value: Any) -> str:
    """Return a component/version entry as escaped plain text."""
    return markdown_text(value, limit=COMPONENT_TEXT_LIMIT)


def label_text(value: Any) -> str:
    """Return a short label as escaped plain text."""
    return markdown_text(value, limit=LABEL_TEXT_LIMIT)


def category(value: Any) -> str:
    """Return a known patch-notice category or a manual-review fallback."""
    text = "" if value is None else str(value).strip()
    return text if text in _CATEGORY_VALUES else "Review Required"


def sha(value: Any) -> str:
    """Return a validated commit SHA for HTML comments and inline code."""
    text = _single_line(value, limit=40).lower()
    if not _SHA_RE.fullmatch(text):
        raise ValueError(f"Invalid commit SHA: {value!r}")
    return text


def short_sha(value: Any) -> str:
    """Return the first eight characters of a validated commit SHA."""
    return sha(value)[:8]


def pr_number(value: Any) -> int | None:
    """Return a validated positive PR number, or None when absent."""
    if value in (None, ""):
        return None
    try:
        number = int(value)
    except (TypeError, ValueError):
        return None
    return number if number > 0 else None


def github_url(value: Any) -> str:
    """Return a safe GitHub URL for Markdown link targets, or an empty string."""
    text = _single_line(value, limit=300)
    if not text:
        return ""
    parsed = urlparse(text)
    if parsed.scheme != "https" or parsed.netloc != "github.com":
        return ""
    if ")" in text or "(" in text:
        return ""
    return text.replace(" ", "%20")


def _single_line(value: Any, *, limit: int) -> str:
    if value is None:
        return ""
    text = str(value)
    text = _CONTROL_RE.sub(" ", text)
    text = _HTML_COMMENT_RE.sub(" ", text)
    text = _HTML_TAG_RE.sub(" ", text)
    text = _WHITESPACE_RE.sub(" ", text).strip()
    if len(text) > limit:
        return text[: limit - 1].rstrip() + "..."
    return text