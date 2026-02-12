#!/usr/bin/env python3
#
# Copyright 2026 Canonical, Ltd.
#
"""
Docs subcommands for `k8s-ci`.

This module provides functionality for documentation management.
"""

import argparse
import os
import subprocess
import sys
import tempfile
from pathlib import Path

# Constants
K8S_SNAP_ROOT = Path(__file__).parent.parent.parent
K8SD_VERSION_FILE = K8S_SNAP_ROOT / "build-scripts/components/k8sd/version"
DOCS_OUTPUT_PATH = K8S_SNAP_ROOT / "docs/canonicalk8s/_parts"


def add_docs_cmds(parser: argparse.ArgumentParser) -> None:
    """
    Register docs-related subcommands to the given CLI parser.

    Args:
        parser: The parent argparse.ArgumentParser to which subcommands will be added.
    """
    docs_parser = parser.add_parser("docs", help="Documentation-related commands.")
    docs_sub = docs_parser.add_subparsers(
        dest="docs_command", required=True, title="docs commands"
    )

    # update-k8sd-api command
    p = docs_sub.add_parser(
        "update-k8sd-api",
        help="Check if k8sd API documentation is up to date.",
    )
    p.add_argument(
        "--k8sd-version-file",
        default=K8SD_VERSION_FILE,
        help=f"Path to the k8sd version file (default: {K8SD_VERSION_FILE})",
    )
    p.add_argument(
        "--docs-output-path",
        default=DOCS_OUTPUT_PATH,
        help=f"Path for DOCS_OUTPUT_PATH environment variable (default: {DOCS_OUTPUT_PATH})",
    )
    p.set_defaults(func=cmd_update_k8sd_api)


def cmd_update_k8sd_api(args: argparse.Namespace) -> int:
    """
    Check if k8sd API documentation is up to date.

    Args:
        args: Parsed command-line arguments.

    Returns:
        0 if documentation is up to date, 1 otherwise.
    """
    # Resolve version file path
    version_file = Path(args.k8sd_version_file)
    if not version_file.is_absolute():
        version_file = K8S_SNAP_ROOT / version_file

    if not version_file.exists():
        print(f"Error: Version file not found: {version_file}")
        return 1

    # Read k8sd version
    k8sd_version = version_file.read_text().strip()
    print(f"Using k8sd version: {k8sd_version}")

    # Create temporary directory for k8sd repo
    with tempfile.TemporaryDirectory() as tmpdir:
        k8sd_repo = Path(tmpdir) / "k8sd"

        # Clone k8sd repository
        print(f"Cloning k8sd repository at {k8sd_version}...")
        result = subprocess.run(
            [
                "git",
                "clone",
                "--depth",
                "1",
                "--branch",
                k8sd_version,
                "https://github.com/canonical/k8sd.git",
                str(k8sd_repo),
            ],
            capture_output=True,
            text=True,
        )

        if result.returncode != 0:
            print(f"Error cloning k8sd repository: {result.stderr}")
            return 1

        # Run go mod download
        print("Downloading Go dependencies...")
        result = subprocess.run(
            ["go", "mod", "download"], cwd=k8sd_repo, capture_output=True, text=True
        )

        if result.returncode != 0:
            print(f"Error downloading Go dependencies: {result.stderr}")
            return 1

        # Run make go.doc
        print("Generating API documentation...")
        env = os.environ.copy()
        docs_path = Path(args.docs_output_path)
        if not docs_path.is_absolute():
            docs_path = K8S_SNAP_ROOT / docs_path
        env["DOCS_OUTPUT_DIR"] = str(docs_path.resolve())
        result = subprocess.run(
            ["make", "go.doc"], cwd=k8sd_repo, capture_output=True, text=True, env=env
        )

        if result.returncode != 0:
            print(f"Error generating documentation: {result.stderr}")
            return 1

        # Check if there are any changes
        result = subprocess.run(
            ["git", "diff"], cwd=k8sd_repo, capture_output=True, text=True
        )

        if result.stdout.strip():
            print("\nError: Detected docs changes in k8sd API documentation.")
            print("Please run the following to regenerate the docs:")
            print("  python3 ci/k8s-ci.py docs update-k8sd-api")
            print("\ngit diff:")
            print(result.stdout)
            return 1

    print("✓ k8sd API documentation is up to date")
    return 0


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    add_docs_cmds(parser)
    args_main = parser.parse_args()
    sys.exit(args_main.func(args_main))
