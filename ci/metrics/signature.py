#
# Copyright 2026 Canonical, Ltd.
#
"""Excerpt extraction, normalisation, and signature hashing."""

from __future__ import annotations

import hashlib
import re
from dataclasses import dataclass
from typing import Iterable, Optional, Tuple

from metrics.scrub import scrub

# Bumping either version changes signature IDs; `fetch-logs --reextract`
# migrates stored jobs whose logs are still retained.
NORMALISER_VERSION = 2
EXTRACTOR_VERSION = 3
MAX_EXCERPT_BYTES = 4096
SIGNATURE_LENGTH = 16

_GH_TIMESTAMP = re.compile(r"^\d{4}-\d{2}-\d{2}T[\d:.]+Z\s?")
# pytest and the harness logger colour their output; left in, the codes stop
# section headers such as "short test summary info" from matching.
_ANSI_ESCAPE = re.compile(r"\x1b\[[0-9;?]*[A-Za-z]")
_CAPTURED_SECTION = re.compile(
    r"^-+ Captured (?P<stream>stdout|stderr|log) (?P<phase>setup|call|teardown) -+$"
)
_CALLED_PROCESS_ERROR = re.compile(r"^subprocess\.CalledProcessError\b")
# Harness log record for a failed command, e.g.
# "2026-10-13 02:40:01 [ WARNING]   stderr=error: unable to contact snap store".
_HARNESS_STDERR = re.compile(
    r"^\d{4}-\d{2}-\d{2} [\d:,.]+\s+\[?\s*[A-Z]+\]?\s+stderr=(?P<text>\S.*)$"
)
_STEP_HEADER = re.compile(r"^##\[group\]Run ")
_FAILURES_HEADER = re.compile(r"^=+\s*FAILURES\s*=+$")
_ERRORS_HEADER = re.compile(r"^=+\s*ERRORS\s*=+$")
_SUMMARY_HEADER = re.compile(r"^=+\s*short test summary info\s*=+$", re.IGNORECASE)
_SECTION_END = re.compile(r"^=+\s*\S.*=+$")
_TEST_SECTION = re.compile(r"^_{3,} .+ _{3,}$")
_TRACEBACK_START = re.compile(r"^Traceback \(most recent call last\):")
_EXCEPTION_LINE = re.compile(
    r"^(?:[A-Za-z_][\w.]*(?:Error|Exception|Failure|Timeout)\b|AssertionError\b|E\s{3,})"
)
_FILE_FRAME = re.compile(r'^\s*File ".*", line \d+, in .+$')
_TEST_FRAME = re.compile(r'^\s*File ".*", line \d+, in test_\w+$')
_SUMMARY_LINE = re.compile(r"^(?:FAILED|ERROR)\s+\S+")
_ERROR_MARKER = re.compile(r"^##\[error\]")
# Wrapper verdicts that say *that* a step failed, never *why*.
_GENERIC_ERROR = re.compile(
    r"^##\[error\](?:Process completed with exit code \d+|Failed after \d+ attempts)\.?$"
)
# Error-shaped lines tools print before a generic wrapper verdict: apt `E:`,
# CLI `Error:`/`fatal:`, Python exception lines, make and subprocess failures.
_DIAGNOSTIC_LINE = re.compile(
    r"^(?:E: \S"
    r"|(?:error|fatal|panic|Error|Fatal|ERROR|FATAL)(?:\[\w+\])?: \S"
    r"|[A-Za-z_][\w.]*(?:Error|Exception|Timeout)\b:? "
    r"|make: \*\*\* "
    r"|Command '.+' returned non-zero exit status \d+)"
)
_NON_DIAGNOSTIC = re.compile(r"^(?:ERROR|Error): Exception:?$")
# Any line a healthy job may print while recovering from an error (retries,
# probes). A rule whose pattern hits these on green jobs is not diagnostic.
_ERRORISH_LINE = re.compile(
    r"\b(?:error|errors|exception|failed|failure|fatal|panic|timed out|timeout)\b",
    re.IGNORECASE,
)

_NORMALISERS: Tuple[Tuple[re.Pattern[str], str], ...] = (
    (re.compile(r"\bk8s-integration-\d+-[0-9a-f]{4,}(?:-[\w.-]+)?\b"), "<instance>"),
    (re.compile(r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:[.,]\d+)?Z?"), "<ts>"),
    (
        re.compile(
            r"\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b", re.I
        ),
        "<uuid>",
    ),
    (re.compile(r"\b0x[0-9a-f]{6,}\b", re.I), "0x<addr>"),
    (re.compile(r"\b(?:\d{1,3}\.){3}\d{1,3}\b"), "<ip>"),
    (re.compile(r"\b(?:[0-9a-f]{0,4}:){2,7}[0-9a-f]{0,4}\b"), "<ipv6>"),
    # Only the runner-specific checkout prefix varies; the repo-relative path
    # after it names the failing file and must survive.
    (re.compile(r"/home/runner/work/[\w.-]+/[\w.-]+"), "<workspace>"),
    (re.compile(r"/home/ubuntu/actions-runner/_work/[\w.-]+/[\w.-]+"), "<workspace>"),
    (re.compile(r"\bpython3\.\d+\b"), "python3.<n>"),
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
        stripped = _ANSI_ESCAPE.sub("", _GH_TIMESTAMP.sub("", line)).rstrip()
        if stripped.strip() and set(stripped.strip()) == {"^"}:
            continue
        lines.append(stripped)
    return tuple(lines)


def _failed_step(lines: Tuple[str, ...]) -> Tuple[str, ...]:
    """Slice a job log to the step that raised the first ``##[error]``.

    Post-failure steps (artifact upload, cleanup) run under ``always()`` and
    would otherwise supply the excerpt's tail and error markers.
    """

    first_error = next(
        (index for index, line in enumerate(lines) if _ERROR_MARKER.match(line)), None
    )
    if first_error is None:
        return lines
    start = 0
    for index in range(first_error, -1, -1):
        if _STEP_HEADER.match(lines[index]):
            start = index
            break
    end = next(
        (
            index
            for index in range(first_error + 1, len(lines))
            if _STEP_HEADER.match(lines[index])
        ),
        len(lines),
    )
    return lines[start:end]


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


def _first_test(block: Tuple[str, ...]) -> Tuple[str, ...]:
    """Return the first ``____ test_name ____`` section of a pytest block."""

    starts = [index for index, line in enumerate(block) if _TEST_SECTION.match(line)]
    if not starts:
        return block
    first, end = starts[0], starts[1] if len(starts) > 1 else len(block)
    return block[first:end]


def _innermost_exception(block: Tuple[str, ...]) -> Tuple[str, ...]:
    block = _first_test(block)
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
        stop = index + 1
        root = tuple(block[start:stop])
        if any(_TEST_FRAME.match(frame) for frame in root):
            return root
        # A root cause raised inside a shared helper (a retry wrapper, util.run)
        # is identical for every test using it; the test frame that reached it
        # is what tells two such failures apart.
        anchor = next(
            (frame for frame in reversed(block) if _TEST_FRAME.match(frame)), None
        )
        return ((anchor.strip(),) if anchor else ()) + root
    return ()


def _captured(block: Tuple[str, ...], stream: str) -> Tuple[str, ...]:
    """Return a test's ``Captured <stream> call`` lines."""

    out = []
    inside = False
    for line in block:
        header = _CAPTURED_SECTION.match(line.strip())
        if header:
            inside = header.group("stream", "phase") == (stream, "call")
            continue
        if inside:
            out.append(line)
    return tuple(out)


def _command_stderr(block: Tuple[str, ...]) -> Optional[str]:
    """Return the failed command's own error line from a pytest test section.

    The harness logs a failed command's stderr before raising, and pytest
    replays it under "Captured log call"; older harness versions let it reach
    "Captured stderr call" instead.
    """

    logged = []
    for line in _captured(block, "log"):
        record = _HARNESS_STDERR.match(line.strip())
        if record:
            logged.append(record.group("text"))
    for source in (logged, _captured(block, "stderr")):
        diagnostic = diagnostic_lines(source)
        if diagnostic:
            return diagnostic[-1]
    return None


def _with_cause(root: Tuple[str, ...], block: Tuple[str, ...]) -> str:
    """Join ``root`` and, for a failed command, append the stderr that says why.

    ``CalledProcessError`` names the command but not the reason, so without
    this an unreachable snap store and a broken refresh hook look identical.
    """

    text = "\n".join(root)
    if _CALLED_PROCESS_ERROR.match(root[-1].strip()):
        cause = _command_stderr(_first_test(block))
        if cause:
            text += "\nstderr: " + cause
    return text


def extract(log: str) -> Excerpt:
    """Extract the most diagnostic failure text from a raw job log."""

    lines = _failed_step(_prepare_lines(log))
    failures = _section(lines, _FAILURES_HEADER)
    if failures:
        root = _innermost_exception(failures)
        if root:
            return Excerpt(_with_cause(root, failures), "failures-block")

    errors = _section(lines, _ERRORS_HEADER)
    if errors:
        root = _innermost_exception(errors)
        if root:
            return Excerpt(_with_cause(root, errors), "errors-block")

    summary = tuple(
        line for line in _section(lines, _SUMMARY_HEADER) if _SUMMARY_LINE.match(line)
    )
    if summary:
        return Excerpt("\n".join(summary), "summary-line")

    informative_markers = [
        index
        for index, line in enumerate(lines)
        if _ERROR_MARKER.match(line) and not _GENERIC_ERROR.match(line)
    ]
    if informative_markers:
        index = informative_markers[-1]
        return Excerpt(lines[index], "error-marker")

    diagnostic = diagnostic_lines(lines)
    if diagnostic:
        return Excerpt(diagnostic[-1], "diagnostic-line")

    tail = tuple(line for line in lines if line.strip())[-15:]
    return Excerpt("\n".join(tail), "log-tail")


def diagnostic_lines(lines: Iterable[str]) -> Tuple[str, ...]:
    """Return the error-shaped lines in ``lines``, stripped of indentation."""

    out = []
    for line in lines:
        stripped = line.strip()
        if _DIAGNOSTIC_LINE.match(stripped) and not _NON_DIAGNOSTIC.match(stripped):
            out.append(stripped)
    return tuple(out)


def error_lines(log: str) -> Tuple[str, ...]:
    """Return every error-shaped line of a whole log, for rule precision checks.

    Deliberately broader than extraction: a passing job's retried-then-healed
    errors are exactly the text a non-diagnostic rule would wrongly match.
    """

    out = []
    for line in _prepare_lines(log):
        stripped = line.strip()
        if _GENERIC_ERROR.match(stripped):
            continue
        errorish = _ERRORISH_LINE.search(stripped) or _EXCEPTION_LINE.match(stripped)
        if errorish or _DIAGNOSTIC_LINE.match(stripped):
            out.append(stripped)
    return tuple(dict.fromkeys(out))


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
