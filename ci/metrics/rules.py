#
# Copyright 2026 Canonical, Ltd.
#
"""Stage B ordered rule-pack classification."""

from __future__ import annotations

import dataclasses
import logging
import re
import sqlite3
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Pattern, Sequence

import yaml

from metrics.models import ClassifiedBy, FailureClass

LOG = logging.getLogger(__name__)

DEFAULT_RULES_PATH = Path(__file__).resolve().parent.parent / "failure_rules.yaml"
MATCHABLE_JOB_FIELDS = frozenset(
    {
        "job_name",
        "caller_job",
        "os",
        "arch",
        "channel",
        "flavor",
        "substrate",
        "test_nodeid",
        "test_file",
        "failed_step_name",
    }
)


class RulePackError(ValueError):
    """Raised when a rule pack is malformed."""


@dataclasses.dataclass(frozen=True)
class RuleContext:
    """Evidence available to a Stage B rule."""

    signature_id: str
    excerpt: str
    job_name: str = ""
    caller_job: str = ""
    os: str = ""
    arch: str = ""
    channel: str = ""
    flavor: str = ""
    substrate: str = ""
    test_nodeid: str = ""
    test_file: str = ""
    failed_step_name: str = ""


@dataclasses.dataclass(frozen=True)
class Rule:
    """One ordered classification rule."""

    id: str
    failure_class: str
    notes: str
    subclass: Optional[str] = None
    confidence: float = 1.0
    signature_ids: tuple[str, ...] = ()
    excerpt_pattern: Optional[Pattern[str]] = None
    job_patterns: Dict[str, Pattern[str]] = dataclasses.field(default_factory=dict)

    @property
    def substrate_guard(self) -> Optional[str]:
        pattern = self.job_patterns.get("substrate")
        if pattern is None:
            return None
        match = re.fullmatch(r"^\^([A-Za-z0-9_.-]+)\$$", pattern.pattern)
        return match.group(1) if match else None

    def matches(self, context: RuleContext) -> bool:
        """Return whether all declared criteria match ``context``."""

        if self.signature_ids and context.signature_id not in self.signature_ids:
            return False
        if self.excerpt_pattern is not None and not self.excerpt_pattern.search(
            context.excerpt
        ):
            return False
        for field, pattern in self.job_patterns.items():
            value = getattr(context, field)
            if not value or not pattern.search(value):
                return False
        return True


@dataclasses.dataclass(frozen=True)
class RulePack:
    """An ordered, validated rule pack."""

    version: int
    rules: tuple[Rule, ...]
    source: Path

    def match(self, context: RuleContext) -> Optional[Rule]:
        for rule in self.rules:
            if rule.matches(context):
                return rule
        return None


def load_rules(path: Optional[Path] = None) -> RulePack:
    """Load and validate the ordered rule pack."""

    rules_path = Path(path) if path is not None else DEFAULT_RULES_PATH
    try:
        payload = yaml.safe_load(rules_path.read_text(encoding="utf-8")) or {}
    except yaml.YAMLError as exc:
        raise RulePackError(f"{rules_path}: invalid YAML: {exc}") from exc

    if not isinstance(payload, dict):
        raise RulePackError(f"{rules_path}: top level must be a mapping")
    version = payload.get("version")
    if not isinstance(version, int) or version < 1:
        raise RulePackError(f"{rules_path}: version must be a positive integer")
    raw_rules = payload.get("rules")
    if not isinstance(raw_rules, list):
        raise RulePackError(f"{rules_path}: rules must be a list")

    seen: Dict[str, int] = {}
    rules: List[Rule] = []
    for index, raw in enumerate(raw_rules):
        rule = _parse_rule(raw, index, rules_path)
        if rule.id in seen:
            raise RulePackError(
                f"{rules_path}: duplicate rule id {rule.id!r} at indexes "
                f"{seen[rule.id]} and {index}"
            )
        seen[rule.id] = index
        rules.append(rule)

    _warn_on_shadowed_rules(rules, rules_path)
    return RulePack(version=version, rules=tuple(rules), source=rules_path)


def _parse_rule(raw: Any, index: int, path: Path) -> Rule:
    where = f"{path}: rule[{index}]"
    if not isinstance(raw, dict):
        raise RulePackError(f"{where}: rule must be a mapping")

    rule_id = _required_string(raw, "id", where)
    failure_class = raw.get("failure_class", raw.get("class"))
    allowed_classes = {member.value for member in FailureClass}
    if failure_class not in allowed_classes:
        raise RulePackError(
            f"{where} ({rule_id}): invalid failure_class {failure_class!r}. "
            f"expected one of {sorted(allowed_classes)}"
        )
    notes = _required_string(raw, "notes", where)
    signature_ids = tuple(_as_list(raw.get("signature_id"), where, rule_id))
    for signature_id in signature_ids:
        if not re.fullmatch(r"[0-9a-f]{16}", signature_id):
            raise RulePackError(
                f"{where} ({rule_id}): signature_id {signature_id!r} is not "
                "a 16-character lowercase hex digest"
            )

    excerpt_pattern = _compile(raw.get("excerpt_pattern"), where, rule_id)
    job_patterns = _parse_job_patterns(raw.get("job") or {}, where, rule_id)
    if not signature_ids and excerpt_pattern is None and not job_patterns:
        raise RulePackError(
            f"{where} ({rule_id}): rule has no criteria and would match everything"
        )

    confidence = raw.get("confidence", 1.0)
    if not isinstance(confidence, (int, float)) or not 0 < confidence <= 1:
        raise RulePackError(
            f"{where} ({rule_id}): confidence must be in (0, 1], got {confidence!r}"
        )

    return Rule(
        id=rule_id,
        failure_class=failure_class,
        subclass=raw.get("subclass"),
        notes=notes,
        confidence=float(confidence),
        signature_ids=signature_ids,
        excerpt_pattern=excerpt_pattern,
        job_patterns=job_patterns,
    )


def _required_string(raw: Dict[str, Any], key: str, where: str) -> str:
    value = raw.get(key)
    if not isinstance(value, str) or not value.strip():
        raise RulePackError(f"{where}: {key} is required and must be a string")
    return value


def _as_list(value: Any, where: str, rule_id: str) -> List[str]:
    if value is None:
        return []
    if isinstance(value, str):
        return [value]
    if isinstance(value, list) and all(isinstance(item, str) for item in value):
        return list(value)
    raise RulePackError(f"{where} ({rule_id}): signature_id must be a string or list")


def _compile(value: Any, where: str, rule_id: str) -> Optional[Pattern[str]]:
    if value is None:
        return None
    if not isinstance(value, str) or not value.strip():
        raise RulePackError(f"{where} ({rule_id}): pattern must be a non-empty string")
    try:
        return re.compile(value, re.IGNORECASE | re.MULTILINE)
    except re.error as exc:
        raise RulePackError(f"{where} ({rule_id}): invalid regex {value!r}: {exc}")


def _parse_job_patterns(raw: Any, where: str, rule_id: str) -> Dict[str, Pattern[str]]:
    if not isinstance(raw, dict):
        raise RulePackError(f"{where} ({rule_id}): job must be a mapping")
    patterns = {}
    for field, pattern in raw.items():
        if field not in MATCHABLE_JOB_FIELDS:
            raise RulePackError(
                f"{where} ({rule_id}): unsupported job field {field!r}. "
                f"expected one of {sorted(MATCHABLE_JOB_FIELDS)}"
            )
        compiled = _compile(pattern, where, rule_id)
        if compiled is None:
            raise RulePackError(f"{where} ({rule_id}): job.{field} is empty")
        patterns[field] = compiled
    return patterns


def _warn_on_shadowed_rules(rules: Sequence[Rule], path: Path) -> None:
    for index, rule in enumerate(rules):
        if not rule.signature_ids:
            continue
        for earlier in rules[:index]:
            if not earlier.signature_ids:
                continue
            if not set(rule.signature_ids) & set(earlier.signature_ids):
                continue
            if _subsumes(earlier, rule):
                LOG.warning(
                    "%s: rule %r is unreachable behind earlier rule %r",
                    path,
                    rule.id,
                    earlier.id,
                )


def _subsumes(earlier: Rule, later: Rule) -> bool:
    if earlier.excerpt_pattern is not None:
        if later.excerpt_pattern is None:
            return False
        if earlier.excerpt_pattern.pattern != later.excerpt_pattern.pattern:
            return False
    for field, pattern in earlier.job_patterns.items():
        later_pattern = later.job_patterns.get(field)
        if later_pattern is None or later_pattern.pattern != pattern.pattern:
            return False
    return True


def _contexts(conn: sqlite3.Connection) -> Iterable[RuleContext]:
    rows = conn.execute(
        """
        SELECT s.signature_id, s.exemplar_excerpt, j.job_name,
               COALESCE(j.caller_job, ''), cfg.os, cfg.arch, cfg.channel,
               cfg.flavor, cfg.substrate, t.nodeid, COALESCE(t.file, ''),
               COALESCE(j.failed_step_name, '')
        FROM signatures s
        JOIN jobs j ON j.signature_id = s.signature_id
        JOIN configs cfg USING (config_id)
        JOIN tests t USING (test_id)
        WHERE j.router_rule_id IS NULL
        ORDER BY s.signature_id, cfg.substrate, j.job_id
        """
    )
    seen = set()
    for row in rows:
        key = (row[0], row[8])
        if key in seen:
            continue
        seen.add(key)
        yield RuleContext(
            signature_id=row[0],
            excerpt=row[1],
            job_name=row[2],
            caller_job=row[3],
            os=row[4],
            arch=row[5],
            channel=row[6],
            flavor=row[7],
            substrate=row[8],
            test_nodeid=row[9],
            test_file=row[10],
            failed_step_name=row[11],
        )


def classify_signatures(conn: sqlite3.Connection, pack: RulePack) -> Dict[str, int]:
    """Apply ``pack`` to stored signatures and write classification rows."""

    summary = {"matched": 0, "unmatched": 0}
    by_rule: Dict[str, int] = {}
    written_global = set()
    now = datetime.now(timezone.utc).isoformat()

    for context in _contexts(conn):
        if context.signature_id in written_global:
            continue
        rule = pack.match(context)
        if rule is None:
            summary["unmatched"] += 1
            continue

        substrate = rule.substrate_guard or "*"
        if substrate == "*":
            written_global.add(context.signature_id)
        conn.execute(
            """
            INSERT INTO signature_classifications (
                signature_id, substrate, ruleset_version, failure_class, subclass,
                rule_id, confidence, classified_by, classified_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(signature_id, substrate, ruleset_version, classified_by)
            DO UPDATE SET
                failure_class = excluded.failure_class,
                subclass = excluded.subclass,
                rule_id = excluded.rule_id,
                confidence = excluded.confidence,
                classified_at = excluded.classified_at
            """,
            (
                context.signature_id,
                substrate,
                pack.version,
                rule.failure_class,
                rule.subclass,
                rule.id,
                rule.confidence,
                ClassifiedBy.RULES.value,
                now,
            ),
        )
        summary["matched"] += 1
        by_rule[rule.id] = by_rule.get(rule.id, 0) + 1

    return {
        **summary,
        **{f"rule:{key}": value for key, value in sorted(by_rule.items())},
    }
