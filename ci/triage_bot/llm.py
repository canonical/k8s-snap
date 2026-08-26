#
# Copyright 2026 Canonical, Ltd.
#
"""Provider-agnostic chat-model factory."""

from __future__ import annotations

from langchain.chat_models import init_chat_model

from .runlog import silence_noisy_loggers

DEFAULT_MODEL = "google_genai:gemini-flash-latest"


def make_llm(model_spec: str = DEFAULT_MODEL, **kwargs):
    """Return a chat model for ``provider:model``."""
    silence_noisy_loggers()
    return init_chat_model(model_spec, temperature=0, **kwargs)
