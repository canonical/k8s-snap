#
# Copyright 2026 Canonical, Ltd.
#
"""The workflow and the CLI must agree about what flags exist.

`.github/workflows/ci-metrics.yaml` invokes the metrics CLI by string. Nothing
type-checks that contract, so a flag the workflow passes but the parser never
defined is invisible until a runner executes it -- and because the failing
command sits several minutes into the job, behind checkout, uv and a full
ingest, the feedback loop is a workflow run rather than a test.

That is not hypothetical: the workflow shipped calling `fetch-logs --since`,
which was never registered on that subparser, so the first real run would have
died with `unrecognized arguments: --since 7d` after doing all the ingest work.

Parsing every invocation here turns that into a local test failure.
"""

import re
from pathlib import Path

import pytest

from cmds.metrics import add_metrics_cmds

WORKFLOW = (
    Path(__file__).resolve().parents[3] / ".github" / "workflows" / "ci-metrics.yaml"
)

# `tox -e k8s-ci -- metrics ...`, possibly continued over backslash-newlines.
INVOCATION = re.compile(
    r"tox -e k8s-ci -- (metrics\b(?:[^\n\\]|\\\s*\n)*)",
    re.MULTILINE,
)


def _parser():
    import argparse

    parser = argparse.ArgumentParser(prog="k8s-ci")
    add_metrics_cmds(parser.add_subparsers(dest="command"))
    return parser


def _invocations():
    text = WORKFLOW.read_text()
    for match in INVOCATION.finditer(text):
        raw = match.group(1).replace("\\\n", " ")
        tokens = []
        for token in raw.split():
            # Shell expansions carry no argparse meaning; substitute a
            # literal so the token still occupies its position as a value.
            # "1" is used because some options are typed (--run-id is an
            # int), and the point of this test is to check flag names and
            # arity, not the runtime values a runner will supply.
            if token.startswith("$") or token.startswith('"$'):
                token = "1"
            tokens.append(token)
        yield " ".join(tokens), tokens


def test_workflow_exists():
    assert WORKFLOW.is_file(), f"missing {WORKFLOW}"


def test_workflow_invokes_the_cli():
    """A regex that silently matches nothing would make this file vacuous."""
    assert list(_invocations()), "no k8s-ci metrics invocations found"


@pytest.mark.parametrize(
    "raw,tokens", list(_invocations()), ids=lambda v: v if isinstance(v, str) else ""
)
def test_every_workflow_invocation_parses(raw, tokens):
    parser = _parser()
    try:
        parser.parse_args(tokens)
    except SystemExit as exc:  # argparse exits 2 on an unknown flag
        pytest.fail(f"workflow command does not parse: {raw} ({exc})")
