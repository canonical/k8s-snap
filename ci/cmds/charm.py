#!/usr/bin/env python3
#
# Copyright 2025 Canonical, Ltd.
#
"""
Charm subcommands for `k8s-ci`.

This module provides functionality around charms and their channels.
"""

import argparse
import os
import shutil
import subprocess
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

    # Integration-test command
    p = charm_sub.add_parser(
        "integration-test",
        help="Run charm integration tests with k8s-snap.",
    )
    p.add_argument(
        "--k8s-snap-path",
        required=True,
        help="Path to the k8s snap file",
    )
    p.add_argument(
        "--charm-channel",
        default="latest/edge",
        help="Charm channel to use (default: latest/edge)",
    )
    p.add_argument(
        "--arch",
        default="amd64",
        help="Architecture (amd64 or arm64, default: amd64)",
    )
    p.add_argument(
        "--k8s-operator-repo",
        help="Path to k8s-operator repository (will clone if not provided)",
    )
    p.add_argument(
        "--k8s-operator-ref",
        help="Git ref to checkout for k8s-operator (branch, tag, or commit)",
    )
    p.add_argument(
        "--workspace",
        help="Workspace directory for test artifacts (default: current directory)",
    )
    p.set_defaults(func=cmd_integration_test)


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


def _run_command(
    cmd: list[str], check: bool = True, capture_output: bool = False
) -> subprocess.CompletedProcess:
    """
    Run a shell command and handle errors.

    Args:
        cmd: Command and arguments as a list.
        check: If True, raises CalledProcessError on non-zero exit.
        capture_output: If True, captures stdout and stderr.

    Returns:
        CompletedProcess instance.
    """
    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(cmd, check=check, capture_output=capture_output, text=True)
    if capture_output and result.stdout:
        print(result.stdout)
    if capture_output and result.stderr:
        print(result.stderr, file=sys.stderr)
    return result


def _setup_lxd() -> None:
    """Set up LXD for testing."""
    print("Setting up LXD...")
    # Disable IPv6 for LXD bridge (charm tests can't handle IPv6 dualstack)
    _run_command(["lxc", "network", "set", "lxdbr0", "ipv6.address", "none"])


def _create_snap_tarball(snap_path: str, workspace: str) -> str:
    """
    Create a tarball containing the k8s snap.

    Args:
        snap_path: Path to the k8s.snap file.
        workspace: Workspace directory for artifacts.

    Returns:
        Path to the created tarball.
    """
    print(f"Creating snap tarball from {snap_path}...")
    snap_installation_dir = os.path.join(workspace, "snap_installation")
    os.makedirs(snap_installation_dir, exist_ok=True)

    # Copy snap to installation directory
    shutil.copy(snap_path, os.path.join(snap_installation_dir, "k8s.snap"))

    # Create tarball
    tarball_path = os.path.join(workspace, "snap_installation.tar.gz")
    _run_command(["tar", "cvzf", tarball_path, "-C", snap_installation_dir, "."])

    print(f"Snap tarball created: {tarball_path}")
    return tarball_path


def _install_tools() -> None:
    """Install required tools (charmcraft, juju, juju-crashdump)."""
    print("Installing charmcraft, juju, and juju-crashdump...")
    tools = ["charmcraft", "juju", "juju-crashdump"]
    for tool in tools:
        _run_command(["sudo", "snap", "install", tool, "--classic"])


def _download_charms(charm_channel: str, arch: str, workspace: str) -> tuple[str, str]:
    """
    Download k8s and k8s-worker charms.

    Args:
        charm_channel: Charm channel to download from.
        arch: Architecture (amd64 or arm64).
        workspace: Workspace directory for artifacts.

    Returns:
        Tuple of (k8s_charm_path, k8s_worker_charm_path).
    """
    print(f"Downloading charms from channel {charm_channel} for {arch}...")

    k8s_charm_file = os.path.join(workspace, f"k8s_ubuntu-22.04-{arch}.charm")
    k8s_worker_charm_file = os.path.join(
        workspace, f"k8s-worker_ubuntu-22.04-{arch}.charm"
    )

    _run_command(
        [
            "juju",
            "download",
            "k8s",
            "--channel",
            charm_channel,
            "--filepath",
            k8s_charm_file,
            "--base",
            "ubuntu@22.04",
        ]
    )

    _run_command(
        [
            "juju",
            "download",
            "k8s-worker",
            "--channel",
            charm_channel,
            "--filepath",
            k8s_worker_charm_file,
            "--base",
            "ubuntu@22.04",
        ]
    )

    print(f"Downloaded charms: {k8s_charm_file}, {k8s_worker_charm_file}")
    return k8s_charm_file, k8s_worker_charm_file


def _bootstrap_juju() -> None:
    """Bootstrap Juju controller on LXD."""
    print("Bootstrapping Juju controller on LXD...")
    _run_command(["juju", "bootstrap", "localhost", "lxd"])


def _clone_k8s_operator(workspace: str, ref: Optional[str] = None) -> str:
    """
    Clone the k8s-operator repository.

    Args:
        workspace: Workspace directory.
        ref: Git ref to checkout (branch, tag, or commit).

    Returns:
        Path to the cloned repository.
    """
    print("Cloning k8s-operator repository...")
    operator_path = os.path.join(workspace, "k8s-operator")

    cmd = [
        "git",
        "clone",
        "https://github.com/canonical/k8s-operator.git",
        operator_path,
    ]
    _run_command(cmd)

    if ref:
        print(f"Checking out ref: {ref}")
        _run_command(["git", "-C", operator_path, "checkout", ref])

    return operator_path


def _run_integration_tests(
    operator_path: str,
    k8s_charm_file: str,
    k8s_worker_charm_file: str,
    snap_tarball: str,
) -> int:
    """
    Run integration tests using tox.

    Args:
        operator_path: Path to k8s-operator repository.
        k8s_charm_file: Path to k8s charm file.
        k8s_worker_charm_file: Path to k8s-worker charm file.
        snap_tarball: Path to snap installation tarball.

    Returns:
        Exit code from the test run.
    """
    print("Running integration tests...")
    print(f"  k8s charm: {k8s_charm_file}")
    print(f"  k8s-worker charm: {k8s_worker_charm_file}")
    print(f"  snap tarball: {snap_tarball}")

    result = _run_command(
        [
            "tox",
            "-r",
            "-e",
            "integration",
            "--",
            "--charm-file",
            k8s_charm_file,
            "--charm-file",
            k8s_worker_charm_file,
            "--snap-installation-resource",
            snap_tarball,
            "tests/integration/test_k8s.py::test_nodes_ready",
        ],
        check=False,
        capture_output=False,
    )

    return result.returncode


def _collect_juju_status(base_path: str) -> None:
    """
    Collect Juju status and create crashdump for debugging.

    Args:
        base_path: Base directory where tmp/ subdirectory will be created for artifacts.
    """
    print("Collecting Juju status...")
    debug_path = os.path.join(base_path, "tmp")
    os.makedirs(debug_path, exist_ok=True)

    # Collect juju status
    status_file = os.path.join(debug_path, "juju-status.txt")
    try:
        result = _run_command(["juju", "status"], check=False, capture_output=True)
        with open(status_file, "w") as f:
            f.write(result.stdout)
            if result.stderr:
                f.write("\nSTDERR:\n")
                f.write(result.stderr)
        print(f"Juju status saved to {status_file}")
    except Exception as e:
        print(f"Failed to collect juju status: {e}")

    # Create crashdump
    try:
        _run_command(
            [
                "juju-crashdump",
                "-s",
                "-m",
                "controller",
                "-a",
                "debug-layer",
                "-a",
                "config",
                "-o",
                debug_path + "/",
            ],
            check=False,
        )
        print(f"Juju crashdump saved to {debug_path}")
    except Exception as e:
        print(f"Failed to create crashdump: {e}")


def cmd_integration_test(args: argparse.Namespace) -> int:
    """
    Command to run charm integration tests.

    Args:
        args: Parsed command-line arguments.

    Returns:
        0 on success, non-zero on failure.
    """
    workspace = args.workspace or os.getcwd()
    workspace = os.path.abspath(workspace)
    print(f"Using workspace: {workspace}")

    # Ensure workspace directory exists
    os.makedirs(workspace, exist_ok=True)

    try:
        # Step 1: Setup LXD
        _setup_lxd()

        # Step 2: Create snap tarball
        snap_tarball = _create_snap_tarball(args.k8s_snap_path, workspace)

        # Step 3: Install tools
        _install_tools()

        # Step 4: Download charms
        k8s_charm_file, k8s_worker_charm_file = _download_charms(
            args.charm_channel, args.arch, workspace
        )

        # Step 5: Bootstrap Juju
        _bootstrap_juju()

        # Step 6: Clone or use existing k8s-operator repository
        if args.k8s_operator_repo:
            operator_path = os.path.abspath(args.k8s_operator_repo)
            print(f"Using existing k8s-operator repository at {operator_path}")
        else:
            operator_path = _clone_k8s_operator(workspace, args.k8s_operator_ref)

        # Step 7: Run integration tests
        os.chdir(operator_path)
        exit_code = _run_integration_tests(
            operator_path, k8s_charm_file, k8s_worker_charm_file, snap_tarball
        )

        # Step 8: Collect debug info on failure
        if exit_code != 0:
            _collect_juju_status(operator_path)

        return exit_code

    except Exception as e:
        print(f"Error running integration tests: {e}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    add_charm_cmds(parser)
    args_main = parser.parse_args()
    sys.exit(args_main.func(args_main))
