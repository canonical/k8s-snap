#!/usr/bin/env python3
#
# Copyright 2026 Canonical, Ltd.
#
"""DEMO SCRIPT -- NOT PART OF THE PRODUCT. DELETE BEFORE THIS PR MERGES.

Runs the real triage-bot decision engine -- classification, duplicate
detection, the existing-support/docs check, enhancement proposals, and the
trust gate -- against live, real k8s-snap issues, through the exact same
``dispatch()`` entry point the GitHub Action uses, and prints a readable
report of what the bot would do to each one.

Safety:
  * Every GitHub write (label, comment, branch, PR) is intercepted by
    ``RecordingGitHubClient`` below and only ever recorded, never sent --
    this holds regardless of ``dry_run``, so this script cannot mutate the
    live repository no matter what.
  * The expensive reproduce -> verify -> reproducer -> diagnose -> fix
    pipeline (self-hosted runner, LXD cluster, tens of minutes) is never
    invoked. A stub raises a clearly-named marker exception the report
    below recognises and narrates honestly, rather than fabricating a fake
    result.
  * Everything else -- fetching the real issue, the real GitHub duplicate
    search, and the real LLM classification/support/enhancement calls --
    runs for real, so the report reflects genuine model output.

Usage (from the repo root):
    PYTHONPATH=ci GEMINI_API_KEY=... python3 ci/demo_triage_5_issues.py
    PYTHONPATH=ci GEMINI_API_KEY=... python3 ci/demo_triage_5_issues.py --issue 2646 --issue 2672
"""

from __future__ import annotations

import argparse
import json
import logging
import os
import subprocess
import sys
import textwrap
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from triage_bot.context import ActionContext  # noqa: E402
from triage_bot.github import GitHubClient  # noqa: E402
from triage_bot.handlers import Runtime, dispatch  # noqa: E402
from triage_bot.router import GitHubEvent  # noqa: E402


class PipelineNotRunInDemo(Exception):
    """Raised by the stub pipeline in place of a real cluster run.

    Recognised by name in ``_render`` below so the report says what actually
    happened (nothing was run) instead of the generic "internal error"
    wording ``handle_triage`` uses for a genuine pipeline crash.
    """


def _demo_pipeline(rt, issue):
    raise PipelineNotRunInDemo()


class _QuietDemoStop(logging.Filter):
    """Mutes the traceback ``handle_triage`` logs for a pipeline exception --
    but only for our own marker exception; a real error still logs loudly."""

    def filter(self, record: logging.LogRecord) -> bool:
        return not (record.exc_info and record.exc_info[0] is PipelineNotRunInDemo)


class RecordingGitHubClient(GitHubClient):
    """Real reads; every write is captured, never sent.

    This holds independent of ``dry_run`` (which the base client also
    respects) as a second, belt-and-suspenders guarantee: this script must
    never be able to mutate the live repository, even if a future change to
    ``dry_run`` handling regressed that check.
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.recorded: list[tuple[str, object]] = []

    def add_labels(self, number, labels):
        if labels:
            self.recorded.append(("add_labels", list(labels)))

    def add_comment(self, number, body):
        if body:
            self.recorded.append(("add_comment", body))

    def remove_label(self, number, label):
        self.recorded.append(("remove_label", label))

    def push_branch(self, branch, cwd):
        self.recorded.append(("push_branch (blocked)", branch))
        return False

    def create_pull_request(self, **kwargs):
        self.recorded.append(("create_pull_request (blocked)", kwargs.get("title")))
        return None

    def delete_branch(self, branch):
        self.recorded.append(("delete_branch (blocked)", branch))


# --- terminal styling (no extra dependency for a script this short-lived) --

# A CI runner pipes stdout (not a tty), but its log viewer still renders
# ANSI colour -- GitHub Actions in particular does. PY_COLORS/FORCE_COLOR is
# the same override tox.ini's [testenv] already sets for every environment,
# so `tox -e demo` gets colour for free; a real terminal still works via
# isatty without either variable set.
_COLOR = (
    sys.stdout.isatty()
    or os.environ.get("PY_COLORS") == "1"
    or bool(os.environ.get("FORCE_COLOR"))
)


def _c(code: str, text: str) -> str:
    return f"\033[{code}m{text}\033[0m" if _COLOR else text


BOLD, DIM, CYAN, GREEN, YELLOW, RED = "1", "2", "36", "32", "33", "31"


def _latest_open_issue_numbers(repo: str, count: int) -> list[int]:
    out = subprocess.run(
        [
            "gh",
            "issue",
            "list",
            "--repo",
            repo,
            "--state",
            "open",
            "--limit",
            str(count),
            "--json",
            "number",
            "--search",
            "sort:created-desc",
        ],
        capture_output=True,
        text=True,
        check=True,
        timeout=30,
    ).stdout
    return [item["number"] for item in json.loads(out)]


def _render(
    number: int, real_issue: dict, event: GitHubEvent, result, recorded
) -> None:
    title = real_issue.get("title", "")
    url = real_issue.get("html_url", "")
    print(_c(BOLD + ";" + CYAN, f"--- #{number}: {title} ---"))
    print(_c(DIM, f"    {url}"))
    trust_word = "trusted" if event.trusted else "untrusted"
    print(
        _c(DIM, f"    reporter association: {event.author_association or '(none)'} ")
        + _c(BOLD, f"[{trust_word}]")
    )

    pipeline_reached = any(
        "PipelineNotRunInDemo" in action for action in result.actions_taken
    )
    other_actions = [a for a in result.actions_taken if "PipelineNotRunInDemo" not in a]

    if pipeline_reached:
        print(
            _c(GREEN, "    -> confirmed bug, trusted reporter: would hand off to the ")
            + _c(BOLD, "reproduce -> verify -> reproducer -> diagnose -> fix")
            + _c(GREEN, " pipeline")
        )
        print(
            _c(DIM, "       (skipped in this demo -- needs a self-hosted LXD runner)")
        )
    else:
        print(f"    {'label:':<12}{_c(YELLOW, result.label or '(none)')}")
        print(f"    {'outcome:':<12}{result.outcome}")
    if other_actions:
        print(f"    {'actions:':<12}{', '.join(other_actions)}")

    for kind, payload in recorded:
        if kind == "add_labels":
            continue  # already reflected in the "actions" line above
        if kind == "add_comment":
            print(_c(DIM, "    would post this comment:"))
            print(textwrap.indent(str(payload).strip(), "      "))
        else:
            print(_c(DIM, f"    would {kind}: {payload}"))
    print()


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", default="canonical/k8s-snap")
    parser.add_argument(
        "--count",
        type=int,
        default=5,
        help="how many of the latest open issues to demo",
    )
    parser.add_argument(
        "--issue",
        type=int,
        action="append",
        dest="issues",
        help="explicit issue number (repeatable); overrides --count",
    )
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.WARNING, format="%(levelname)s %(name)s: %(message)s"
    )
    logging.getLogger("triage_bot").addFilter(_QuietDemoStop())

    gh = RecordingGitHubClient(repo=args.repo, dry_run=True)
    ctx = ActionContext(repo=args.repo, dry_run=True, auto_pr=False)
    rt = Runtime(ctx=ctx, gh=gh, pipeline=_demo_pipeline)

    issue_numbers = args.issues or _latest_open_issue_numbers(args.repo, args.count)

    print(
        _c(
            BOLD,
            f"\n=== Autonomous triage bot -- live demo on "
            f"{len(issue_numbers)} issue(s) from {args.repo} ===\n",
        )
    )
    print(
        _c(
            DIM,
            "Real GitHub reads and real LLM calls; every write below is "
            "intercepted and only printed, never sent.\n",
        )
    )

    had_error = False
    for number in issue_numbers:
        gh.recorded.clear()
        real_issue = gh.get_issue(number)
        payload = {
            "action": "opened",
            "issue": {
                "number": number,
                "labels": real_issue.get("labels", []),
                "author_association": real_issue.get("author_association", ""),
            },
        }
        event = GitHubEvent.from_payload(payload, bot_logins=ctx.bot_logins)
        try:
            result = dispatch(event, rt)
        except Exception as exc:
            # A per-issue failure must not silently exit 0 -- that would look
            # like a working demo when nothing actually ran (this hid a real
            # missing-secret misconfiguration the first time this ran in CI).
            had_error = True
            print(_c(RED, f"--- #{number}: demo-side error: {exc!r} ---\n"))
            continue
        _render(number, real_issue, event, result, gh.recorded)

    return 1 if had_error else 0


if __name__ == "__main__":
    raise SystemExit(main())
