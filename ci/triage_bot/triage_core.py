#
# Copyright 2026 Canonical, Ltd.
#
"""Deterministic triage labeling and issue parsing.

Action-owned counterpart to the project-owned skills: everything here is a pure
function of the issue text (or a single structured LLM call), with no cluster
access. It is the analog of the reference bot's ``selectTriageLabels`` -- it
decides the ``kind/`` and ``area/`` labels, detects a missing inspection
tarball, and flags likely duplicates, all offline-testable.

The regexes and the sanitiser are carried over verbatim from the original graph
nodes; only the plumbing changed (plain arguments instead of a graph state).
"""

from __future__ import annotations

import re
from pathlib import Path
from typing import Optional

from .llm import DEFAULT_MODEL, make_llm
from .schema import (
    Classification,
    EnhancementProposal,
    ExistingSupport,
    FixVerification,
    ImplementationIdea,
    RetriageDecision,
)

# GitHub renders issue-form fields as "### <label>\n\n<value>". These map the
# bug_report.yaml labels to the field keys the classifier prompt cares about.
_TEMPLATE_FIELDS = {
    "Summary": "summary",
    "What Should Happen Instead?": "expected",
    "Reproduction Steps": "reproduction",
    "System information": "system_info",
    "Can you suggest a fix?": "suggested_fix",
}

_HEADER_RE = re.compile(r"^#{1,6}\s+(?P<label>.+?)\s*$", re.MULTILINE)

# Detect an inspection report by its tarball extension, in either the link
# text or the URL. A non-tarball attachment (e.g. .zip) does not count.
_TARBALL_RE = re.compile(r"\.(?:tar\.gz|tgz|tar)\b", re.IGNORECASE)

_KIND_LABELS = {"bug", "enhancement", "documentation", "question"}
_AREA_LABELS = {
    "network",
    "storage",
    "dns",
    "ingress",
    "observability",
    "security",
    "cluster-lifecycle",
    "api",
    "snap",
    "performance",
}

# missing_info items are LLM free text posted verbatim in a bot comment, so
# they are attacker-influenceable via prompt injection. Restrict to a safe
# character allowlist: this structurally defuses markdown links ``[](url)``,
# inline HTML ``<...>``, @mentions, and scheme/``www.`` autolinks in one rule,
# rather than chasing an ever-growing blocklist.
_UNSAFE_COMMENT_CHARS = re.compile(r"[^A-Za-z0-9 ,.:;/_+-]")


def sanitize_comment_text(text: str, limit: int = 80) -> str:
    """Defang attacker-influenceable text before it enters a bot comment."""
    cleaned = _UNSAFE_COMMENT_CHARS.sub(" ", text)
    # The allowlist keeps ':' and '/', so a bare "https://host" would survive
    # and GitHub would autolink it. Break the two autolink triggers GFM uses:
    # the "scheme://" separator and a "www." host. No \b on www.: '_' is a
    # regex word char and the one autolink delimiter the allowlist keeps, so
    # "_www.evil.co" must defang too.
    cleaned = re.sub(r"(?i)://", " ", cleaned)
    cleaned = re.sub(r"(?i)www\.", "www ", cleaned)
    return re.sub(r"\s+", " ", cleaned).strip()[:limit]


def parse_template(body: str) -> dict:
    """Split an issue body into fields keyed by the bug-report template.

    Falls back gracefully: unknown sections are ignored, and a body with no
    recognised headers yields an empty mapping (classification still runs on
    the raw summary/title).
    """
    matches = list(_HEADER_RE.finditer(body or ""))
    fields: dict[str, str] = {}
    for i, m in enumerate(matches):
        label = m.group("label").strip()
        key = _TEMPLATE_FIELDS.get(label)
        if not key:
            continue
        start = m.end()
        end = matches[i + 1].start() if i + 1 < len(matches) else len(body)
        value = body[start:end].strip()
        # Issue forms render an unfilled field as "_No response_".
        if value and value != "_No response_":
            fields[key] = value
    return fields


def has_tarball(body: str, comments: list[str]) -> bool:
    haystack = "\n".join([body or "", *(comments or [])])
    return bool(_TARBALL_RE.search(haystack))


def find_duplicate(title: str, candidates: list[dict]) -> Optional[dict]:
    """Return the candidate that is a strong duplicate of ``title``, if any.

    Uses Jaccard overlap over 4+ char title tokens and requires a floor of
    shared tokens so short, generic titles do not false-positive.
    """
    title_tokens = set(re.findall(r"[a-z0-9]{4,}", (title or "").lower()))
    if len(title_tokens) < 4:
        return None
    for cand in candidates:
        cand_tokens = set(re.findall(r"[a-z0-9]{4,}", cand.get("title", "").lower()))
        shared = title_tokens & cand_tokens
        union = title_tokens | cand_tokens
        if len(shared) >= 3 and len(shared) / len(union) >= 0.6:
            return cand
    return None


def dedup_query_tokens(title: str) -> list[str]:
    """The 4+ char tokens used to build a duplicate-search query."""
    return re.findall(r"[a-zA-Z0-9]{4,}", (title or "").lower())


def _classify_prompt(title: str, fields: dict, tarball: bool) -> str:
    parts = [
        "You are triaging a GitHub issue for the canonical/k8s-snap project.",
        "Classify it using ONLY these labels.",
        f"kind labels: {sorted(_KIND_LABELS)}",
        f"area labels: {sorted(_AREA_LABELS)}",
        "If required bug-report information is missing, list what is missing"
        " in `missing_info` (e.g. 'reproduction steps', 'inspection tarball').",
        "",
        f"Title: {title}",
    ]
    for key in ("summary", "expected", "reproduction", "suggested_fix"):
        if fields.get(key):
            parts.append(f"{key}: {fields[key]}")
    parts.append(f"inspection tarball attached: {tarball}")
    return "\n".join(parts)


def classify(
    *,
    title: str,
    fields: dict,
    tarball: bool,
    model_spec: str = DEFAULT_MODEL,
) -> Classification:
    """One structured LLM call producing typed ``kind``/``area``/``missing``."""
    llm = make_llm(model_spec).with_structured_output(Classification)
    result: Classification = llm.invoke(_classify_prompt(title, fields, tarball))
    kind = [f"kind/{k}" for k in result.kind_labels if k in _KIND_LABELS]
    area = [f"area/{a}" for a in result.area_labels if a in _AREA_LABELS]
    missing = [c for c in (sanitize_comment_text(m) for m in result.missing_info) if c][
        :5
    ]
    # A bug with no tarball is always missing the inspection report.
    if "kind/bug" in kind and not tarball:
        if not any("tarball" in m or "inspection" in m for m in missing):
            missing.append("inspection tarball")
    return Classification(
        kind_labels=kind,
        area_labels=area,
        missing_info=missing,
        summary=result.summary.strip(),
    )


def decide_retriage(
    *, latest_comment: str, report: str, model_spec: str = DEFAULT_MODEL
) -> RetriageDecision:
    """Decide whether a new comment warrants re-running triage.

    Conservative by construction: a bare acknowledgement ("thanks", "any
    update?") must not burn a cluster run, but genuinely new reproduction
    details or an attached inspection report should.
    """

    prompt = (
        "You decide whether a new comment on a previously-triaged k8s-snap "
        "issue provides materially new information (new reproduction steps, "
        "versions, logs, or an inspection report) that justifies re-running "
        "automated triage. Acknowledgements, pings, or unrelated discussion "
        "do NOT.\n\n"
        f"Prior triage report:\n{report[:4000]}\n\n"
        f"New comment:\n{latest_comment[:2000]}"
    )
    llm = make_llm(model_spec).with_structured_output(RetriageDecision)
    return llm.invoke(prompt)


def verify_fix(
    *, latest_comment: str, report: str, model_spec: str = DEFAULT_MODEL
) -> FixVerification:
    """Classify a maintainer comment as confirming or rejecting the fix."""

    prompt = (
        "A draft fix was proposed for a k8s-snap issue and is awaiting "
        "confirmation. Classify the new comment as 'confirmed' (the fix works "
        "/ is approved), 'rejected' (the fix is wrong or the problem "
        "persists), or 'inconclusive' (neither).\n\n"
        f"Fix context:\n{report[:4000]}\n\n"
        f"New comment:\n{latest_comment[:2000]}"
    )
    llm = make_llm(model_spec).with_structured_output(FixVerification)
    return llm.invoke(prompt)


# Documentation source tree and the site it publishes to. The mapping is
# mechanical (``snap/howto/contribute.md`` -> ``.../snap/howto/contribute``),
# so links are derived here rather than invented by a model.
DOCS_DIR = "docs/canonicalk8s"
DOCS_BASE_URL = "https://documentation.ubuntu.com/canonical-kubernetes/latest"
_DOCS_SKIP = ("_build", "_parts", "_dev", "_templates")


def doc_inventory(root: Path) -> list[str]:
    """Every publishable doc page, as paths relative to :data:`DOCS_DIR`."""
    base = root / DOCS_DIR
    if not base.is_dir():
        return []
    return sorted(
        str(p.relative_to(base))
        for p in base.rglob("*.md")
        if not any(part in _DOCS_SKIP for part in p.relative_to(base).parts)
    )


def doc_url(page: str) -> str:
    """The published URL for a page path from :func:`doc_inventory`."""
    slug = page[: -len(".md")] if page.endswith(".md") else page
    if slug.endswith("/index"):
        slug = slug[: -len("/index")]
    return f"{DOCS_BASE_URL}/{slug}"


def check_existing_support(
    *,
    title: str,
    body: str,
    pages: list[str],
    model_spec: str = DEFAULT_MODEL,
) -> ExistingSupport:
    """Decide whether the issue asks for something k8s-snap already does.

    Reporters routinely file bugs against behaviour that has since been fixed,
    and feature requests for capabilities that already ship. Answering those
    with a documentation link is both faster and more useful than a cluster
    run. Any page the model cites is checked against the real inventory, so a
    plausible-looking but non-existent link can never reach the issue.
    """
    prompt = (
        "You triage issues for the Canonical Kubernetes snap (k8s-snap).\n"
        "Decide whether the report asks for behaviour the project ALREADY "
        "provides: a feature that already ships, or a bug already fixed.\n"
        "Set already_supported only when you are confident; an unfamiliar "
        "request is not evidence that it is missing.\n"
        "If it is supported, explain briefly how to use it, and cite the "
        "documentation pages that cover it, chosen ONLY from the list below "
        "and copied exactly. If no page in the list covers it, return no "
        "doc_paths and put the exact commands a user should run in "
        "instructions.\n\n"
        f"Issue title: {title}\n"
        f"Issue body:\n{body[:4000]}\n\n"
        "Documentation pages:\n" + "\n".join(pages)
    )
    llm = make_llm(model_spec).with_structured_output(ExistingSupport)
    result: ExistingSupport = llm.invoke(prompt)
    known = set(pages)
    return ExistingSupport(
        already_supported=result.already_supported,
        explanation=sanitize_comment_text(result.explanation, limit=1200),
        doc_paths=[p for p in result.doc_paths if p in known],
        instructions=sanitize_comment_text(result.instructions, limit=1200),
    )


def propose_enhancement(
    *,
    title: str,
    body: str,
    pages: list[str],
    model_spec: str = DEFAULT_MODEL,
) -> EnhancementProposal:
    """Surface implementation ideas and workarounds for a feature request.

    For a feature that does not yet exist, this is more useful than a generic
    "enhancement noted" comment: it tells the reporter whether they can do
    something today, and gives implementers concrete starting points.
    """
    prompt = (
        "You are a senior engineer on the Canonical Kubernetes snap project.\n"
        "A user has opened a FEATURE REQUEST (not a bug). Your job is to:\n"
        "1. Check whether any workaround already exists that satisfies the "
        "request TODAY, even partially (e.g. a flag, a CLI command, an "
        "annotation, a service stop command). If a documentation page below "
        "covers it, cite it exactly as listed; otherwise leave doc_paths empty "
        "rather than guessing a URL.\n"
        "2. Propose 1-3 concrete, minimal implementation ideas ranked by "
        "effort. Each idea must name the specific file or component to change "
        "and include an example command or code snippet.\n"
        "Be specific: cite real k8s-snap CLI flags (`k8s set`, "
        "`k8s bootstrap --file`), snap service names (`k8s.kubelet`), "
        "kubelet flags (`--register-node=false`), or annotation keys. "
        "Do not invent flags that do not exist.\n"
        "Documentation pages available for context (cite only real pages):\n"
        + "\n".join(pages[:60])
        + f"\n\nIssue title: {title}\nIssue body:\n{body[:3000]}"
    )
    llm = make_llm(model_spec).with_structured_output(EnhancementProposal)
    result: EnhancementProposal = llm.invoke(prompt)
    known = set(pages)
    return EnhancementProposal(
        workaround_exists=result.workaround_exists,
        workaround_instructions=sanitize_comment_text(
            result.workaround_instructions, limit=1200
        ),
        workaround_doc_paths=[p for p in result.workaround_doc_paths if p in known],
        ideas=[
            ImplementationIdea(
                title=sanitize_comment_text(idea.title, limit=120),
                description=sanitize_comment_text(idea.description, limit=800),
                example=sanitize_comment_text(idea.example, limit=600),
                effort=(
                    idea.effort
                    if idea.effort in ("workaround", "small", "medium", "large")
                    else "medium"
                ),
            )
            for idea in result.ideas[:4]
        ],
    )
