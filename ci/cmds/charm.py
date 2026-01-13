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
import sys
from typing import Optional

import requests


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


def _query_charm_info(charm_name: str, auth_token: Optional[str] = None) -> dict:
    """
    Query the Charmhub API for charm information.

    Args:
        charm_name: The name of the charm to query.
        auth_token: Optional authentication token for private charms.

    Returns:
        JSON response from the Charmhub API.
    """
    url = f"https://api.charmhub.io/v2/charms/info/{charm_name}?fields=channel-map"
    headers = {"Content-Type": "application/json"}

    if auth_token:
        headers["Authorization"] = f"Bearer {auth_token}"

    resp = requests.get(url, headers=headers, timeout=20)
    resp.raise_for_status()
    return resp.json()


def _check_channel_available(charm_info: dict, channel: str) -> bool:
    """
    Check if a specific charm channel is available.

    Args:
        charm_info: The charm information from the Charmhub API.
        channel: The full channel name to check (e.g., '1.33/stable').

    Returns:
        True if the channel exists, False otherwise.
    """
    channel_map = charm_info.get("channel-map", [])
    for entry in channel_map:
        ch = entry.get("channel", {})
        name = ch.get("name")
        if name == channel:
            return True
    return False


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

    charm_info = _query_charm_info(args.charm_name, auth_token)

    if _check_channel_available(charm_info, channel):
        print(f"channel '{channel}' available for charm '{args.charm_name}'")
        return 0

    print(f"no matching charm channel '{channel}' for '{args.charm_name}'")
    return 1


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    add_charm_cmds(parser)
    args_main = parser.parse_args()
    sys.exit(args_main.func(args_main))
