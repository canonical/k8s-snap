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
_UNSAFE_COMMENT_CHARS = re.compile(r"[^A-Za-z0-9 ,.:;/_+=-]")


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


_FENCE_RUN_RE = re.compile(r"`{3,}")


def sanitize_fenced_text(text: str, limit: int = 1200) -> str:
    """Defang attacker-influenceable text destined for a ``` fenced block.

    A fence already blocks markdown/HTML interpretation, so the character
    allowlist :func:`sanitize_comment_text` needs is not: it would flatten
    multi-line commands to one line and strip quotes/parens a shell command
    needs to mean what it says. The one thing a fence does not defend
    against is content escaping it: a run of 3+ backticks landing at the
    start of a line closes the fence early, letting the rest of the text
    render as ordinary (interpreted) markdown. Capping every run to 2 makes
    that impossible regardless of where it falls, without touching anything
    else -- newlines, quoting, and command syntax survive intact.
    """
    return _FENCE_RUN_RE.sub("``", text).strip()[:limit]


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
        cand_tokens = set(
            re.findall(r"[a-z0-9]{4,}", (cand.get("title") or "").lower())
        )
        shared = title_tokens & cand_tokens
        union = title_tokens | cand_tokens
        if len(shared) >= 3 and len(shared) / len(union) >= 0.6:
            return cand
    return None


def dedup_query_tokens(title: str) -> list[str]:
    """The 4+ char tokens used to build a duplicate-search query."""
    return re.findall(r"[a-zA-Z0-9]{4,}", (title or "").lower())


# Caps each bug-report template field injected into the classify prompt. A
# reporter can paste an arbitrarily large log dump under "Reproduction
# Steps"; every other prompt in this module caps its user-controlled text
# (body, report, comment) and this one should not be the exception.
_FIELD_CHARS_CAP = 1500


def _classify_prompt(title: str, fields: dict, tarball: bool) -> str:
    parts = [
        "You are triaging a GitHub issue for the canonical/k8s-snap project.",
        "Classify it using ONLY these labels.",
        f"kind labels: {sorted(_KIND_LABELS)}",
        f"area labels: {sorted(_AREA_LABELS)}",
        "If required bug-report information is missing, list what is missing"
        " in `missing_info` (e.g. 'reproduction steps', 'inspection tarball').",
        "Only list missing_info if the issue cannot even be attempted to be reproduced. Do NOT ask for an inspection tarball if the reproduction steps are clear enough to try.",
        "",
        f"Title: {title}",
    ]
    for key in ("summary", "expected", "reproduction", "suggested_fix"):
        if fields.get(key):
            parts.append(f"{key}: {fields[key][:_FIELD_CHARS_CAP]}")
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
    # Defensive: the prompt asks for bare names, but a model can echo back a
    # GitHub-style prefix it has seen elsewhere in training data. Strip one if
    # present rather than silently dropping an otherwise-correct label.
    kind = [
        f"kind/{k}"
        for raw in result.kind_labels
        if (k := raw.strip().removeprefix("kind/").strip()) in _KIND_LABELS
    ]
    area = [
        f"area/{a}"
        for raw in result.area_labels
        if (a := raw.strip().removeprefix("area/").strip()) in _AREA_LABELS
    ]
    missing = [c for c in (sanitize_comment_text(m) for m in result.missing_info) if c]
    missing = missing[:5]
    return Classification(
        kind_labels=kind,
        area_labels=area,
        missing_info=missing,
        summary=result.summary.strip(),
    )


def decide_retriage(
    *,
    latest_comment: str,
    report: str,
    prior_request: str = "",
    model_spec: str = DEFAULT_MODEL,
) -> RetriageDecision:
    """Decide how a new comment on a parked issue affects its triage state.

    Conservative by construction: a bare acknowledgement ("thanks", "any
    update?") must not burn a cluster run, but genuinely new reproduction
    details or an attached inspection report should. Distinct from that is
    an explicit decline -- the reporter saying they cannot or will not
    provide what was asked for -- which should stop the bot from asking
    again rather than being read as silence.
    """

    prompt = (
        "You decide how a new comment on a previously-triaged k8s-snap "
        "issue affects whether to re-run automated triage. Choose exactly "
        "one outcome:\n"
        "- 'retriage': the comment provides materially new information (new "
        "reproduction steps, versions, logs, or an inspection report) that "
        "justifies re-running automated triage.\n"
        "- 'declined': the reporter explicitly says they cannot or will not "
        "provide what was asked for (no access to the system anymore, no "
        "time, don't know how, etc.) -- there is nothing further to wait "
        "for.\n"
        "- 'no_new_info': an acknowledgement, ping, or unrelated discussion; "
        "neither of the above.\n\n"
        f"What was asked for:\n{prior_request[:1000]}\n\n"
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
# Cap the inventory injected into a prompt: an uncapped list is unbounded
# prompt cost that only grows as docs are added. Shared by both doc-aware
# checks so one cannot silently outgrow the other.
_DOCS_PROMPT_CAP = 400


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
    root: Optional[Path] = None,
    model_spec: str = DEFAULT_MODEL,
) -> ExistingSupport:
    """Decide whether the issue asks for something k8s-snap already does.

    Reporters routinely file bugs against behaviour that has since been fixed,
    and feature requests for capabilities that already ship. Answering those
    with a documentation link is both faster and more useful than a cluster
    run. Any page the model cites is checked against the real inventory, so a
    plausible-looking but non-existent link can never reach the issue.
    """
    from langchain_core.tools import tool
    from langgraph.prebuilt import create_react_agent

    @tool
    def read_doc(path: str) -> str:
        """Read the full content of a documentation page. `path` must be exactly one of the available pages."""
        if not root:
            return "Error: documentation root not available"
        if path not in pages:
            return f"Error: path must be one of the available pages."
        try:
            return (root / DOCS_DIR / path).read_text(encoding="utf-8")
        except Exception as e:
            return f"Error reading {path}: {e}"

    prompt = (
        "You triage issues for the Canonical Kubernetes snap (k8s-snap).\n"
        "Decide whether the report asks for behaviour the project ALREADY "
        "provides: a feature that already ships, or a bug that is ALREADY FIXED.\n"
        "IMPORTANT: If the user is reporting that an existing, supported feature is BROKEN or failing, "
        "this is a new BUG, NOT already supported! Only set already_supported=true if the feature works "
        "and they just don't know how to use it, or if the exact bug is documented as fixed.\n"
        "Set already_supported only when you are confident; an unfamiliar "
        "request is not evidence that it is missing.\n"
        "If it is supported, explain briefly how to use it, and ALWAYS put the exact terminal commands "
        "a user should run in the `instructions` field (do not wrap in markdown ticks). "
        "Additionally, cite the documentation pages that cover it, chosen ONLY from the list below "
        "and copied exactly. If no page in the list covers it, just return no "
        "doc_paths.\n\n"
        "Use the read_doc tool to inspect the content of these pages to see if they answer the user's question.\n\n"
        f"Issue title: {title}\n"
        f"Issue body:\n{body[:4000]}\n\n"
        "Documentation pages available to read:\n" + "\n".join(f"- {p}" for p in pages[:_DOCS_PROMPT_CAP])
    )
    llm = make_llm(model_spec)
    agent = create_react_agent(llm, tools=[read_doc], prompt=prompt)
    result = agent.invoke({"messages": [("user", "Check existing support for this issue.")]})

    answer = next((str(m.content) for m in reversed(result["messages"]) if getattr(m, "type", "") == "ai" and m.content), "")
    if not answer:
        return ExistingSupport(already_supported=False)

    llm2 = make_llm(model_spec).with_structured_output(ExistingSupport)
    structured_result = llm2.invoke(
        f"An agent analyzed the issue and produced this report:\n\n{answer}\n\n"
        "Extract the structured result from this analysis. Use ONLY doc paths from this list:\n"
        + "\n".join(f"- {p}" for p in pages[:_DOCS_PROMPT_CAP])
    )

    known = set(pages)
    return ExistingSupport(
        already_supported=structured_result.already_supported,
        explanation=sanitize_comment_text(structured_result.explanation, limit=1200),
        doc_paths=[p for p in structured_result.doc_paths if p in known],
        instructions=sanitize_fenced_text(structured_result.instructions, limit=1200),
    )


def propose_enhancement(
    *,
    title: str,
    body: str,
    pages: list[str],
    root: Optional[Path] = None,
    model_spec: str = DEFAULT_MODEL,
) -> EnhancementProposal:
    """Surface implementation ideas and workarounds for a feature request.

    For a feature that does not yet exist, this is more useful than a generic
    "enhancement noted" comment: it tells the reporter whether they can do
    something today, and gives implementers concrete starting points.
    """
    from langchain_core.tools import tool
    from langgraph.prebuilt import create_react_agent

    @tool
    def read_doc(path: str) -> str:
        """Read the full content of a documentation page. `path` must be exactly one of the available pages."""
        if not root:
            return "Error: documentation root not available"
        if path not in pages:
            return f"Error: path must be one of the available pages."
        try:
            return (root / DOCS_DIR / path).read_text(encoding="utf-8")
        except Exception as e:
            return f"Error reading {path}: {e}"

    prompt = (
        "You are a senior engineer on the Canonical Kubernetes snap project.\n"
        "A user has opened a FEATURE REQUEST (not a bug). Your job is to:\n"
        "1. Check whether any workaround already exists that satisfies the "
        "request TODAY, even partially (e.g. a flag, a CLI command, an "
        "annotation, a service stop command). If a documentation page below "
        "covers it, cite it exactly as listed; otherwise leave doc_paths empty "
        "rather than guessing a URL.\n"
        "The workaround_instructions field MUST contain ONLY the raw terminal "
        "command(s) or code. Do NOT wrap it in markdown ticks (```), provide "
        "NO prose, and NO explanations, as it will be wrapped in a code block "
        "automatically.\n"
        "2. Propose 1-3 concrete, minimal implementation ideas ranked by "
        "effort. Each idea must name the specific file or component to change "
        "and include an example command or code snippet.\n"
        "Be specific: cite real k8s-snap CLI flags (`k8s set`, "
        "`k8s bootstrap --file`), snap service names (`k8s.kubelet`), "
        "kubelet flags (`--register-node=false`), or annotation keys. "
        "Do not invent flags that do not exist.\n"
        "Use the read_doc tool to inspect the content of these pages if needed to find workarounds.\n\n"
        f"Issue title: {title}\n"
        f"Issue body:\n{body[:3000]}\n\n"
        "Documentation pages available for context (cite only real pages):\n"
        + "\n".join(f"- {p}" for p in pages[:_DOCS_PROMPT_CAP])
    )
    llm = make_llm(model_spec)
    agent = create_react_agent(llm, tools=[read_doc], prompt=prompt)
    result = agent.invoke({"messages": [("user", "Propose an enhancement and find workarounds for this issue.")]})

    answer = next((str(m.content) for m in reversed(result["messages"]) if getattr(m, "type", "") == "ai" and m.content), "")
    if not answer:
        return EnhancementProposal()

    llm2 = make_llm(model_spec).with_structured_output(EnhancementProposal)
    structured_result = llm2.invoke(
        f"An agent analyzed the issue and produced this report:\n\n{answer}\n\n"
        "Extract the structured result from this analysis. Use ONLY doc paths from this list:\n"
        + "\n".join(f"- {p}" for p in pages[:_DOCS_PROMPT_CAP])
    )

    known = set(pages)
    return EnhancementProposal(
        workaround_exists=structured_result.workaround_exists,
        workaround_instructions=sanitize_fenced_text(
            structured_result.workaround_instructions, limit=1200
        ),
        workaround_doc_paths=[p for p in structured_result.workaround_doc_paths if p in known],
        ideas=[
            ImplementationIdea(
                title=sanitize_comment_text(idea.title, limit=120),
                description=sanitize_comment_text(idea.description, limit=800),
                example=sanitize_fenced_text(idea.example, limit=600),
                effort=(
                    idea.effort
                    if idea.effort in ("workaround", "small", "medium", "large")
                    else "medium"
                ),
            )
            for idea in structured_result.ideas[:4]
        ],
    )
