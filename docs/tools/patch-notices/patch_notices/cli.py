# Copyright 2026 Canonical, Ltd.
# See LICENSE file for licensing details.

"""CLI entry point for the patch-notices tool.

Commands
--------
fetch      Pull the PR/commit delta since the last documented SHA.
review     Run AI triage and write the Markdown workbook for manual editing.
finalize   Parse the edited workbook, update state, and write a clean export.
generate   Fetch + triage + insert into release notes in one step (CI use).
pr-body    Collate per-track summary JSONs into a PR body (CI use).
"""

from __future__ import annotations

import argparse
import json
import sys
from datetime import date
from pathlib import Path
from typing import Any

from rich.console import Console

from patch_notices import ai, fetcher, metadata, workbook

console = Console()


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _expand_track(version: str, source: str) -> str:
    """Expand a short version like '1.32' to the full channel string."""
    if "/" in version:
        return version  # already a full channel
    if source == "charm":
        return f"{version}/stable"
    return f"{version}-classic/stable"


def _make_channel_key(track: str, source: str) -> str:
    """Return the patch-metadata.json key, e.g. 'snap:1.32-classic/stable'."""
    return f"charm:{track}" if source == "charm" else f"snap:{track}"


def _fetch_delta(source: str, channel_key: str) -> list[dict[str, Any]]:
    """Fetch the commit delta for a track and return it."""
    if source == "charm":
        return fetcher.fetch_charm_delta(channel_key)
    return fetcher.fetch_snap_delta(channel_key)


def _write_json(path: str, data: dict[str, Any]) -> None:
    """Write *data* as JSON to *path*, creating parent directories as needed."""
    out = Path(path)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(data, indent=2))


# ---------------------------------------------------------------------------
# Commands
# ---------------------------------------------------------------------------

_WORKBOOK_PATH = "patch_notices_review.md"
_EXPORT_PATH = "patch_notices_output.md"


def cmd_fetch(args: argparse.Namespace) -> None:
    track = _expand_track(args.track, args.source)
    channel_key = _make_channel_key(track, args.source)
    console.print(f"[bold]Fetching delta for track:[/bold] {track} [dim](source: {args.source})[/dim]")
    delta = _fetch_delta(args.source, channel_key)
    console.print(f"[green]Found {len(delta)} commits.[/green]")
    console.print(f"[green]Delta written:[/green] {fetcher.delta_path(channel_key)}")


def cmd_review(args: argparse.Namespace) -> None:
    track = _expand_track(args.track, args.source)
    channel_key = _make_channel_key(track, args.source)
    console.print(f"[bold]Running AI triage for track:[/bold] {track} [dim](source: {args.source})[/dim]")
    delta = fetcher.load_delta(channel_key)
    triage_result = ai.triage(delta, source=args.source)
    workbook.write(triage_result, output_path=_WORKBOOK_PATH, channel_key=channel_key)
    console.print(f"[green]Workbook written:[/green] {_WORKBOOK_PATH}")


def cmd_finalize(args: argparse.Namespace) -> None:
    if not Path(_WORKBOOK_PATH).exists():
        console.print(f"[red]No workbook found at {_WORKBOOK_PATH}. Run `review` first.[/red]")
        sys.exit(1)
    channel_key = workbook.read_channel_key(_WORKBOOK_PATH)
    track = channel_key.split(":", 1)[1].split("/")[0].replace("-classic", "")
    console.print(f"[bold]Finalizing track:[/bold] {track}")

    try:
        delta = fetcher.load_delta(channel_key)
    except FileNotFoundError:
        console.print("[red]No delta found. Run `fetch` first.[/red]")
        sys.exit(1)

    included_shas = workbook.parse_included_shas(_WORKBOOK_PATH)
    if included_shas:
        workbook.export_clean(_WORKBOOK_PATH, _EXPORT_PATH)
        console.print(f"[green]Clean export written:[/green] {_EXPORT_PATH}")
        if not delta:
            console.print("[red]Workbook contains included items, but the saved delta is empty. Re-run `fetch` and `review`.[/red]")
            sys.exit(1)
        latest_sha = delta[-1]["sha"]
    else:
        if delta:
            latest_sha = delta[-1]["sha"]
            console.print("[yellow]No included items — advancing bookmark to delta head (nothing exported).[/yellow]")
        else:
            existing = metadata.load().get("tracks", {}).get(channel_key, {})
            latest_sha = existing.get("last_documented_sha")
            if not latest_sha:
                console.print("[red]No existing state for this track.[/red]")
                sys.exit(1)
            console.print("[dim]No new commits — updating date only.[/dim]")

    metadata.update(channel_key, latest_sha)
    console.print(f"[green]State updated.[/green] Latest SHA: {latest_sha}")


def cmd_generate(args: argparse.Namespace) -> None:
    track = _expand_track(args.track, args.source)
    channel_key = _make_channel_key(track, args.source)
    d = date.today()
    today = f"{d.strftime('%b')} {d.day}, {d.year}"
    console.print(f"[bold]Generating patch notice:[/bold] {track} [dim](source: {args.source})[/dim]")

    delta = _fetch_delta(args.source, channel_key)

    if not delta:
        _write_json(args.summary_out, {
            "track": channel_key,
            "status": "up-to-date",
            "date": today,
            "included": [],
            "discarded": [],
            "limited_context": [],
        })
        console.print("[dim]Up to date — no new commits.[/dim]")
        return

    triage_result = ai.triage(delta, source=args.source)
    included = [r for r in triage_result if r["triage"]["action"] == "include"]
    discarded = [r for r in triage_result if r["triage"]["action"] == "discard"]

    if not included:
        summary = workbook.build_track_summary(triage_result, channel_key, today)
        summary["status"] = "all-discarded"
        _write_json(args.summary_out, summary)
        metadata.update(channel_key, delta[-1]["sha"])
        console.print(
            f"[yellow]All {len(discarded)} commits discarded — bookmark advanced, "
            "release notes unchanged.[/yellow]"
        )
        return

    # Insert before updating state so a failure here is retryable.
    workbook.insert_patch_notice(args.release_notes, triage_result, today, args.source)
    metadata.update(channel_key, delta[-1]["sha"])

    summary = workbook.build_track_summary(triage_result, channel_key, today)
    _write_json(args.summary_out, summary)

    console.print(
        f"[green]✓[/green] {track} — "
        f"[green]{len(included)} included[/green], "
        f"[dim]{len(discarded)} discarded[/dim]"
        + (f", [yellow]{len(summary['limited_context'])} ⚠️[/yellow]" if summary["limited_context"] else "")
    )


def cmd_pr_body(args: argparse.Namespace) -> None:
    summary_dir = Path(args.summaries_dir)
    if not summary_dir.is_dir():
        console.print(f"[red]Directory not found:[/red] {args.summaries_dir}")
        sys.exit(1)

    summaries = [json.loads(f.read_text()) for f in sorted(summary_dir.glob("*.json"))]
    if not summaries:
        console.print(f"[red]No summary JSON files found in[/red] {args.summaries_dir}")
        sys.exit(1)

    body = workbook.build_pr_body(summaries)

    if args.output == "-":
        print(body)
    else:
        Path(args.output).write_text(body)
        console.print(f"[green]PR body written:[/green] {args.output}")


# ---------------------------------------------------------------------------
# Argument parser
# ---------------------------------------------------------------------------


def build_parser() -> argparse.ArgumentParser:
    # Shared --track / --source for commands that operate on a track.
    track_parent = argparse.ArgumentParser(add_help=False)
    
    track_parent.add_argument(
        "--track",
        required=True,
        help="Release version, e.g. '1.32'.",
    )
    track_parent.add_argument(
        "--source",
        default="snap",
        choices=["snap", "charm"],
        help="'snap' (default) or 'charm'.",
    )

    parser = argparse.ArgumentParser(
        prog="patch-notices",
        description="Patch-notices updater for Canonical Kubernetes.",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    p = sub.add_parser(
        "fetch",
        parents=[track_parent],
        help="Pull the PR delta since the last documented commit.",
    )
    p.set_defaults(func=cmd_fetch)

    p = sub.add_parser(
        "review",
        parents=[track_parent],
        help="Run AI triage and write the Markdown workbook.",
    )
    p.set_defaults(func=cmd_review)

    p = sub.add_parser(
        "finalize",
        help="Close the loop: update state and write the clean export.",
    )
    p.set_defaults(func=cmd_finalize)

    p = sub.add_parser(
        "generate",
        parents=[track_parent],
        help="Fetch, triage, and write directly to the release notes file.",
    )
    p.add_argument(
        "--release-notes",
        required=True,
        help="Path to the release notes file to update.",
    )
    p.add_argument(
        "--summary-out",
        required=True,
        help="Path to write the per-track summary JSON.",
    )
    p.set_defaults(func=cmd_generate)

    p = sub.add_parser(
        "pr-body",
        help="Generate the PR body markdown from per-track summary JSON files.",
    )
    p.add_argument(
        "--summaries-dir",
        required=True,
        help="Directory containing per-track summary JSON files.",
    )
    p.add_argument(
        "--output",
        default="-",
        help="Output file path. Use '-' for stdout (default).",
    )
    p.set_defaults(func=cmd_pr_body)

    return parser


def main() -> None:
    args = build_parser().parse_args()
    args.func(args)