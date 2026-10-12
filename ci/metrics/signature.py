#
# Copyright 2026 Canonical, Ltd.
#
"""Excerpt extraction, normalisation, and signature hashing."""

from __future__ import annotations

import hashlib
import re
from dataclasses import dataclass
from typing import Optional, Tuple

from metrics.scrub import scrub

NORMALISER_VERSION = 1
EXTRACTOR_VERSION = 1
MAX_EXCERPT_BYTES = 4096
SIGNATURE_LENGTH = 16

_GH_TIMESTAMP = re.compile(r"^\d{4}-\d{2}-\d{2}T[\d:.]+Z\s?")
_FAILURES_HEADER = re.compile(r"^=+\s*FAILURES\s*=+$")
_ERRORS_HEADER = re.compile(r"^=+\s*ERRORS\s*=+$")
_SUMMARY_HEADER = re.compile(r"^=+\s*short test summary info\s*=+$", re.IGNORECASE)
_SECTION_END = re.compile(r"^=+\s*\S.*=+$")
_TRACEBACK_START = re.compile(r"^Traceback \(most recent call last\):")
_EXCEPTION_LINE = re.compile(
    r"^(?:[A-Za-z_][\w.]*(?:Error|Exception|Failure|Timeout)\b|AssertionError\b|E\s{3,})"
)
_FILE_FRAME = re.compile(r'^\s*File ".*", line \d+, in .+$')
_SUMMARY_LINE = re.compile(r"^(?:FAILED|ERROR)\s+\S+")
_ERROR_MARKER = re.compile(r"^##\[error\]")
_USELESS_ERROR = re.compile(r"^##\[error\]Process completed with exit code 1\.?$")

_NORMALISERS: Tuple[Tuple[re.Pattern[str], str], ...] = (
    (re.compile(r"\bk8s-integration-\d+-[0-9a-f]{4,}(?:-[\w.-]+)?\b"), "<instance>"),
    (re.compile(r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:[.,]\d+)?Z?"), "<ts>"),
    (re.compile(r"\b(?:\d{1,3}\.){3}\d{1,3}\b"), "<ip>"),
    (re.compile(r"\b(?:[0-9a-f]{0,4}:){2,7}[0-9a-f]{0,4}\b"), "<ipv6>"),
    (
        re.compile(r"/home/runner/work/[\w.-]+/[\w.-]+(?:/[^\s:)]*)?"),
        "<workspace>",
    ),
    (
        re.compile(r"/home/ubuntu/actions-runner/_work/[\w.-]+/[\w.-]+(?:/[^\s:)]*)?"),
        "<workspace>",
    ),
    (re.compile(r"\bline \d+\b"), "line <n>"),
    (re.compile(r"\b\d+(\s+restarts?\b)"), r"<n>\1"),
    (re.compile(r"[ \t]+"), " "),
)


@dataclass(frozen=True)
class Excerpt:
    text: str
    source: str


@dataclass(frozen=True)
class Fingerprint:
    signature_id: str
    normalised: str
    excerpt: str
    source: str
    normaliser_version: int = NORMALISER_VERSION
    extractor_version: int = EXTRACTOR_VERSION


def _prepare_lines(log: str) -> Tuple[str, ...]:
    lines = []
    for line in log.splitlines():
        stripped = _GH_TIMESTAMP.sub("", line).rstrip()
        if stripped.strip() and set(stripped.strip()) == {"^"}:
            continue
        lines.append(stripped)
    return tuple(lines)


def _section(lines: Tuple[str, ...], header: re.Pattern[str]) -> Tuple[str, ...]:
    start: Optional[int] = None
    for index, line in enumerate(lines):
        if header.match(line):
            start = index + 1
            break
    if start is None:
        return ()

    out = []
    for line in lines[start:]:
        if _SECTION_END.match(line) and not header.match(line):
            break
        out.append(line)
    return tuple(out)


def _innermost_exception(block: Tuple[str, ...]) -> Tuple[str, ...]:
    for index, line in enumerate(block):
        if not _EXCEPTION_LINE.match(line):
            continue
        frame_indices = []
        cursor = index - 1
        while cursor >= 0 and not _TRACEBACK_START.match(block[cursor]):
            if _FILE_FRAME.match(block[cursor]):
                frame_indices.append(cursor)
                if len(frame_indices) == 2:
                    break
            cursor -= 1
        start = min(frame_indices) if frame_indices else index
        end = index + 1
        return tuple(block[start:end])
    return ()


def extract(log: str) -> Excerpt:
    """Extract the most diagnostic failure text from a raw job log."""

    lines = _prepare_lines(log)
    failures = _section(lines, _FAILURES_HEADER)
    if failures:
        root = _innermost_exception(failures)
        if root:
            return Excerpt("\n".join(root), "failures-block")

    errors = _section(lines, _ERRORS_HEADER)
    if errors:
        root = _innermost_exception(errors)
        if root:
            return Excerpt("\n".join(root), "errors-block")

    summary = tuple(
        line for line in _section(lines, _SUMMARY_HEADER) if _SUMMARY_LINE.match(line)
    )
    if summary:
        return Excerpt("\n".join(summary), "summary-line")

    informative_markers = [
        index
        for index, line in enumerate(lines)
        if _ERROR_MARKER.match(line) and not _USELESS_ERROR.match(line)
    ]
    if informative_markers:
        index = informative_markers[-1]
        return Excerpt(lines[index], "error-marker")

    tail = tuple(line for line in lines if line.strip())[-15:]
    return Excerpt("\n".join(tail), "log-tail")


def normalise(text: str) -> str:
    """Collapse run-to-run variation without blanket digit removal."""

    normalised_lines = []
    for line in text.splitlines():
        for pattern, replacement in _NORMALISERS:
            line = pattern.sub(replacement, line)
        line = line.strip()
        if line:
            normalised_lines.append(line)
    return "\n".join(normalised_lines)


def signature_id(normalised: str) -> str:
    """Return the deterministic signature ID for a normalised excerpt."""

    return hashlib.sha256(normalised.encode("utf-8")).hexdigest()[:SIGNATURE_LENGTH]


def fingerprint(log: str) -> Fingerprint:
    """Extract, scrub, normalise, and hash a job log."""

    extracted = extract(log)
    safe = scrub(extracted.text)
    normalised = normalise(safe)
    excerpt = safe.encode("utf-8")[:MAX_EXCERPT_BYTES].decode("utf-8", errors="ignore")
    return Fingerprint(
        signature_id=signature_id(normalised),
        normalised=normalised[:MAX_EXCERPT_BYTES],
        excerpt=excerpt,
        source=extracted.source,
    )
