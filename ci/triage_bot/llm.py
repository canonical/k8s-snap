#
# Copyright 2026 Canonical, Ltd.
#
"""Provider-agnostic chat-model factory.

One ``provider:model`` string selects everything. Switching provider is pure
config: ``google_genai:gemini-flash-latest`` -> ``anthropic:claude-...`` ->
``openai:gpt-...``. Each provider reads its own API key from the environment
(``GOOGLE_API_KEY`` for Gemini); secrets never come from CLI flags.
"""

from __future__ import annotations

DEFAULT_MODEL = "google_genai:gemini-flash-latest"


def make_llm(model_spec: str = DEFAULT_MODEL, **kwargs):
    """Return a chat model for ``provider:model``."""
    # Imported lazily so the base CLI can read DEFAULT_MODEL without pulling
    # the LangChain stack.
    from langchain.chat_models import init_chat_model

    from .runlog import silence_noisy_loggers

    # Every LLM path goes through here, so this is the one place that can keep
    # provider chatter out of the run log for callers that never run a skill
    # (the classifier, retriage, fix verification).
    silence_noisy_loggers()
    return init_chat_model(model_spec, temperature=0, **kwargs)
