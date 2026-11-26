#
# Copyright 2026 Canonical, Ltd.
#
"""Minimal chat-completions client for the rule proposer.

Speaks the OpenAI-compatible ``/chat/completions`` API that both GitHub
Models and OpenRouter expose, so the proposer needs no SDK. In GitHub Actions
the default provider authenticates with the workflow's own ``GITHUB_TOKEN``
(``permissions: models: read``), so no extra secret is involved.
"""

from __future__ import annotations

import dataclasses
import json
import logging
import os
import re
import time
from typing import Any, Dict, List, Mapping, Optional

LOG = logging.getLogger(__name__)

MAX_ATTEMPTS = 4
MAX_RETRY_AFTER_S = 60
REQUEST_TIMEOUT_S = 180


class LLMError(RuntimeError):
    """Raised when the model endpoint fails or returns something unusable."""


@dataclasses.dataclass(frozen=True)
class Provider:
    name: str
    api_base: str
    default_model: str
    token_env: tuple[str, ...]


PROVIDERS: Dict[str, Provider] = {
    "github": Provider(
        "github",
        "https://models.github.ai/inference",
        "openai/gpt-4.1",
        ("GITHUB_TOKEN", "GH_TOKEN"),
    ),
    "openrouter": Provider(
        "openrouter",
        "https://openrouter.ai/api/v1",
        "google/gemini-3.8-flash",
        ("OPENROUTER_API_KEY",),
    ),
}


def resolve_provider(
    name: str = "auto", environ: Optional[Mapping[str, str]] = None
) -> Provider:
    """Return the named provider, or for ``auto`` the first one with a token."""

    env = os.environ if environ is None else environ
    if name != "auto":
        try:
            return PROVIDERS[name]
        except KeyError:
            raise LLMError(
                f"unknown provider {name!r}; expected one of {sorted(PROVIDERS)}"
            ) from None
    for provider in (PROVIDERS["openrouter"], PROVIDERS["github"]):
        if any((env.get(var) or "").strip() for var in provider.token_env):
            return provider
    raise LLMError(
        "no model credentials: set GITHUB_TOKEN (GitHub Models) or "
        "OPENROUTER_API_KEY"
    )


class ChatClient:
    """Send chat messages to one model and return its reply text."""

    def __init__(
        self,
        provider: Provider,
        *,
        model: Optional[str] = None,
        token: Optional[str] = None,
        session: Any = None,
        environ: Optional[Mapping[str, str]] = None,
    ) -> None:
        env = os.environ if environ is None else environ
        self.provider = provider
        self.model = model or provider.default_model
        self.token = token or next(
            (env[var] for var in provider.token_env if (env.get(var) or "").strip()),
            None,
        )
        if not self.token:
            raise LLMError(
                f"{provider.name}: none of {list(provider.token_env)} is set"
            )
        if session is None:
            # Imported lazily, like metrics.gh: k8s-ci.py loads every command
            # module at startup, sometimes under a bare python3.
            import requests

            session = requests.Session()
        self.session = session
        self.calls_made = 0

    def complete(
        self, messages: List[Dict[str, str]], *, max_tokens: int = 2000
    ) -> str:
        """Return the reply to ``messages``, retrying rate limits and 5xx."""

        url = f"{self.provider.api_base}/chat/completions"
        payload = {
            "model": self.model,
            "messages": messages,
            "temperature": 0,
            "max_tokens": max_tokens,
            "response_format": {"type": "json_object"},
        }
        headers = {
            "Authorization": f"Bearer {self.token}",
            "Content-Type": "application/json",
            "Accept": "application/json",
        }
        last_error = "no attempt made"
        for attempt in range(MAX_ATTEMPTS):
            response = self.session.post(
                url, headers=headers, json=payload, timeout=REQUEST_TIMEOUT_S
            )
            self.calls_made += 1
            if response.status_code == 429 or response.status_code >= 500:
                wait = _retry_after(response, attempt)
                last_error = f"HTTP {response.status_code}"
                LOG.warning("%s from %s; retrying in %ss", last_error, url, wait)
                time.sleep(wait)
                continue
            if response.status_code >= 400:
                raise LLMError(
                    f"HTTP {response.status_code} from {url}: {response.text[:300]}"
                )
            try:
                body = response.json()
                return str(body["choices"][0]["message"]["content"] or "")
            except (ValueError, KeyError, IndexError, TypeError):
                content_type = response.headers.get("content-type", "")
                raise LLMError(
                    f"{url} returned no chat completion ({content_type}): "
                    f"{response.text[:200]!r}"
                ) from None
        raise LLMError(f"{url}: giving up after {MAX_ATTEMPTS} attempts ({last_error})")


def _retry_after(response: Any, attempt: int) -> int:
    header = response.headers.get("retry-after", "")
    if header.isdigit():
        return min(int(header), MAX_RETRY_AFTER_S)
    return min(2 ** (attempt + 1), MAX_RETRY_AFTER_S)


_FENCE = re.compile(r"^```(?:json)?\s*|\s*```$", re.MULTILINE)


def parse_json_reply(text: str) -> Dict[str, Any]:
    """Return the first JSON object in a model reply, tolerating code fences."""

    cleaned = _FENCE.sub("", text.strip())
    start = cleaned.find("{")
    if start < 0:
        raise LLMError(f"reply contains no JSON object: {text[:200]!r}")
    try:
        value, _end = json.JSONDecoder().raw_decode(cleaned, start)
    except json.JSONDecodeError as exc:
        raise LLMError(f"reply is not valid JSON: {exc}") from None
    if not isinstance(value, dict):
        raise LLMError("reply JSON is not an object")
    return value
