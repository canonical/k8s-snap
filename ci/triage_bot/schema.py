#
# Copyright 2026 Canonical, Ltd.
#
"""Typed result of the classify node.

Kept intentionally flat: some JSON-schema keys (e.g. ``additionalProperties``)
are dropped by the Gemini adapter, so nested/complex schemas are avoided.
"""

from __future__ import annotations

from typing import Literal, Optional

from pydantic import BaseModel, Field


class Classification(BaseModel):
    kind_labels: list[str] = Field(default_factory=list)
    area_labels: list[str] = Field(default_factory=list)
    missing_info: list[str] = Field(default_factory=list)
    summary: str = ""


class ExistingSupport(BaseModel):
    """Whether k8s-snap already does what the issue asks for."""

    already_supported: bool = False
    explanation: str = ""
    # Repo-relative doc pages backing the explanation. Validated against the
    # real inventory before use, so an invented page cannot become a link.
    doc_paths: list[str] = Field(default_factory=list)
    # Concrete commands to use the feature, needed when no page documents it.
    instructions: str = ""


class ImplementationIdea(BaseModel):
    """One concrete implementation approach for a feature request."""

    title: str = ""
    description: str = ""
    # Concrete commands or code snippet the reporter can try today.
    example: str = ""
    effort: str = ""  # "workaround" | "small" | "medium" | "large"


class EnhancementProposal(BaseModel):
    """Implementation ideas the bot surfaced for a feature request."""

    # Whether a workaround already exists that the reporter can use right now.
    workaround_exists: bool = False
    workaround_instructions: str = ""
    # Repo-relative doc pages backing the workaround, validated against the
    # real inventory before use (same contract as ExistingSupport.doc_paths).
    workaround_doc_paths: list[str] = Field(default_factory=list)
    ideas: list[ImplementationIdea] = Field(default_factory=list)


# --- pipeline stage results (skills return these as structured output) ---

SkipReason = Literal[
    "not-actionable",
    "missing-details",
    "unsupported-version",
    "host-specific",
    "unsupported-runtime",
    "maintainer-override",
]
Verdict = Literal["bug", "intended-behavior", "unclear"]
Confidence = Literal["high", "medium", "low"]


class ReproduceResult(BaseModel):
    reproducible: bool = False
    skipped: bool = False
    skipped_reason: Optional[SkipReason] = None
    # The command run and the failure observed. Required when reproducible, so
    # the claim rests on something the agent saw rather than on the report text.
    evidence: str = ""


class ReproducerResult(BaseModel):
    """The end-to-end test that captures the bug, proven red before any fix."""

    test_path: str = ""
    # The ``-k`` selector that runs just this test.
    test_selector: str = ""
    fails_before_fix: bool = False
    failure_output: str = ""


class DiagnoseResult(BaseModel):
    confidence: Optional[Confidence] = None


class VerifyResult(BaseModel):
    verdict: Verdict = "unclear"
    confidence: Confidence = "low"


class FixResult(BaseModel):
    fixed: bool = False
    commit_message: Optional[str] = None


class TriageResult(BaseModel):
    """Aggregated outcome of the reproduce -> verify -> reproducer -> fix run."""

    completed_stage: Literal["reproduce", "verify", "reproducer", "fix"] = "reproduce"
    reproducible: bool = False
    skipped: bool = False
    skipped_reason: Optional[SkipReason] = None
    verdict: Optional[Verdict] = None
    diagnosis_confidence: Optional[Confidence] = None
    fixed: bool = False
    commit_message: Optional[str] = None
    pr_url: Optional[str] = None
    test_path: Optional[str] = None


class RetriageDecision(BaseModel):
    outcome: Literal["retriage", "declined", "no_new_info"] = "no_new_info"
    reason: str = ""


class FixVerification(BaseModel):
    status: Literal["confirmed", "rejected", "inconclusive"] = "inconclusive"
    reasoning: str = ""
