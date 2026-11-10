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
K8S_OPERATOR_URL = "https://github.com/canonical/k8s-operator.git"
K8S_OPERATOR_DEFAULT_REF = "main"


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
        help=(
            "Regenerate k8sd API documentation in place and fail if it differs "
            "from what is committed."
        ),
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

    p = docs_sub.add_parser(
        "update-charm-reference",
        help="Regenerate charm references and fail if they differ from what is committed.",
    )
    p.add_argument(
        "--operator-ref",
        default="main",
        help=(
            "k8s-operator branch or tag (default: main; use release-1.xx for release docs). "
            "Falls back to main if it does not exist."
        ),
    )
    p.set_defaults(func=cmd_update_charm_reference)


def _resolve_operator_ref(ref: str) -> str:
    """Return ref if it exists on k8s-operator, otherwise the default branch."""
    result = subprocess.run(
        [
            "git",
            "ls-remote",
            "--exit-code",
            K8S_OPERATOR_URL,
            f"refs/heads/{ref}",
            f"refs/tags/{ref}",
        ],
        capture_output=True,
        text=True,
    )
    # Exit code 2 means the remote is reachable but has no matching ref.
    if result.returncode == 2:
        print(
            f"k8s-operator has no branch or tag '{ref}', "
            f"falling back to '{K8S_OPERATOR_DEFAULT_REF}'"
        )
        return K8S_OPERATOR_DEFAULT_REF
    result.check_returncode()
    return ref


def cmd_update_charm_reference(args: argparse.Namespace) -> int:
    """Generate references from the matching operator branch and detect drift."""
    output = K8S_SNAP_ROOT / "docs/canonicalk8s/_parts/charms"
    try:
        operator_ref = _resolve_operator_ref(args.operator_ref)
        with tempfile.TemporaryDirectory() as tmpdir:
            operator = Path(tmpdir) / "k8s-operator"
            subprocess.run(
                [
                    "git",
                    "clone",
                    "--depth",
                    "1",
                    "--branch",
                    operator_ref,
                    "--",
                    K8S_OPERATOR_URL,
                    str(operator),
                ],
                check=True,
                capture_output=True,
                text=True,
            )
            revision = subprocess.run(
                ["git", "rev-parse", "HEAD"],
                cwd=operator,
                check=True,
                capture_output=True,
                text=True,
            ).stdout.strip()
            print(f"Using k8s-operator {operator_ref} at {revision}")
            subprocess.run(
                [
                    sys.executable,
                    str(K8S_SNAP_ROOT / "docs/tools/generate_charms_reference.py"),
                    str(operator / "charms/worker/k8s/charmcraft.yaml"),
                    str(operator / "charms/worker/charmcraft.yaml"),
                    "--output-dir",
                    str(output),
                ],
                check=True,
            )
        status = subprocess.run(
            [
                "git",
                "status",
                "--porcelain",
                "--untracked-files=all",
                "--",
                str(output),
            ],
            cwd=K8S_SNAP_ROOT,
            check=True,
            capture_output=True,
            text=True,
        ).stdout.strip()
        if status:
            diff = subprocess.run(
                ["git", "--no-pager", "diff", "HEAD", "--", str(output)],
                cwd=K8S_SNAP_ROOT,
                check=True,
                capture_output=True,
                text=True,
            ).stdout
            print("Error: charm reference differs from what is committed.")
            print("Regenerate locally and commit the resulting files:")
            print(
                "  python3 ci/k8s-ci.py docs update-charm-reference"
                f" --operator-ref {operator_ref}"
            )
            print("  git add docs/canonicalk8s/_parts/charms")
            print(status)
            print(diff)
            return 1
    except subprocess.CalledProcessError as error:
        print(f"Error updating charm reference: {error}")
        if error.stderr:
            print(error.stderr)
        return 1
    except OSError as error:
        print(f"Error updating charm reference: {error}")
        return 1
    print("Charm reference is up to date")
    return 0


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

        # Run make go.doc and set output to _parts/ dir
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

    # Compare docs generated in _parts/ to what is committed. The charms
    # directory is generated by update-charm-reference, so it is excluded.
    # `git status` also reports new and deleted files.
    docs_dir = docs_path.resolve()
    pathspec = ["--", ".", ":(exclude)charms"]
    result = subprocess.run(
        ["git", "status", "--porcelain", "--untracked-files=all", *pathspec],
        cwd=docs_dir,
        capture_output=True,
        text=True,
    )

    if result.returncode != 0:
        print(f"Error checking for docs changes: {result.stderr}")
        return 1

    status = result.stdout.rstrip()
    if status:
        diff = subprocess.run(
            ["git", "diff", "HEAD", *pathspec],
            cwd=docs_dir,
            capture_output=True,
            text=True,
        )
        try:
            display_dir = docs_dir.relative_to(K8S_SNAP_ROOT.resolve())
        except ValueError:
            display_dir = docs_dir
        print(f"\nError: k8sd API documentation in {display_dir} is out of date.")
        print("To fix, run the following locally and commit the resulting changes:")
        print("  python3 ci/k8s-ci.py docs update-k8sd-api")
        print(f"  git add {display_dir}")
        print("\ngit status:")
        print(status)
        if diff.stdout.strip():
            print("\ngit diff:")
            print(diff.stdout)
        return 1

    print("✓ k8sd API documentation is up to date")
    return 0


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    add_docs_cmds(parser)
    args_main = parser.parse_args()
    sys.exit(args_main.func(args_main))
