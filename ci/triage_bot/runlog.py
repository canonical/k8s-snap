#
# Copyright 2026 Canonical, Ltd.
#
"""Structured, redacted per-tool-call logging for the triage skills.

A single ``RunLogger`` callback handler is the attach point, passed to the
skill agent via ``config={"callbacks": [...]}``. It emits one redacted record
per LangGraph node boundary, per LLM call, and per tool invocation (the shell
commands the agent runs and what they returned). This is the *granular*
trace -- every single tool call inside a stage -- so it logs at DEBUG; the
high-level narrative (``[reproduce] reproducible=True``, and so on) comes
from ``pipeline.py`` at INFO, which is the default level. Any LLM or tool
failure is the one exception: it logs at ERROR regardless, so it survives a
default-level configuration and a failed run is diagnosable from the job log
alone, without reproducing it. Records also go, optionally, to a
newline-delimited JSON sink (machine). Secrets are value-substituted out of
every record before it is written.
"""

from __future__ import annotations

import json
import logging
import time
from pathlib import Path
from typing import Any, Callable, Optional

from langchain_core.callbacks import BaseCallbackHandler

log = logging.getLogger("triage_bot")

_NOISY_LOGGERS = (
    "httpx",
    "httpcore",
    "langchain_google_genai._function_utils",
    "google_genai.models",
)


def silence_noisy_loggers() -> None:
    """Raise third-party loggers that emit benign noise to WARNING."""
    for name in _NOISY_LOGGERS:
        logging.getLogger(name).setLevel(logging.WARNING)


class CredentialRedactor:
    """Replace known secret values with ``[REDACTED:KEY]`` markers."""

    def __init__(self, secrets: dict[str, str]):
        # Only redact non-trivial values to avoid masking short/empty env vars.
        self._replacements = [
            (val, f"[REDACTED:{key}]")
            for key, val in secrets.items()
            if val and len(val) >= 8
        ]

    def redact(self, text: str) -> str:
        for secret, marker in self._replacements:
            text = text.replace(secret, marker)
        return text

    def redact_record(self, obj):
        """Redact raw string values before serialisation (escape-safe)."""
        if isinstance(obj, str):
            return self.redact(obj)
        if isinstance(obj, dict):
            return {k: self.redact_record(v) for k, v in obj.items()}
        if isinstance(obj, list):
            return [self.redact_record(v) for v in obj]
        return obj


def _usage(response: Any) -> dict[str, int]:
    """Best-effort token usage extraction from an LLM response."""
    try:
        for gen in response.generations:
            for g in gen:
                msg = getattr(g, "message", None)
                usage = getattr(msg, "usage_metadata", None)
                if usage:
                    return {
                        "input_tokens": int(usage.get("input_tokens", 0)),
                        "output_tokens": int(usage.get("output_tokens", 0)),
                    }
    except (AttributeError, TypeError):
        pass
    return {"input_tokens": 0, "output_tokens": 0}


# Tool payloads (a shell command's combined output) run to thousands of
# characters. The sink keeps them whole; the human log line is clipped so a
# run stays readable.
_DETAIL_CHARS = 400


def _clip(text: Any) -> str:
    collapsed = " ".join(str(text).split())
    if len(collapsed) <= _DETAIL_CHARS:
        return collapsed
    return f"{collapsed[:_DETAIL_CHARS]}... (+{len(collapsed) - _DETAIL_CHARS} chars)"


class RunLogger(BaseCallbackHandler):
    """Emit one structured record per node, LLM call, tool call, and failure."""

    def __init__(
        self,
        run_id: str,
        redactor: CredentialRedactor,
        sink: Optional[Callable[[dict], None]] = None,
    ):
        self.run_id = run_id
        self._redactor = redactor
        self._sink = sink

    def _emit(self, record: dict) -> None:
        record = self._redactor.redact_record({"run_id": self.run_id, **record})
        event = str(record.get("event", ""))
        detail = record.get("detail")
        # Routine records are DEBUG (the granular per-tool trace); failures are
        # ERROR so they surface even at the default INFO level.
        log.log(
            logging.ERROR if event.endswith("error") else logging.DEBUG,
            "[%s] %s%s",
            record.get("node", "-"),
            event,
            f" {_clip(detail)}" if detail else "",
        )
        if self._sink is not None:
            self._sink(record)

    def on_chain_start(self, serialized, inputs, **kwargs) -> None:
        node = (kwargs.get("metadata") or {}).get("langgraph_node")
        # LangGraph fires on_chain_start for the node's own Pregel step and
        # again for any inner Runnable (e.g. a conditional-edge function) that
        # inherits the node metadata. At the true node boundary the run name
        # equals the node name, so gate on that to log exactly one record per
        # node entry (public metadata, not an internal tag).
        if node and kwargs.get("name") == node:
            self._emit({"node": node, "event": "node_start", "ts": time.time()})

    def on_llm_end(self, response, **kwargs) -> None:
        self._emit({"event": "llm_end", "tokens": _usage(response), "ts": time.time()})

    def on_tool_start(self, serialized, input_str, **kwargs) -> None:
        self._emit(
            {
                "event": "tool_start",
                "tool": (serialized or {}).get("name", "shell"),
                "detail": str(input_str),
                "ts": time.time(),
            }
        )

    def on_tool_end(self, output, **kwargs) -> None:
        # LangChain hands back a ToolMessage; the shell output it wraps is the
        # part worth logging, not the envelope.
        detail = getattr(output, "content", output)
        self._emit({"event": "tool_end", "detail": str(detail), "ts": time.time()})

    def on_tool_error(self, error, **kwargs) -> None:
        self._emit({"event": "tool_error", "detail": repr(error), "ts": time.time()})

    def on_llm_error(self, error, **kwargs) -> None:
        self._emit({"event": "llm_error", "detail": repr(error), "ts": time.time()})


def jsonl_sink(path: str) -> Callable[[dict], None]:
    """Return a sink that appends each record as one JSON line to ``path``.

    Creates the parent directory up front: a path under a fresh per-run or
    per-issue directory (e.g. ``.triage/runlogs/issue-42.jsonl``) is the
    common case, and the first write must not fail with ``FileNotFoundError``
    just because nothing has created that directory yet.
    """
    Path(path).parent.mkdir(parents=True, exist_ok=True)

    def _write(record: dict) -> None:
        with open(path, "a", encoding="utf-8") as fh:
            fh.write(json.dumps(record, default=str) + "\n")

    return _write
