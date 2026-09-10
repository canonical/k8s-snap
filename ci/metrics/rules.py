#
# Copyright 2026 Canonical, Ltd.
#
"""Stage B: the rule pack -- deterministic classification of failure signatures.

Stage A (:mod:`metrics.router`) settles the failures it can decide from job
metadata alone. Everything it defers arrives here with a scrubbed evidence
excerpt and a ``signature_id``, and this module decides the class from an
ordered, reviewable rule file (``ci/failure_rules.yaml``).

Two design choices are load-bearing:

**Rules are ordered and first-match-wins.** Failure text is not mutually
exclusive -- a job that dies because LXD ran out of disk will happily print a
product-looking error on the way down. Ordering lets us state precedence
explicitly (infrastructure before product) instead of hoping the regexes never
overlap.

**A rule normally pins a ``signature_id``.** Pinned rules cannot drift: the
signature is a hash of the normalised failure text, so a pin means "this exact
failure", not "anything that looks a bit like it". Regex rules exist for
families that genuinely vary (disk-full messages quote a different path every
time) and are deliberately the exception, because a loose regex placed early in
the file silently swallows unrelated failures and makes the metrics lie.

Class and subclass are validated against :data:`metrics.models.SUBCLASSES` when
the pack loads, so a typo fails at load time rather than inventing a category
that quietly never matches anything.
"""

import dataclasses
import logging
import re
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Sequence

import yaml

from metrics.models import SUBCLASSES, ClassifiedBy, FailureClass, JobFailure

LOG = logging.getLogger(__name__)

RULESET_SCHEMA_VERSION = 1

DEFAULT_RULES_PATH = Path(__file__).resolve().parent.parent / "failure_rules.yaml"

# Job fields a rule may constrain. Restricted on purpose: allowing arbitrary
# attribute lookup would let a rule depend on a classification field and
# produce order-dependent results.
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


class RulePackError(Exception):
    """Raised when the rule pack is malformed. Always fatal -- never skipped.

    A rule pack that silently drops a broken rule would under-classify without
    telling anyone, which is the exact failure mode this system exists to stop.
    """


@dataclasses.dataclass
class Rule:
    """One classification rule."""

    id: str
    failure_class: str
    subclass: Optional[str] = None
    owner: Optional[str] = None
    issue: Optional[int] = None
    notes: Optional[str] = None
    confidence: float = 1.0

    signature_ids: List[str] = dataclasses.field(default_factory=list)
    excerpt_pattern: Optional[re.Pattern] = None
    job_patterns: Dict[str, re.Pattern] = dataclasses.field(default_factory=dict)

    def matches(self, job: JobFailure) -> bool:
        """True when every criterion this rule declares holds for ``job``.

        Criteria are ANDed. A rule with no criteria never matches; that is
        checked at load time, since such a rule would classify everything.
        """
        if self.signature_ids and job.signature_id not in self.signature_ids:
            return False

        if self.excerpt_pattern is not None:
            if not job.evidence_excerpt:
                return False
            if not self.excerpt_pattern.search(job.evidence_excerpt):
                return False

        for field, pattern in self.job_patterns.items():
            value = getattr(job, field, None)
            if not value or not pattern.search(str(value)):
                return False

        return True


@dataclasses.dataclass
class RulePack:
    """An ordered, validated collection of rules."""

    version: int
    rules: List[Rule]
    source: Optional[Path] = None

    def match(self, job: JobFailure) -> Optional[Rule]:
        for rule in self.rules:
            if rule.matches(job):
                return rule
        return None

    def apply(self, job: JobFailure) -> Optional[Rule]:
        """Classify ``job`` in place. Returns the rule used, or ``None``.

        Router verdicts are not overwritten: Stage A decided from metadata,
        which is stronger evidence than log text. A job whose log says
        "connection refused" because the runner vanished underneath it should
        stay ``infra.runner``.
        """
        if job.classified_by == ClassifiedBy.ROUTER.value:
            return None

        rule = self.match(job)
        if rule is None:
            return None

        job.failure_class = rule.failure_class
        job.subclass = rule.subclass
        job.confidence = rule.confidence
        job.classified_by = ClassifiedBy.RULES.value
        job.classifier_version = f"rules-v{self.version}"
        job.ruleset_version = self.version
        job.rule_id = rule.id
        return rule


def load_rules(path: Optional[Path] = None) -> RulePack:
    """Load and validate the rule pack.

    A missing file is not an error: before the pack is seeded, Stage B simply
    classifies nothing and every deferred failure stays ``unknown``, which is
    the honest answer.
    """
    path = Path(path) if path else DEFAULT_RULES_PATH
    if not path.exists():
        LOG.warning("No rule pack at %s; Stage B will classify nothing", path)
        return RulePack(version=0, rules=[], source=path)

    try:
        payload = yaml.safe_load(path.read_text()) or {}
    except yaml.YAMLError as exc:
        raise RulePackError(f"{path}: invalid YAML: {exc}") from exc

    if not isinstance(payload, dict):
        raise RulePackError(f"{path}: top level must be a mapping")

    version = payload.get("version")
    if not isinstance(version, int) or version < 1:
        raise RulePackError(f"{path}: 'version' must be a positive integer")

    raw_rules = payload.get("rules") or []
    if not isinstance(raw_rules, list):
        raise RulePackError(f"{path}: 'rules' must be a list")

    rules: List[Rule] = []
    seen_ids: Dict[str, int] = {}
    for index, raw in enumerate(raw_rules):
        rule = _parse_rule(raw, index, path)
        if rule.id in seen_ids:
            raise RulePackError(
                f"{path}: duplicate rule id {rule.id!r} "
                f"(also at index {seen_ids[rule.id]})"
            )
        seen_ids[rule.id] = index
        rules.append(rule)

    _warn_on_shadowed_rules(rules, path)
    LOG.debug("Loaded %s rule(s) from %s (v%s)", len(rules), path, version)
    return RulePack(version=version, rules=rules, source=path)


def _parse_rule(raw: Any, index: int, path: Path) -> Rule:
    where = f"{path}: rule[{index}]"
    if not isinstance(raw, dict):
        raise RulePackError(f"{where}: must be a mapping")

    rule_id = raw.get("id")
    if not rule_id or not isinstance(rule_id, str):
        raise RulePackError(f"{where}: 'id' is required and must be a string")

    failure_class = raw.get("class")
    valid_classes = {member.value for member in FailureClass}
    if failure_class not in valid_classes:
        raise RulePackError(
            f"{where} ({rule_id}): unknown class {failure_class!r} -- "
            f"expected one of {sorted(valid_classes)}"
        )

    subclass = raw.get("subclass")
    allowed = SUBCLASSES[FailureClass(failure_class)]
    if subclass is not None and subclass not in allowed:
        raise RulePackError(
            f"{where} ({rule_id}): subclass {subclass!r} is not valid for "
            f"class {failure_class!r} -- expected one of {allowed}"
        )

    signature_ids = _as_list(raw.get("signature_id"))
    for sig in signature_ids:
        if not re.fullmatch(r"[0-9a-f]{16}", sig):
            raise RulePackError(
                f"{where} ({rule_id}): signature_id {sig!r} is not a "
                "16-character lowercase hex digest"
            )

    excerpt_pattern = _compile(raw.get("excerpt_pattern"), where, rule_id)

    job_patterns: Dict[str, re.Pattern] = {}
    job_criteria = raw.get("job") or {}
    if not isinstance(job_criteria, dict):
        raise RulePackError(f"{where} ({rule_id}): 'job' must be a mapping")
    for field, pattern in job_criteria.items():
        if field not in MATCHABLE_JOB_FIELDS:
            raise RulePackError(
                f"{where} ({rule_id}): job field {field!r} is not matchable -- "
                f"expected one of {sorted(MATCHABLE_JOB_FIELDS)}"
            )
        compiled = _compile(pattern, where, rule_id)
        if compiled is None:
            raise RulePackError(f"{where} ({rule_id}): job.{field} is empty")
        job_patterns[field] = compiled

    if not signature_ids and excerpt_pattern is None and not job_patterns:
        raise RulePackError(
            f"{where} ({rule_id}): rule has no criteria and would match every "
            "failure; give it a signature_id, excerpt_pattern, or job matcher"
        )

    confidence = raw.get("confidence", 1.0)
    if not isinstance(confidence, (int, float)) or not 0 < confidence <= 1:
        raise RulePackError(
            f"{where} ({rule_id}): confidence must be in (0, 1], got {confidence!r}"
        )

    issue = raw.get("issue")
    if issue is not None and not isinstance(issue, int):
        raise RulePackError(f"{where} ({rule_id}): 'issue' must be an issue number")

    return Rule(
        id=rule_id,
        failure_class=failure_class,
        subclass=subclass,
        owner=raw.get("owner"),
        issue=issue,
        notes=raw.get("notes"),
        confidence=float(confidence),
        signature_ids=signature_ids,
        excerpt_pattern=excerpt_pattern,
        job_patterns=job_patterns,
    )


def _compile(pattern: Any, where: str, rule_id: str) -> Optional[re.Pattern]:
    if pattern is None:
        return None
    if not isinstance(pattern, str) or not pattern.strip():
        raise RulePackError(f"{where} ({rule_id}): pattern must be a non-empty string")
    try:
        return re.compile(pattern, re.IGNORECASE | re.MULTILINE)
    except re.error as exc:
        raise RulePackError(f"{where} ({rule_id}): invalid regex {pattern!r}: {exc}")


def _as_list(value: Any) -> List[str]:
    if value is None:
        return []
    if isinstance(value, str):
        return [value]
    if isinstance(value, list) and all(isinstance(v, str) for v in value):
        return list(value)
    raise RulePackError(f"expected a string or list of strings, got {value!r}")


def _warn_on_shadowed_rules(rules: Sequence[Rule], path: Path) -> None:
    """Warn when a pinned rule is unreachable behind an earlier rule.

    Only fires when the earlier rule *subsumes* the later one: it pins the
    same signature and adds no further criteria of its own. A broad rule that
    also constrains, say, the substrate is a deliberate and useful split (one
    signature meaning different things in different contexts), so warning
    about it would train people to ignore the warning.

    Not fatal, but a silently dead rule is worse than a noisy log line.
    """
    for i, rule in enumerate(rules):
        if not rule.signature_ids:
            continue
        for earlier in rules[:i]:
            if not earlier.signature_ids:
                continue
            if earlier.excerpt_pattern is not None or earlier.job_patterns:
                continue
            shadowed = set(rule.signature_ids) & set(earlier.signature_ids)
            if shadowed:
                LOG.warning(
                    "%s: rule %r is unreachable behind earlier rule %r for %s",
                    path,
                    rule.id,
                    earlier.id,
                    ", ".join(sorted(shadowed)),
                )


def classify_with_rules(jobs: Iterable[JobFailure], pack: RulePack) -> Dict[str, Any]:
    """Apply ``pack`` to every job and return a summary of what it settled."""
    matched = 0
    skipped_router = 0
    unmatched = 0
    by_rule: Dict[str, int] = {}

    for job in jobs:
        if job.classified_by == ClassifiedBy.ROUTER.value:
            skipped_router += 1
            continue
        rule = pack.apply(job)
        if rule is None:
            unmatched += 1
            continue
        matched += 1
        by_rule[rule.id] = by_rule.get(rule.id, 0) + 1

    return {
        "ruleset_version": pack.version,
        "matched": matched,
        "unmatched": unmatched,
        "skipped_router": skipped_router,
        "by_rule": dict(sorted(by_rule.items(), key=lambda kv: -kv[1])),
    }
