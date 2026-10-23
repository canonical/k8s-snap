#
# Copyright 2026 Canonical, Ltd.
#
"""Secret scrubbing for stored CI failure excerpts."""

from __future__ import annotations

import re
from dataclasses import dataclass
from typing import Iterable, Tuple

SCRUBBER_VERSION = 1
REDACTED = "<redacted>"


@dataclass(frozen=True)
class SecretFinding:
    line_number: int
    rule_name: str


@dataclass(frozen=True)
class Rule:
    name: str
    pattern: re.Pattern[str]
    replacement: str = REDACTED


def _case_insensitive(pattern: str) -> re.Pattern[str]:
    return re.compile(pattern, re.IGNORECASE)


SENSITIVE_NAME = (
    r"(?:token|secret|password|passwd|pwd|apikey|api_key|credential|"
    r"private_key|privatekey|auth|macaroon|session|cookie|certificate_key)"
)

TOKEN_RULES: Tuple[Rule, ...] = (
    Rule("github-token", _case_insensitive(r"\bgh[pousr]_[A-Za-z0-9]{16,}")),
    Rule("github-pat", _case_insensitive(r"\bgithub_pat_[A-Za-z0-9_]{20,}")),
    Rule(
        "jwt",
        _case_insensitive(
            r"\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}"
        ),
    ),
    Rule("aws-key-id", _case_insensitive(r"\b(?:AKIA|ASIA)[0-9A-Z]{16}\b")),
    Rule("slack-token", _case_insensitive(r"\bxox[abprs]-[A-Za-z0-9-]{10,}")),
    Rule(
        "snapcraft-macaroon",
        _case_insensitive(r"\b[A-Za-z0-9_-]{40,}={0,2}@[A-Za-z0-9.-]+\b"),
    ),
    Rule("k8s-bootstrap-token", _case_insensitive(r"\b[a-z0-9]{6}\.[a-z0-9]{16}\b")),
    Rule(
        "base64-blob",
        re.compile(r"(?=[A-Za-z0-9+/_-]*[A-Z])[A-Za-z0-9+/_-]{60,}={0,2}"),
    ),
    Rule(
        "url-userinfo",
        _case_insensitive(r"(://)[^\s/:@]+:[^\s/@]+(?=@)"),
        replacement=r"\1" + REDACTED,
    ),
    Rule(
        "authorization-header",
        _case_insensitive(r"\b(authorization\s*:\s*)\S+"),
        replacement=r"\1" + REDACTED,
    ),
    Rule(
        "basic-auth",
        _case_insensitive(r"\b((?:bearer|basic)\s+)[A-Za-z0-9._~+/=-]{12,}"),
        replacement=r"\1" + REDACTED,
    ),
)

ASSIGNMENT_RULES: Tuple[Rule, ...] = (
    Rule(
        "assignment-json",
        _case_insensitive(
            r"(\"[A-Za-z0-9_]{0,40}" + SENSITIVE_NAME + r"\"\s*:\s*\")[^\"]+"
        ),
        replacement=r"\1" + REDACTED,
    ),
    Rule(
        "assignment-eq",
        _case_insensitive(
            r"(\b[A-Za-z0-9_]{0,40}" + SENSITIVE_NAME + r"\s*=\s*)[^\s,;'\"`]+"
        ),
        replacement=r"\1" + REDACTED,
    ),
    Rule(
        "assignment-colon",
        _case_insensitive(r"(\b[A-Za-z0-9_]{0,40}" + SENSITIVE_NAME + r"\s*:\s+)\S+"),
        replacement=r"\1" + REDACTED,
    ),
    Rule(
        "cli-flag",
        _case_insensitive(r"(--" + SENSITIVE_NAME + r"(?:=|\s+))\S+"),
        replacement=r"\1" + REDACTED,
    ),
)

PEM_RULES: Tuple[Rule, ...] = (
    Rule(
        "private-key-block",
        _case_insensitive(r"-----BEGIN [A-Z ]*PRIVATE KEY-----.*"),
    ),
    Rule("certificate-block", _case_insensitive(r"-----BEGIN CERTIFICATE-----.*")),
    Rule(
        "kubeconfig-data",
        _case_insensitive(
            r"(\b(?:client-key-data|client-certificate-data|certificate-authority-data):\s*)\S+"
        ),
        replacement=r"\1" + REDACTED,
    ),
)

RULES: Tuple[Rule, ...] = PEM_RULES + TOKEN_RULES + ASSIGNMENT_RULES
_ALREADY_MASKED = re.compile(r"\*{3,}")


def scrub_line(line: str) -> str:
    """Redact suspected secrets from one line without dropping context."""

    result = line
    for rule in RULES:
        result = rule.pattern.sub(rule.replacement, result)
    return result


def scrub(text: str) -> str:
    """Redact suspected secrets from a text block."""

    return "\n".join(scrub_line(line) for line in text.splitlines())


def _match_is_already_safe(match: re.Match[str]) -> bool:
    payload = match.group(0)
    return REDACTED in payload or _ALREADY_MASKED.search(payload) is not None


def find_secrets(text: str) -> Tuple[SecretFinding, ...]:
    """Return secret-like findings in deterministic line and rule order."""

    findings = []
    for number, line in enumerate(text.splitlines(), start=1):
        for rule in RULES:
            if any(
                not _match_is_already_safe(match)
                for match in rule.pattern.finditer(line)
            ):
                findings.append(SecretFinding(number, rule.name))
                break
    return tuple(findings)


def has_secrets(texts: Iterable[str]) -> bool:
    """Return whether any supplied text contains an unredacted secret."""

    return any(find_secrets(text) for text in texts)
