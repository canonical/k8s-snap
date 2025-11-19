#!/usr/bin/env python3
#
# Copyright 2025 Canonical, Ltd.
#
"""
Charm subcommands for `k8s-ci`.

This module provides functionality around charms and their releases.
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
    charm_parser = parser.add_parser("charm", help="Check charm release availability.")
    charm_sub = charm_parser.add_subparsers(
        dest="charm_command", required=True, title="charm commands"
    )

    # Release-available command
    p = charm_sub.add_parser(
        "release-available",
        help="Check if a charm release is available for the specified Kubernetes version.",
    )
    p.add_argument(
        "release",
        help="Kubernetes release version (e.g., '1.33', 'v1.33', 'v1.33.0', 'release-1.33')",
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
    p.set_defaults(func=cmd_release_available)


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
    print(resp.json())
    return resp.json()


def _check_release_available(charm_info: dict, release: str) -> bool:
    """
    Check if a specific release track is available in the charm info.

    Args:
        charm_info: The charm information from the Charmhub API.
        release: The release version to check (e.g., '1.30').

    Returns:
        True if the release track exists, False otherwise.
    """
    # Check if channel-map exists and contains the release track
    channel_map = charm_info.get("channel-map", [])

    # Look for channels that match the release track
    # Channels follow the pattern: <track>/<risk> (e.g., "1.34/stable", "1.34/beta")
    for channel_info in channel_map:
        channel = channel_info.get("channel", {})
        track = channel.get("track")
        if track.startswith(release):
            return True

    return False


def cmd_release_available(args: argparse.Namespace) -> int:
    """
    Command to check if a charm release is available.

    Args:
        args: Parsed command-line arguments.

    Returns:
        0 if the release is available, 1 otherwise.
    """
    auth_token = args.charm_auth or os.environ.get("CHARMCRAFT_AUTH")
    release = args.release.lstrip("v").replace("release-", "")

    charm_info = _query_charm_info(args.charm_name, auth_token)

    if _check_release_available(charm_info, release):
        print(f"Charm release {args.release} is available for {args.charm_name}")
        return 0
    else:
        print(f"no matching {args.charm_name} charm version for release {args.release}")
        return 1


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    add_charm_cmds(parser)
    args_main = parser.parse_args()
    sys.exit(args_main.func(args_main))
