#
# Copyright 2026 Canonical, Ltd.
#
"""Excerpt extraction, normalisation and signature hashing.

This is the module that turns "74 failed jobs" into "5 things to fix". It has
three responsibilities, in order:

1. **Extract** the few lines of a multi-hundred-kilobyte log that actually say
   why the job failed.
2. **Normalise** away everything that varies between two runs of the same
   failure -- timestamps, container names, addresses, paths, durations.
3. **Hash** the result into a stable ``signature_id``.

Extraction order is the load-bearing design decision here, and it is not the
obvious one. pytest's ``short test summary info`` line is the tempting source
because it is trivially greppable, but on this repository it is close to
worthless: the integration suite wraps waits in ``tenacity.Retrying``, so the
summary line reports ``tenacity.RetryError`` for a large share of failures.
Fingerprinting on it collapses unrelated bugs -- a broken upgrade path, a
dead kube-proxy, a slow image pull -- into one meaningless bucket.

The ``=== FAILURES ===`` block carries the full exception chain, and the
*innermost* exception is the discriminating one. For real job 101894429657 the
summary line says ``tenacity.RetryError`` while the innermost exception says
``AssertionError: Service kube-proxy should be active, but it is inactive``.
Only the second is a bug report. Extraction therefore walks the chain to its
root and prefers that, falling back to the summary line and then to the
generic runner error marker.
"""

import hashlib
import re
from typing import List, NamedTuple, Optional

from metrics.scrub import scrub

NORMALISER_VERSION = 1
EXTRACTOR_VERSION = 1

# Hard cap on stored evidence. Large enough for an exception chain plus
# context, small enough that 90 days of failures stay a rounding error.
MAX_EXCERPT_BYTES = 4096

SIGNATURE_LENGTH = 16

# GitHub prefixes every log line with an RFC3339 timestamp.
_GH_TIMESTAMP = re.compile(r"^\d{4}-\d{2}-\d{2}T[\d:.]+Z\s?")

_FAILURES_HEADER = re.compile(r"^=+\s*FAILURES\s*=+$")
_SUMMARY_HEADER = re.compile(r"^=+\s*short test summary info\s*=+$")
_SECTION_END = re.compile(r"^=+\s*\S.*=+$")
_TRACEBACK_START = re.compile(r"^Traceback \(most recent call last\):")
_EXCEPTION_LINE = re.compile(
    r"^(?:[A-Za-z_][\w.]*(?:Error|Exception|Failure|Timeout)\b|E\s{3,})"
)
_ERROR_MARKER = re.compile(r"^##\[error\]")
_USELESS_ERRORS = re.compile(r"^##\[error\]Process completed with exit code \d+\.?$")
_FAILED_LINE = re.compile(r"^FAILED\s+\S+")


class Excerpt(NamedTuple):
    """Extracted evidence plus where it came from."""

    text: str
    source: str


def strip_timestamps(lines: List[str]) -> List[str]:
    return [_GH_TIMESTAMP.sub("", line).rstrip() for line in lines]


def _section(lines: List[str], header: "re.Pattern[str]") -> List[str]:
    """Return the lines of a pytest ``=== NAME ===`` section."""
    start = None
    for index, line in enumerate(lines):
        if header.match(line):
            start = index + 1
            break
    if start is None:
        return []
    out: List[str] = []
    for line in lines[start:]:
        if _SECTION_END.match(line) and not header.match(line):
            break
        out.append(line)
    return out


def _innermost_exception(block: List[str]) -> Optional[List[str]]:
    """Pull the root exception and its immediate frame out of a FAILURES block.

    An exception chain is printed outermost-last: each ``Traceback`` is
    followed by its exception line, and the phrase "The above exception was the
    direct cause of the following exception" separates links. The first
    exception line in the block is therefore the root cause.

    The frame immediately above the exception is kept as well. Without it,
    ``AssertionError: assert False`` from two unrelated tests would hash
    identically; with it, the failing function name discriminates them.
    """
    for index, line in enumerate(block):
        if not _EXCEPTION_LINE.match(line):
            continue
        context_start = index
        # Walk back over the frame lines belonging to this traceback, keeping
        # at most the last two frames -- enough to identify the call site
        # without dragging in the whole tenacity/pytest scaffolding.
        frames = 0
        cursor = index - 1
        while cursor >= 0 and frames < 4:
            candidate = block[cursor]
            if _TRACEBACK_START.match(candidate):
                context_start = cursor + 1
                break
            if not candidate.strip():
                break
            context_start = cursor
            frames += 1
            cursor -= 1
        end = index + 1
        return block[context_start:end]
    return None


def extract(log: str) -> Excerpt:
    """Extract the most diagnostic excerpt available from a job log."""
    lines = strip_timestamps(log.splitlines())

    failures = _section(lines, _FAILURES_HEADER)
    if failures:
        root = _innermost_exception(failures)
        if root:
            return Excerpt("\n".join(root), "failures-block")

    summary = [
        line for line in _section(lines, _SUMMARY_HEADER) if _FAILED_LINE.match(line)
    ]
    if summary:
        return Excerpt("\n".join(summary), "summary-line")

    # Runner-level errors. "Process completed with exit code 1" carries no
    # information (confirmed against the annotations API, which returns only
    # that string), so it is used only as a last resort and with context.
    markers = [i for i, line in enumerate(lines) if _ERROR_MARKER.match(line)]
    informative = [i for i in markers if not _USELESS_ERRORS.match(lines[i])]
    if informative:
        index = informative[-1]
        start = max(0, index - 5)
        end = index + 1
        return Excerpt("\n".join(lines[start:end]), "error-marker")
    if markers:
        index = markers[-1]
        start = max(0, index - 40)
        context = [line for line in lines[start:index] if line.strip()]
        return Excerpt(
            "\n".join(context[-15:] + [lines[index]]), "error-marker-context"
        )

    tail = [line for line in lines[-40:] if line.strip()]
    return Excerpt("\n".join(tail[-15:]), "log-tail")


# Ordered normalisation rules. Each collapses a source of run-to-run variation
# that would otherwise fragment one failure into many signatures.
_NORMALISERS = (
    # Instance names embed a random suffix per session.
    (re.compile(r"\bk8s-integration-\d+-[0-9a-f]{4,}(-\w+)?"), "<instance>"),
    (
        re.compile(r"\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b"),
        "<uuid>",
    ),
    (re.compile(r"\b0x[0-9a-f]{6,}\b"), "<addr>"),
    (re.compile(r"\b(?:[0-9a-f]{2}:){5}[0-9a-f]{2}\b"), "<mac>"),
    (re.compile(r"\b(?:\d{1,3}\.){3}\d{1,3}(?:/\d{1,2})?\b"), "<ip>"),
    (re.compile(r"\b(?:[0-9a-f]{0,4}:){3,7}[0-9a-f]{0,4}\b"), "<ipv6>"),
    (re.compile(r"\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:[.,]\d+)?"), "<ts>"),
    # Absolute runner paths differ per runner and per workspace. Self-hosted
    # runners use /home/<user>/actions-runner/_work/<repo>/<repo> while
    # GitHub-hosted ones use /home/runner/work/<repo>/<repo>; both must
    # collapse to the same token or the identical failure hashes differently
    # depending on which runner happened to pick it up.
    (
        re.compile(r"/home/[\w.-]+/(?:actions-runner/_work|work)/[\w.-]+/[\w.-]+"),
        "<workspace>",
    ),
    (re.compile(r"/(?:tmp|var/tmp)/[\w.-]*(?:tmp|pytest)[\w.-]*"), "<tmp>"),
    # Line numbers move with every refactor; the file and function do not.
    (re.compile(r"\bline \d+\b"), "line <n>"),
    (re.compile(r"\bpid=\d+"), "pid=<n>"),
    (re.compile(r"\b\d+\.\d+\s?s(?:econds)?\b"), "<duration>"),
    (re.compile(r"\(\d+:\d{2}:\d{2}\)"), "<duration>"),
    (
        re.compile(r"\b\d+\s+(?:failed|passed|warnings?|errors?|skipped)\b"),
        "<n> \\g<0>",
    ),
    (re.compile(r"\b\d{4,}\b"), "<n>"),
    (re.compile(r"[ \t]+"), " "),
)


def normalise(text: str) -> str:
    """Collapse run-to-run variation so equal failures hash equally.

    Deliberately conservative about numbers: only long integers and clearly
    quantitative contexts are replaced. Blanket digit removal would fold
    "exit code 1" into "exit code 2" and merge genuinely different failures.
    """
    lines = []
    for line in text.splitlines():
        for pattern, replacement in _NORMALISERS:
            line = pattern.sub(replacement, line)
        stripped = line.strip()
        # Python 3.11+ prints a caret line under the failing sub-expression.
        # It carries no information the frame above does not, and its width
        # tracks identifier lengths, so it is dropped rather than hashed.
        if stripped and set(stripped) != {"^"}:
            lines.append(stripped)
    return "\n".join(lines)


def signature_id(normalised: str) -> str:
    """Stable fingerprint for a normalised excerpt."""
    digest = hashlib.sha256(normalised.encode("utf-8")).hexdigest()
    return digest[:SIGNATURE_LENGTH]


class Fingerprint(NamedTuple):
    signature_id: str
    normalised: str
    excerpt: str
    source: str
    normaliser_version: int = NORMALISER_VERSION
    extractor_version: int = EXTRACTOR_VERSION


def fingerprint(log: str) -> Fingerprint:
    """Full pipeline: extract, scrub, normalise, hash.

    Scrubbing happens before normalisation and before truncation, so no code
    path can persist an unscrubbed byte.
    """
    extracted = extract(log)
    safe = scrub(extracted.text)
    normalised = normalise(safe)
    truncated = safe.encode("utf-8")[:MAX_EXCERPT_BYTES].decode(
        "utf-8", errors="ignore"
    )
    return Fingerprint(
        signature_id=signature_id(normalised),
        normalised=normalised[:MAX_EXCERPT_BYTES],
        excerpt=truncated,
        source=extracted.source,
    )
