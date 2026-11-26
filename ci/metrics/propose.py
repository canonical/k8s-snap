#
# Copyright 2026 Canonical, Ltd.
#
"""Stage C: ask a model to propose rules for unclassified signatures.

The model only *proposes*. Every proposal is checked deterministically before
it can reach ``failure_rules.yaml``, and the result goes to a human as a pull
request:

1. the rule is well-formed, tier-prefixed, unpinned, with a bounded regex;
2. its examples are copied from the real evidence, not invented;
3. the pack still loads with the rule in its tier, so every example in the
   pack is still classified by its own rule (no shadowing);
4. it classifies the signature it was proposed for;
5. it does not change the class of anything already classified;
6. it hits at most 10% of successful runs of the same tests.

A rejected proposal is sent back to the model with the reason, once by
default; the model may also decline with ``"decision": "unknown"``, which is
preferred to a guess.
"""

from __future__ import annotations

import dataclasses
import json
import logging
import re
import sqlite3
import tempfile
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, Sequence, Tuple

from metrics.gh import GitHubError
from metrics.llm import LLMError, parse_json_reply
from metrics.precision import (
    DEFAULT_MAX_SUCCESS_HIT_RATE,
    DEFAULT_RUNS,
    DEFAULT_SAMPLE,
    PrecisionClient,
    check_rules,
    successful_jobs,
)
from metrics.rules import (
    MATCHABLE_JOB_FIELDS,
    RuleContext,
    RulePack,
    RulePackError,
    _contexts,
    load_rules,
)
from metrics.scrub import scrub
from metrics.signature import (
    _ERRORS_HEADER,
    _FAILURES_HEADER,
    _failed_step,
    _first_test,
    _prepare_lines,
    _section,
    diagnostic_lines,
)

LOG = logging.getLogger(__name__)

DEFAULT_LIMIT = 5
DEFAULT_ATTEMPTS = 2
MAX_PATTERN_CHARS = 200
MAX_NOTES_CHARS = 800
MAX_LOG_CHARS = 12000
MAX_LINE_CHARS = 400
SKILL_PATH = (
    Path(__file__).resolve().parents[2]
    / ".agents"
    / "skills"
    / "ci-failure-rules"
    / "SKILL.md"
)

TIERS = (
    ("infra.runner", "infra-"),
    ("infra.provisioning", "infra-"),
    ("external.dependency", "external-"),
    ("ci.config", "ci-"),
    ("product.bug", "product-"),
    ("test.bug", "test-"),
)
TIER_INDEX = {failure_class: index for index, (failure_class, _) in enumerate(TIERS)}
TIER_PREFIX = dict(TIERS)
RULE_KEYS = (
    "id",
    "failure_class",
    "subclass",
    "confidence",
    "excerpt_pattern",
    "job",
    "examples",
    "notes",
)
_RULE_ID = re.compile(r"^(?:infra|external|ci|product|test)-[a-z0-9]+(?:-[a-z0-9]+)*$")
_SUBCLASS = re.compile(r"^[a-z][a-z0-9_]{1,40}$")
# Nested quantifiers such as (a+)+ or (.*)* backtrack catastrophically.
_NESTED_QUANTIFIER = re.compile(r"\((?:[^()\\]|\\.)*[*+}]\)\s*[*+{]")

SYSTEM_PROMPT = """\
You maintain the CI failure classification rules of the canonical/k8s-snap
repository. You are shown one failure signature that no rule classifies yet,
with its evidence, and the current rule pack. Propose ONE rule that
classifies it, or decline.

Reply with a single JSON object and nothing else:

{"decision": "rule", "reason": "<what the evidence shows, 1-3 sentences>",
 "rule": {"id": "...", "failure_class": "...", "subclass": "...",
          "confidence": 1.0, "excerpt_pattern": "...",
          "job": {"<field>": "<regex>"},
          "examples": ["<excerpt text>"] or [{"excerpt": "...", "job": {...}}],
          "notes": "<why this class, citing the evidence>"}}

or {"decision": "unknown", "reason": "<what evidence is missing>"}.

Hard constraints (proposals breaking them are rejected automatically):
- Decline when the evidence does not show where the fix belongs. A wrong
  class files a misleading issue; "unknown" stays visible.
- id is lowercase-hyphenated and starts with the tier prefix: infra-,
  external-, ci-, product- or test-; it must not exist in the pack yet.
- failure_class is one of: infra.runner, infra.provisioning,
  external.dependency, ci.config, product.bug, test.bug.
- excerpt_pattern is a single-line Python regex (case-insensitive,
  multiline) of at most 200 characters, matched against the excerpt. It must
  match the target excerpt and name the cause, not a generic symptom such as
  "Failed to meet condition" or "exit status 1" alone; scope a generic
  message to one test with job.test_nodeid instead. No nested quantifiers.
- job is optional; its keys are job metadata fields: {fields}.
- examples are copied verbatim from the target excerpt (whole lines, may be
  trimmed to the lines that matter); when job guards are used, give examples
  as objects carrying the matching job fields.
- Never use signature_id.
- The rule must not change the class of any signature another rule already
  classifies, and it is checked against successful runs of the same tests:
  more than 10% hits there means it matches noise, not the cause.

The authoring guide used by the team follows.
"""


@dataclasses.dataclass(frozen=True)
class Target:
    signature_id: str
    excerpt: str
    jobs: int
    runs: int
    substrates: Tuple[str, ...]
    test_nodeids: Tuple[str, ...]
    job_names: Tuple[str, ...]
    failed_steps: Tuple[str, ...]
    job_urls: Tuple[str, ...]
    exemplar_job_id: Optional[int]
    first_seen: str
    last_seen: str


@dataclasses.dataclass
class Proposal:
    signature_id: str
    jobs: int
    decision: str = "error"
    reason: str = ""
    rule_id: Optional[str] = None
    failure_class: Optional[str] = None
    covered_signatures: int = 0
    covered_jobs: int = 0
    precision: str = ""
    attempts: int = 0
    rule: Optional[Dict[str, Any]] = None
    job_url: str = ""

    def as_dict(self) -> Dict[str, Any]:
        return dataclasses.asdict(self)


class ProposalRejected(ValueError):
    """A proposal failed one of the deterministic checks."""


def unknown_targets(conn: sqlite3.Connection) -> List[Target]:
    """Return unclassified signatures, most failed jobs first."""

    rows = conn.execute("""
        SELECT j.signature_id, s.exemplar_excerpt, s.exemplar_job_id,
               COUNT(*) AS jobs, COUNT(DISTINCT j.run_id) AS runs,
               json_group_array(DISTINCT cfg.substrate),
               json_group_array(DISTINCT t.nodeid),
               json_group_array(DISTINCT j.job_name),
               json_group_array(DISTINCT COALESCE(j.failed_step_name, '')),
               json_group_array(DISTINCT j.html_url),
               MIN(COALESCE(r.started_at, r.completed_at)),
               MAX(COALESCE(r.started_at, r.completed_at))
        FROM v_job_class v
        JOIN jobs j USING (job_id)
        JOIN signatures s ON s.signature_id = j.signature_id
        JOIN configs cfg ON cfg.config_id = j.config_id
        JOIN tests t ON t.test_id = j.test_id
        JOIN runs r ON r.run_id = j.run_id AND r.attempt = j.attempt
        WHERE v.failure_class = 'unknown' AND j.signature_id IS NOT NULL
        GROUP BY j.signature_id
        ORDER BY jobs DESC, j.signature_id
        """).fetchall()

    def values(raw: str, limit: int = 5) -> Tuple[str, ...]:
        return tuple(sorted(value for value in json.loads(raw) if value))[:limit]

    return [
        Target(
            signature_id=row[0],
            excerpt=row[1],
            exemplar_job_id=row[2],
            jobs=row[3],
            runs=row[4],
            substrates=values(row[5]),
            test_nodeids=values(row[6]),
            job_names=values(row[7], 3),
            failed_steps=values(row[8], 3),
            job_urls=values(row[9], 3),
            first_seen=row[10] or "",
            last_seen=row[11] or "",
        )
        for row in rows
    ]


def log_evidence(log: str) -> str:
    """Return the scrubbed part of a job log that explains its failure."""

    step = _failed_step(_prepare_lines(log))
    block = _first_test(_section(step, _FAILURES_HEADER)) or _first_test(
        _section(step, _ERRORS_HEADER)
    )
    lines = list(block or step)
    if len(lines) > 160:
        errors = list(dict.fromkeys(diagnostic_lines(lines)))[-40:]
        lines = lines[:20] + ["[...]"] + errors + ["[...]"] + lines[-80:]
    text = "\n".join(scrub(line)[:MAX_LINE_CHARS] for line in lines)
    return text[-MAX_LOG_CHARS:]


def _job_fields(target: Target) -> Dict[str, str]:
    return {
        "substrate": ", ".join(target.substrates) or "(none)",
        "test_nodeid": ", ".join(target.test_nodeids) or "(none)",
        "job_name": " | ".join(target.job_names),
        "failed_step_name": ", ".join(target.failed_steps) or "(none)",
    }


def _pack_summary(pack: RulePack) -> str:
    lines = []
    for rule in pack.rules:
        pattern = rule.excerpt_pattern.pattern if rule.excerpt_pattern else "-"
        guards = ", ".join(
            f"{field}={value.pattern!r}" for field, value in rule.job_patterns.items()
        )
        lines.append(
            f"- {rule.id} [{rule.failure_class}/{rule.subclass}] "
            f"pattern={pattern!r}" + (f" job: {guards}" if guards else "")
        )
    return "\n".join(lines)


def build_messages(
    target: Target, evidence: str, pack: RulePack, skill: str
) -> List[Dict[str, str]]:
    fields = "\n".join(
        f"  {key}: {value}" for key, value in _job_fields(target).items()
    )
    user = f"""\
## Current rule pack, in match order (first match wins)

{_pack_summary(pack)}

## Target signature {target.signature_id}

{target.jobs} failed jobs in {target.runs} runs, {target.first_seen} to {target.last_seen}.
Job metadata:
{fields}

### Excerpt (the exact text rules match)

```
{target.excerpt}
```

### Failed step from the exemplar job log (scrubbed, trimmed)

```
{evidence or "(log no longer available)"}
```
"""
    system = SYSTEM_PROMPT.replace("{fields}", ", ".join(sorted(MATCHABLE_JOB_FIELDS)))
    return [
        {"role": "system", "content": system + "\n" + skill},
        {"role": "user", "content": user},
    ]


def _check_regex(value: Any, where: str) -> str:
    if not isinstance(value, str) or not value.strip():
        raise ProposalRejected(f"{where} must be a non-empty string")
    if len(value) > MAX_PATTERN_CHARS:
        raise ProposalRejected(f"{where} is longer than {MAX_PATTERN_CHARS} chars")
    if "\n" in value or "\\n" in value:
        raise ProposalRejected(f"{where} must be a single-line pattern")
    if _NESTED_QUANTIFIER.search(value):
        raise ProposalRejected(f"{where} has a nested quantifier")
    try:
        re.compile(value)
    except re.error as exc:
        raise ProposalRejected(f"{where} is not a valid regex: {exc}") from None
    return value


def _flatten(text: str) -> str:
    return " ".join(text.split())


def sanitise_rule(raw: Any, pack: RulePack, corpus: str) -> Dict[str, Any]:
    """Return a clean rule dict, or raise ``ProposalRejected`` with the reason."""

    if not isinstance(raw, dict):
        raise ProposalRejected("rule must be an object")
    extra = set(raw) - set(RULE_KEYS)
    if extra:
        raise ProposalRejected(f"unsupported rule keys: {sorted(extra)}")

    rule_id = raw.get("id")
    if not isinstance(rule_id, str) or not _RULE_ID.match(rule_id):
        raise ProposalRejected(f"id {rule_id!r} is not tier-prefixed kebab-case")
    if any(rule.id == rule_id for rule in pack.rules):
        raise ProposalRejected(f"id {rule_id!r} already exists")
    failure_class = raw.get("failure_class")
    if failure_class not in TIER_INDEX:
        raise ProposalRejected(f"failure_class {failure_class!r} is not allowed")
    if not rule_id.startswith(TIER_PREFIX[failure_class]):
        raise ProposalRejected(
            f"id {rule_id!r} must start with {TIER_PREFIX[failure_class]!r}"
        )
    subclass = raw.get("subclass")
    if not isinstance(subclass, str) or not _SUBCLASS.match(subclass):
        raise ProposalRejected(f"subclass {subclass!r} must be short snake_case")
    confidence = raw.get("confidence", 1.0)
    if not isinstance(confidence, (int, float)) or not 0.5 <= confidence <= 1:
        raise ProposalRejected("confidence must be between 0.5 and 1")
    notes = raw.get("notes")
    if not isinstance(notes, str) or not notes.strip():
        raise ProposalRejected("notes are required")

    clean: Dict[str, Any] = {
        "id": rule_id,
        "failure_class": failure_class,
        "subclass": subclass,
    }
    if confidence != 1:
        clean["confidence"] = round(float(confidence), 2)
    clean["excerpt_pattern"] = _check_regex(
        raw.get("excerpt_pattern"), "excerpt_pattern"
    )
    job = raw.get("job") or {}
    if not isinstance(job, dict) or set(job) - MATCHABLE_JOB_FIELDS:
        raise ProposalRejected(f"job keys must be among {sorted(MATCHABLE_JOB_FIELDS)}")
    if job:
        clean["job"] = {
            field: _check_regex(value, f"job.{field}") for field, value in job.items()
        }

    examples = raw.get("examples")
    if not isinstance(examples, list) or not examples:
        raise ProposalRejected("at least one example is required")
    flat_corpus = _flatten(corpus)
    clean_examples: List[Any] = []
    for index, example in enumerate(examples):
        excerpt = example.get("excerpt") if isinstance(example, dict) else example
        if not isinstance(excerpt, str) or not excerpt.strip():
            raise ProposalRejected(f"examples[{index}] has no excerpt text")
        for line in excerpt.splitlines():
            if line.strip() and _flatten(line) not in flat_corpus:
                raise ProposalRejected(
                    f"examples[{index}] is not copied from the evidence: "
                    f"{line.strip()[:120]!r}"
                )
        if isinstance(example, dict):
            fields = example.get("job") or {}
            if not isinstance(fields, dict) or set(fields) - MATCHABLE_JOB_FIELDS:
                raise ProposalRejected(f"examples[{index}].job has unknown fields")
            clean_examples.append({"excerpt": excerpt, "job": dict(fields)})
        else:
            clean_examples.append(excerpt)
    clean["examples"] = clean_examples
    clean["notes"] = _flatten(notes)[:MAX_NOTES_CHARS]
    return clean


def _tier_position(rules: Sequence[Dict[str, Any]], failure_class: str) -> int:
    tier = TIER_INDEX[failure_class]
    for index, rule in enumerate(rules):
        if TIER_INDEX.get(rule.get("failure_class"), len(TIERS)) > tier:
            return index
    return len(rules)


def candidate_pack(pack_text: str, rule: Dict[str, Any]) -> RulePack:
    """Load the pack with ``rule`` inserted into its tier."""

    import yaml

    payload = yaml.safe_load(pack_text)
    rules = payload["rules"]
    rules.insert(_tier_position(rules, rule["failure_class"]), rule)
    with tempfile.TemporaryDirectory() as tmp:
        path = Path(tmp) / "failure_rules.yaml"
        path.write_text(yaml.safe_dump(payload, sort_keys=False), encoding="utf-8")
        try:
            return load_rules(path)
        except RulePackError as exc:
            reason = str(exc).replace(f"{path}: ", "")
            raise ProposalRejected(f"the pack would not load: {reason}") from None


def _contexts_by_signature(
    contexts: Sequence[RuleContext],
) -> Dict[str, List[RuleContext]]:
    grouped: Dict[str, List[RuleContext]] = {}
    for context in contexts:
        grouped.setdefault(context.signature_id, []).append(context)
    return grouped


def _job_count(conn: sqlite3.Connection, signature_id: str, substrate: str) -> int:
    row = conn.execute(
        """
        SELECT COUNT(*) FROM jobs j JOIN configs cfg USING (config_id)
        WHERE j.signature_id = ? AND cfg.substrate = ? AND j.router_rule_id IS NULL
        """,
        (signature_id, substrate),
    ).fetchone()
    return int(row[0])


def check_impact(
    conn: sqlite3.Connection,
    contexts: Sequence[RuleContext],
    old: RulePack,
    new: RulePack,
    rule_id: str,
    target: str,
) -> Tuple[int, int]:
    """Return ``(signatures, jobs)`` newly classified by ``rule_id``.

    Rejects the proposal if it misses its target or moves any context that
    another rule already classifies.
    """

    moved = []
    covered: Dict[str, int] = {}
    for context in contexts:
        before, after = old.match(context), new.match(context)
        if before is not None and (after is None or after.id != before.id):
            moved.append(f"{context.signature_id} ({before.id})")
        elif before is None and after is not None and after.id == rule_id:
            covered[context.signature_id] = covered.get(
                context.signature_id, 0
            ) + _job_count(conn, context.signature_id, context.substrate)
    if moved:
        raise ProposalRejected(
            "it would reclassify signatures other rules own: " + ", ".join(moved[:3])
        )
    if target not in covered:
        raise ProposalRejected(
            "it does not match the target signature's excerpt and job metadata"
        )
    return len(covered), sum(covered.values())


def render_rule_yaml(rule: Dict[str, Any]) -> str:
    """Render one rule as an indented list item in the pack's style."""

    import yaml

    class _Dumper(yaml.SafeDumper):
        def increase_indent(self, flow: bool = False, indentless: bool = False) -> Any:
            # Indent list items under their key, as the hand-written pack does.
            return super().increase_indent(flow, False)

    def _str(dumper: Any, value: str) -> Any:
        style = "|" if "\n" in value else None
        return dumper.represent_scalar("tag:yaml.org,2002:str", value, style=style)

    _Dumper.add_representer(str, _str)
    text = yaml.dump(
        [rule],
        Dumper=_Dumper,
        sort_keys=False,
        width=80,
        allow_unicode=True,
        default_flow_style=False,
    )
    return "".join(
        "  " + line if line.strip() else line for line in text.splitlines(True)
    )


def insert_rules(pack_text: str, rules: Sequence[Dict[str, Any]]) -> str:
    """Insert ``rules`` into the pack text at their tiers and bump ``version``."""

    import yaml

    for rule in rules:
        payload = yaml.safe_load(pack_text)
        position = _tier_position(payload["rules"], rule["failure_class"])
        lines = pack_text.splitlines(True)
        block = "\n" + render_rule_yaml(rule)
        if position < len(payload["rules"]):
            next_id = payload["rules"][position]["id"]
            anchor = next(
                index
                for index, line in enumerate(lines)
                if re.match(rf"^  - id: ['\"]?{re.escape(next_id)}['\"]?\s*$", line)
            )
            # Insert above the next tier's banner comment, not between the
            # banner and its first rule.
            while anchor > 0 and (
                not lines[anchor - 1].strip()
                or lines[anchor - 1].lstrip().startswith("#")
            ):
                anchor -= 1
            lines.insert(anchor, block)
        else:
            if lines and not lines[-1].endswith("\n"):
                lines[-1] += "\n"
            lines.append(block)
        pack_text = "".join(lines)
    if rules:
        pack_text = re.sub(
            r"^version: (\d+)[ \t]*$",
            lambda match: f"version: {int(match.group(1)) + 1}",
            pack_text,
            count=1,
            flags=re.MULTILINE,
        )
    return pack_text


Completer = Callable[[List[Dict[str, str]]], str]


def propose_rules(
    conn: sqlite3.Connection,
    pack_path: Path,
    complete: Completer,
    client: Optional[PrecisionClient],
    *,
    limit: int = DEFAULT_LIMIT,
    attempts: int = DEFAULT_ATTEMPTS,
    skill: Optional[str] = None,
    sample: int = DEFAULT_SAMPLE,
    runs: int = DEFAULT_RUNS,
    max_success_hit_rate: float = DEFAULT_MAX_SUCCESS_HIT_RATE,
) -> Tuple[List[Proposal], str]:
    """Propose and check rules for up to ``limit`` unknown signatures.

    Returns the per-signature outcomes and the pack text with every accepted
    rule inserted (unchanged when nothing was accepted). Accepted rules are
    checked against each other: each later proposal sees the earlier ones.
    """

    original_text = pack_path.read_text(encoding="utf-8")
    pack_text = original_text
    pack = load_rules(pack_path)
    if skill is None:
        skill = SKILL_PATH.read_text(encoding="utf-8") if SKILL_PATH.exists() else ""
    contexts = list(_contexts(conn))
    by_signature = _contexts_by_signature(contexts)
    healthy = successful_jobs(conn, client, runs) if client is not None else {}
    log_cache: Dict[int, Optional[Tuple[str, ...]]] = {}
    accepted: List[Dict[str, Any]] = []
    proposals: List[Proposal] = []

    for target in unknown_targets(conn):
        if len(proposals) >= limit:
            break
        if any(pack.match(ctx) for ctx in by_signature.get(target.signature_id, [])):
            continue  # classified by a rule accepted earlier in this run
        proposal = Proposal(
            signature_id=target.signature_id,
            jobs=target.jobs,
            job_url=target.job_urls[0] if target.job_urls else "",
        )
        proposals.append(proposal)
        evidence = _fetch_evidence(client, target)
        corpus = target.excerpt + "\n" + evidence
        messages = build_messages(target, evidence, pack, skill)
        for attempt in range(1, attempts + 1):
            proposal.attempts = attempt
            try:
                reply = complete(messages)
                answer = parse_json_reply(reply)
            except LLMError as exc:
                proposal.decision, proposal.reason = "error", str(exc)
                break
            proposal.reason = _flatten(str(answer.get("reason", "")))[:MAX_NOTES_CHARS]
            if answer.get("decision") != "rule":
                proposal.decision = "unknown"
                break
            try:
                rule = sanitise_rule(answer.get("rule"), pack, corpus)
                new_pack = candidate_pack(pack_text, rule)
                signatures, jobs = check_impact(
                    conn, contexts, pack, new_pack, rule["id"], target.signature_id
                )
                verdict = _precision(
                    conn,
                    new_pack,
                    client,
                    rule["id"],
                    healthy,
                    log_cache,
                    sample=sample,
                    runs=runs,
                    max_success_hit_rate=max_success_hit_rate,
                )
            except ProposalRejected as exc:
                proposal.decision = "rejected"
                proposal.reason = f"{exc} (model: {proposal.reason})"
                proposal.rule = (
                    answer.get("rule") if isinstance(answer.get("rule"), dict) else None
                )
                messages = messages + [
                    {"role": "assistant", "content": reply},
                    {
                        "role": "user",
                        "content": f"Rejected: {exc}. Fix the rule, or reply "
                        'with {"decision": "unknown", "reason": "..."}.',
                    },
                ]
                continue
            proposal.decision = "accepted"
            proposal.rule = rule
            proposal.rule_id = rule["id"]
            proposal.failure_class = rule["failure_class"]
            proposal.covered_signatures = signatures
            proposal.covered_jobs = jobs
            proposal.precision = verdict
            accepted.append(rule)
            pack_text = insert_rules(pack_text, [rule])
            pack = new_pack
            break

    if accepted:
        final_text = insert_rules(original_text, accepted)
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "failure_rules.yaml"
            path.write_text(final_text, encoding="utf-8")
            load_rules(path)
        return proposals, final_text
    return proposals, original_text


def _fetch_evidence(client: Optional[PrecisionClient], target: Target) -> str:
    if client is None or target.exemplar_job_id is None:
        return ""
    try:
        _status, log = client.download_job_log(int(target.exemplar_job_id))
    except GitHubError as exc:
        LOG.warning("Could not fetch log for job %s: %s", target.exemplar_job_id, exc)
        return ""
    return log_evidence(log) if log else ""


def _precision(
    conn: sqlite3.Connection,
    pack: RulePack,
    client: Optional[PrecisionClient],
    rule_id: str,
    healthy: Dict[str, List[int]],
    log_cache: Dict[int, Optional[Tuple[str, ...]]],
    **kwargs: Any,
) -> str:
    if client is None:
        return "unchecked (offline)"
    (report,) = check_rules(
        conn,
        pack,
        client,
        only={rule_id},
        healthy=healthy,
        log_cache=log_cache,
        **kwargs,
    )
    if report.verdict == "imprecise":
        raise ProposalRejected(
            f"it matches {report.success_hits}/{report.success_checked} successful "
            f"jobs, e.g. {report.examples[0] if report.examples else '?'!r}"
        )
    if report.verdict == "unchecked":
        return "unchecked (no successful runs of these tests sampled)"
    return f"{report.success_hits}/{report.success_checked} green hits"


def render_report(proposals: Sequence[Proposal], *, model: str = "") -> str:
    """Render the run as Markdown for a job summary or PR body."""

    accepted = [p for p in proposals if p.decision == "accepted"]
    lines = [
        "## CI rule proposals",
        "",
        f"{len(accepted)} of {len(proposals)} unclassified signatures got a rule"
        + (f" (model `{model}`)." if model else "."),
        "Every rule below passed the deterministic checks in `ci/metrics/propose.py`;"
        " review the class and the notes before merging.",
        "",
        "| Signature | Jobs | Outcome | Rule | Covers | Precision |",
        "|---|---:|---|---|---|---|",
    ]
    for p in proposals:
        signature = (
            f"[`{p.signature_id}`]({p.job_url})" if p.job_url else f"`{p.signature_id}`"
        )
        covers = (
            f"{p.covered_signatures} sig / {p.covered_jobs} jobs"
            if p.decision == "accepted"
            else ""
        )
        rule = f"`{p.rule_id}` ({p.failure_class})" if p.rule_id else ""
        lines.append(
            f"| {signature} | {p.jobs} | {p.decision} | {rule} | {covers} | {p.precision} |"
        )
    for p in proposals:
        if p.decision == "accepted" or not p.reason:
            continue
        lines.extend(
            ["", f"- `{p.signature_id}` {p.decision}: {p.reason.replace('|', '/')}"]
        )
    for p in accepted:
        rule = p.rule or {}
        pattern = str(rule.get("excerpt_pattern", "")).replace("`", "'")
        lines.extend(
            [
                "",
                f"### `{p.rule_id}`",
                "",
                f"- Class: `{rule.get('failure_class')}/{rule.get('subclass')}`",
                f"- Pattern: `{pattern}`",
                f"- Why: {rule.get('notes', '')}",
                f"- Evidence: {p.reason}",
            ]
        )
    return "\n".join(lines) + "\n"
