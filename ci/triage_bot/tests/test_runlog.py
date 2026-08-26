#
# Copyright 2026 Canonical, Ltd.
#
"""Guards on the run log.

These pin the post-mortem contract: a failed run must be diagnosable from the
job log alone. The shell commands the agent ran, what they returned, and any
LLM/tool failure all have to be recorded -- with secrets scrubbed first. They
also pin the level split: routine per-tool-call records are DEBUG-only (the
high-level ``[stage]`` narrative that appears at the default INFO level lives
in ``pipeline.py``, not here), while failures always reach INFO and above.
"""

from __future__ import annotations

import json
import logging

from triage_bot.runlog import CredentialRedactor, RunLogger, jsonl_sink


def _logger(sink=None, secrets=None):
    return RunLogger("run-1", CredentialRedactor(secrets or {}), sink=sink)


def test_tool_calls_are_recorded(caplog):
    records = []
    rl = _logger(sink=records.append)

    with caplog.at_level(logging.DEBUG, logger="triage_bot"):
        rl.on_tool_start({"name": "shell"}, "k8s status --wait-ready")
        rl.on_tool_end("exit=0\ncluster status: ready")

    assert [r["event"] for r in records] == ["tool_start", "tool_end"]
    assert records[0]["detail"] == "k8s status --wait-ready"
    assert "k8s status --wait-ready" in caplog.text
    assert "cluster status: ready" in caplog.text


def test_tool_start_and_end_events_are_at_debug_level(caplog):
    rl = _logger()

    with caplog.at_level(logging.INFO, logger="triage_bot"):
        rl.on_tool_start({"name": "shell"}, "some noisy diagnostic command")
        rl.on_tool_end("noisy output nobody asked to see by default")

    assert "some noisy diagnostic command" not in caplog.text
    assert "noisy output nobody asked to see" not in caplog.text

    with caplog.at_level(logging.DEBUG, logger="triage_bot"):
        rl.on_tool_start({"name": "shell"}, "some noisy diagnostic command")
        rl.on_tool_end("noisy output nobody asked to see by default")

    assert "some noisy diagnostic command" in caplog.text
    assert "noisy output nobody asked to see" in caplog.text


def test_failures_are_recorded_at_error_level(caplog):
    rl = _logger()

    with caplog.at_level(logging.INFO, logger="triage_bot"):
        rl.on_llm_error(RuntimeError("400 INVALID_ARGUMENT: model turn"))
        rl.on_tool_error(TimeoutError("timed out"))

    assert {r.levelname for r in caplog.records} == {"ERROR"}
    assert "400 INVALID_ARGUMENT" in caplog.text
    assert "timed out" in caplog.text


def test_secrets_reach_neither_the_log_nor_the_sink(caplog):
    records = []
    rl = _logger(sink=records.append, secrets={"GH_TOKEN": "ghp_supersecret"})

    with caplog.at_level(logging.DEBUG, logger="triage_bot"):
        rl.on_tool_end("pushed to https://x-access-token:ghp_supersecret@github.com")

    assert "ghp_supersecret" not in caplog.text
    assert "ghp_supersecret" not in records[0]["detail"]
    assert "[REDACTED:GH_TOKEN]" in records[0]["detail"]


def test_long_output_is_clipped_in_the_log_but_kept_whole_in_the_sink(caplog):
    records = []
    rl = _logger(sink=records.append)

    with caplog.at_level(logging.DEBUG, logger="triage_bot"):
        rl.on_tool_end("x" * 5000)

    assert len(records[0]["detail"]) == 5000
    assert len(caplog.text) < 1000
    assert "chars)" in caplog.text


def test_jsonl_sink_creates_a_missing_parent_directory(tmp_path):
    # A fresh per-issue directory (.triage/runlogs/issue-42.jsonl) is the
    # common case; the first write must not raise FileNotFoundError just
    # because nothing created that directory yet.
    path = tmp_path / "runlogs" / "issue-42.jsonl"

    sink = jsonl_sink(str(path))
    sink({"event": "tool_start", "detail": "k8s status"})

    lines = path.read_text(encoding="utf-8").splitlines()
    assert json.loads(lines[0]) == {"event": "tool_start", "detail": "k8s status"}


def test_jsonl_sink_appends_one_line_per_record(tmp_path):
    path = tmp_path / "run.jsonl"
    sink = jsonl_sink(str(path))

    sink({"event": "tool_start"})
    sink({"event": "tool_end"})

    lines = path.read_text(encoding="utf-8").splitlines()
    assert [json.loads(ln)["event"] for ln in lines] == ["tool_start", "tool_end"]


def test_dict_style_tool_input_surfaces_the_bare_command_at_debug(caplog):
    rl = _logger()

    with caplog.at_level(logging.DEBUG, logger="triage_bot"):
        rl.on_tool_start(
            {"name": "shell"},
            "{'command': 'k8s status --wait-ready'}",
            inputs={"command": "k8s status --wait-ready"},
        )

    assert "k8s status --wait-ready" in caplog.text
    assert "{" not in caplog.text
    assert "'command'" not in caplog.text


def test_shell_tool_start_reaches_the_structured_sink(caplog):
    records = []
    rl = _logger(sink=records.append)

    with caplog.at_level(logging.INFO, logger="triage_bot"):
        rl.on_tool_start(
            {"name": "shell"},
            "{'command': 'k8s status --wait-ready'}",
            inputs={"command": "k8s status --wait-ready"},
        )

    assert records[0]["event"] == "tool_start"
    assert records[0]["tool"] == "shell"
    assert records[0]["detail"] == "k8s status --wait-ready"
