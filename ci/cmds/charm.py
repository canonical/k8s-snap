#!/usr/bin/env python3
#
# Copyright 2026 Canonical, Ltd.
#
"""
Charm subcommands for `k8s-ci`.

This module provides functionality around charms and their channels.
"""

import argparse
import os
from http.client import HTTPException

from cmds.store import channel_available, query_info


def add_charm_cmds(parser: argparse.ArgumentParser) -> None:
    """
    Register charm-related subcommands to the given CLI parser.

    Args:
        parser: The parent argparse.ArgumentParser to which subcommands will be added.
    """
    charm_parser = parser.add_parser(
        "charm", help="Charm-related commands (e.g., check channel availability)."
    )
    charm_sub = charm_parser.add_subparsers(
        dest="charm_command", required=True, title="charm commands"
    )

    # Channel-available command
    p = charm_sub.add_parser(
        "channel-available",
        help="Check if a charm channel exists for the specified Kubernetes minor version.",
    )
    p.add_argument(
        "channel",
        help="Channel to check (e.g., '1.33/stable', '1.33/beta')",
    )
    p.add_argument(
        "--charm-auth",
        default=None,
        help="Charmcraft authentication token (or CHARMCRAFT_AUTH env var)",
    )
    p.add_argument(
        "--charm-name",
        default="k8s",
        help="Name of the charm to check (default: k8s)",
    )
    p.set_defaults(func=cmd_channel_available)


def cmd_channel_available(args: argparse.Namespace) -> int:
    """
    Command to check if a charm channel exists.

    Args:
        args: Parsed command-line arguments.

    Returns:
        0 if the channel is available, 1 otherwise.
    """
    auth_token = args.charm_auth or os.environ.get("CHARMCRAFT_AUTH")
    channel = args.channel

    try:
        charm_info = query_info("charm", args.charm_name, auth_token)
    except (OSError, HTTPException, ValueError) as exc:
        print(f"failed to query charm '{args.charm_name}': {exc}")
        return 1

    if channel_available(charm_info, channel):
        print(f"channel '{channel}' available for charm '{args.charm_name}'")
        return 0

    print(f"no matching charm channel '{channel}' for '{args.charm_name}'")
    return 1
