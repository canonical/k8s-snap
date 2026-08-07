#
# Copyright 2026 Canonical, Ltd.
#
"""``k8s-ci triage`` subcommand.

Event-driven entry point for the label-FSM triage bot. In CI it consumes the
GitHub Actions webhook payload (``GITHUB_EVENT_PATH``); locally it can be driven
manually with ``--issue``/``--action`` or from a saved fixture. The heavy
LangGraph/LangChain stack is imported lazily so the base ``k8s-ci`` CLI stays
dependency-light.
"""

import argparse
import json
import logging
import os
import sys

from triage_bot.llm import DEFAULT_MODEL

log = logging.getLogger("triage_bot.cli")


def _configure_logging(level: str) -> None:
    """Send the run log to stderr, keeping stdout a clean JSON result.

    Without this the bot's own records sit below the root logger's default
    level and are discarded -- a failed run would then have to be reproduced
    to be understood. At the default INFO level you get one line per pipeline
    stage (``[reproduce] reproducible=True``, and so on); the granular
    per-tool-call trace (every shell command and its output) only appears at
    ``--log-level DEBUG``. LLM/tool failures always log at ERROR, regardless
    of the configured level.
    """
    logging.basicConfig(
        level=getattr(logging, level),
        format="%(asctime)s %(levelname)-7s %(name)s: %(message)s",
        stream=sys.stderr,
    )


def add_triage_cmds(parser):
    """Register the ``triage`` command group on the ``k8s-ci`` CLI."""
    triage = parser.add_parser("triage", help="Autonomous issue triage bot.")
    sub = triage.add_subparsers(
        dest="triage_command", required=True, title="triage commands"
    )
    _add_run_parser(sub)


def _add_run_parser(sub):
    """Register the ``run`` subcommand that handles one event."""
    p = sub.add_parser("run", help="Handle one issue/comment event.")
    src = p.add_mutually_exclusive_group()
    src.add_argument(
        "--event-file",
        metavar="PATH",
        help="GitHub Actions webhook payload (defaults to $GITHUB_EVENT_PATH).",
    )
    src.add_argument(
        "--issue", type=int, help="Issue number, for manual/local invocation."
    )
    p.add_argument(
        "--action",
        default="opened",
        help="Event action for --issue mode (opened|reopened|closed|created).",
    )
    p.add_argument("--model", default=DEFAULT_MODEL, help="provider:model spec.")
    p.add_argument(
        "--dry-run",
        action="store_true",
        help="Compute transitions but write nothing to GitHub.",
    )
    p.add_argument(
        "--auto-pr",
        action="store_true",
        help="Open a draft fix PR when the fix stage succeeds.",
    )
    p.add_argument(
        "--bot-login",
        action="append",
        default=[],
        metavar="LOGIN",
        help="Additional bot login to ignore as a comment author (repeatable).",
    )
    p.add_argument(
        "--jsonl", metavar="PATH", help="Append structured run records here."
    )
    p.add_argument(
        "--log-level",
        default="INFO",
        choices=("DEBUG", "INFO", "WARNING", "ERROR"),
        help=(
            "Verbosity on stderr (default: INFO -- one line per pipeline "
            "stage; DEBUG adds every shell command the agent runs)."
        ),
    )
    p.set_defaults(func=cmd_triage_run)


def _load_event(args, gh) -> dict:
    """Build a webhook-style payload from the chosen input source.

    In ``--issue`` mode the run is maintainer-initiated, so the issue is
    fetched for its real labels (needed to route ``created`` events) and the
    actor is stamped ``OWNER`` -- a manual invocation is inherently trusted.
    """
    if args.issue is not None:
        issue = gh.get_issue(args.issue)
        payload = {
            "action": args.action,
            "issue": {
                "number": args.issue,
                "labels": issue.get("labels", []),
                "author_association": "OWNER",
            },
        }
        if args.action == "created":
            payload["comment"] = {
                "user": {"login": "maintainer"},
                "author_association": "OWNER",
            }
        return payload
    path = args.event_file or os.environ.get("GITHUB_EVENT_PATH")
    if not path:
        raise ValueError(
            "no event source: pass --event-file, --issue, or set GITHUB_EVENT_PATH"
        )
    with open(path, encoding="utf-8") as fh:
        return json.load(fh)


def cmd_triage_run(args: argparse.Namespace) -> int:
    from triage_bot.context import ActionContext
    from triage_bot.github import GitHubClient, GitHubError
    from triage_bot.handlers import Runtime, dispatch
    from triage_bot.pipeline import run_pipeline
    from triage_bot.router import GitHubEvent

    _configure_logging(args.log_level)
    dry_run = args.dry_run
    ctx = ActionContext(
        dry_run=dry_run,
        auto_pr=args.auto_pr,
        triage_model=args.model,
        jsonl_path=args.jsonl,
    ).with_bot_logins(args.bot_login)

    gh = GitHubClient(repo=ctx.repo, dry_run=dry_run)
    try:
        payload = _load_event(args, gh)
    except (OSError, ValueError, json.JSONDecodeError) as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1
    except GitHubError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1

    event = GitHubEvent.from_payload(payload, bot_logins=ctx.bot_logins)

    rt = Runtime(ctx=ctx, gh=gh, pipeline=run_pipeline)
    try:
        result = dispatch(event, rt)
    except GitHubError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1
    except Exception:  # a bad run must not crash the workflow
        log.exception("triage failed")
        return 1

    print(
        json.dumps(
            {
                "action": result.action,
                "outcome": result.outcome,
                "label": result.label,
                "actions_taken": result.actions_taken,
            },
            indent=2,
        )
    )
    return 0


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    _add_run_parser(parser.add_subparsers(dest="command", required=True))
    parsed = parser.parse_args()
    sys.exit(parsed.func(parsed))
