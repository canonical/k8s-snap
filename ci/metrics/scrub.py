#
# Copyright 2026 Canonical, Ltd.
#
"""Secret scrubbing for log excerpts.

Every excerpt stored by the metrics pipeline passes through :func:`scrub`
before it is written anywhere. This module exists because the pipeline does
something the rest of CI does not: it takes job logs, which are access-
controlled and expire after 90 days, and distils them into artifacts that are
kept and shared. Anything leaked here is leaked durably.

GitHub already masks registered secrets as ``***`` in log output, so this is a
second layer rather than the only one. It is needed because masking only
covers values GitHub knows about. It does not cover credentials that a test
*derives* at runtime -- kubeconfig contents, bootstrap tokens, client
certificates, join tokens minted by k8sd -- and those are exactly the things
that end up in an integration-test traceback.

Design stance: **fail closed**. Patterns are deliberately broad, and a false
positive costs a slightly less readable excerpt while a false negative costs a
credential. Where a pattern cannot be made specific, the whole line is
dropped. :func:`find_secrets` provides the audit path used by the release gate
("grep all committed excerpts -- zero findings").
"""

import re
from typing import List, NamedTuple, Tuple

SCRUBBER_VERSION = 1

REDACTED = "<redacted>"
DROPPED = "<line dropped: possible secret>"


class Rule(NamedTuple):
    name: str
    pattern: "re.Pattern[str]"
    # When True the entire line is discarded rather than partially rewritten.
    drop_line: bool = False
    # Replacement template. Patterns that need to keep a prefix capture it in
    # group 1 and use ``\1<redacted>``; Python's ``re`` only supports
    # fixed-width lookbehind, so prefixes are captured rather than asserted.
    replacement: str = REDACTED


def _c(pattern: str) -> "re.Pattern[str]":
    return re.compile(pattern, re.IGNORECASE)


# Token shapes. These are matched anywhere in a line and replaced in place,
# because the surrounding text is usually the useful part of the excerpt.
TOKEN_RULES: Tuple[Rule, ...] = (
    Rule("github-token", _c(r"\bgh[pousr]_[A-Za-z0-9]{16,}")),
    Rule("github-pat", _c(r"\bgithub_pat_[A-Za-z0-9_]{20,}")),
    # Three base64url segments -- a JWT, which is a bearer credential.
    Rule("jwt", _c(r"\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}")),
    Rule("aws-key-id", _c(r"\b(?:AKIA|ASIA)[0-9A-Z]{16}\b")),
    Rule("slack-token", _c(r"\bxox[abprs]-[A-Za-z0-9-]{10,}")),
    Rule("snapcraft-macaroon", _c(r"\b[A-Za-z0-9_-]{40,}={0,2}@[A-Za-z0-9.-]+\b")),
    # Kubernetes bootstrap tokens have a fixed, unmistakable shape.
    Rule("k8s-bootstrap-token", _c(r"\b[a-z0-9]{6}\.[a-z0-9]{16}\b")),
    # Credentials embedded in a URL's userinfo section.
    Rule(
        "url-userinfo",
        _c(r"(://)[^\s/:@]+:[^\s/@]+(?=@)"),
        replacement=r"\1" + REDACTED,
    ),
    Rule(
        "authorization-header",
        _c(r"\b(authorization\s*:\s*)\S+"),
        replacement=r"\1" + REDACTED,
    ),
    Rule(
        "basic-auth",
        _c(r"\b((?:bearer|basic)\s+)[A-Za-z0-9._~+/=-]{12,}"),
        replacement=r"\1" + REDACTED,
    ),
)

# Assignments. The *name* signals sensitivity, so the value is redacted
# whatever it looks like.
SENSITIVE_NAME = (
    r"(?:token|secret|password|passwd|pwd|apikey|api_key|credential|"
    r"private_key|privatekey|auth|macaroon|session|cookie|certificate_key)"
)

ASSIGNMENT_RULES: Tuple[Rule, ...] = (
    Rule(
        "assignment-json",
        _c(r"(\"[A-Za-z0-9_]{0,40}" + SENSITIVE_NAME + r"\"\s*:\s*\")[^\"]+"),
        replacement=r"\1" + REDACTED,
    ),
    Rule(
        "assignment-eq",
        _c(r"(\b[A-Za-z0-9_]{0,40}" + SENSITIVE_NAME + r"\s*=\s*)[^\s,;'\"]+"),
        replacement=r"\1" + REDACTED,
    ),
    Rule(
        "assignment-colon",
        _c(r"(\b[A-Za-z0-9_]{0,40}" + SENSITIVE_NAME + r"\s*:\s+)\S+"),
        replacement=r"\1" + REDACTED,
    ),
    Rule(
        "cli-flag",
        _c(r"(--" + SENSITIVE_NAME + r"[= ])\S+"),
        replacement=r"\1" + REDACTED,
    ),
)

# Whole-line kills. Used where no partial redaction can be trusted: the
# secret has no fixed boundary, or the line is a payload rather than a message.
LINE_RULES: Tuple[Rule, ...] = (
    Rule(
        "private-key-block", _c(r"-----BEGIN [A-Z ]*PRIVATE KEY-----"), drop_line=True
    ),
    Rule("certificate-block", _c(r"-----BEGIN CERTIFICATE-----"), drop_line=True),
    Rule(
        "kubeconfig-data",
        _c(r"\b(client-key-data|client-certificate-data|certificate-authority-data):"),
        drop_line=True,
    ),
    Rule(
        "cloud-init",
        _c(r"^\s*#cloud-config|\bcloud-init\b.*\b(password|token)\b"),
        drop_line=True,
    ),
    # A long unbroken base64-ish run is a payload, not a message. 60 chars is
    # above anything a normal traceback line contains but below a real blob.
    Rule("base64-blob", _c(r"[A-Za-z0-9+/]{60,}={0,2}"), drop_line=True),
)

# Values GitHub itself masked. Kept as-is: already safe, and useful context.
_ALREADY_MASKED = re.compile(r"\*{3,}")


def scrub_line(line: str) -> str:
    """Scrub a single line, returning the redacted form.

    Order matters: whole-line rules are evaluated first so a line that must be
    dropped is never partially rewritten and thereby made to look safe.
    """
    for rule in LINE_RULES:
        if rule.pattern.search(line):
            return DROPPED

    result = line
    for rule in TOKEN_RULES + ASSIGNMENT_RULES:
        result = rule.pattern.sub(rule.replacement, result)
    return result


def scrub(text: str) -> str:
    """Scrub a multi-line excerpt."""
    return "\n".join(scrub_line(line) for line in text.splitlines())


def find_secrets(text: str) -> List[Tuple[int, str]]:
    """Return ``(line_number, rule_name)`` for every suspected secret.

    This is the audit primitive. It is run over already-scrubbed, already-
    committed excerpts to prove the scrubber worked, so it must not depend on
    :func:`scrub` having been called.

    A match whose payload is already a redaction marker is not a finding:
    ``Authorization: <redacted>`` still matches the header rule by design, and
    reporting it would drown the audit in noise from its own output.
    """
    findings: List[Tuple[int, str]] = []
    for number, line in enumerate(text.splitlines(), start=1):
        if line == DROPPED:
            continue
        for rule in LINE_RULES + TOKEN_RULES + ASSIGNMENT_RULES:
            for match in rule.pattern.finditer(line):
                payload = match.group(0)
                if REDACTED in payload or _ALREADY_MASKED.search(payload):
                    continue
                findings.append((number, rule.name))
                break
    return findings
