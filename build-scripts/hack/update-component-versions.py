#!/usr/bin/env python3

USAGE = "Update component versions for Canonical Kubernetes"

DESCRIPTION = """
Update the component versions that we use to build Canonical Kubernetes. This
script updates the individual `build-scripts/components/<component>/version`
files. The logic for each component is different, and is managed by configuration
options found in this script.
"""


import argparse
import json
import logging
import subprocess
import sys
import tempfile
import yaml
from packaging.version import Version
from pathlib import Path
from typing import Callable, Optional
import re
import util
import urllib.request


logging.basicConfig(level=logging.INFO)

LOG = logging.getLogger(__name__)

DIR = Path(__file__).absolute().parent
SNAPCRAFT = DIR.parent.parent / "snap/snapcraft.yaml"
COMPONENTS = DIR.parent / "components"
CHARTS = DIR.parent.parent / "k8s" / "manifests" / "charts"

# Version marker for latest Kubernetes version. Expected to be one of:
#
# - "https://dl.k8s.io/release/stable.txt"
# - "https://dl.k8s.io/release/stable-1.xx.txt"
# - "https://dl.k8s.io/release/latest-1.xx.txt" (e.g. for release candidate builds)
KUBERNETES_VERSION_MARKER = "https://dl.k8s.io/release/stable.txt"

# Containerd release branch to track. The most recent tag in the branch will be used.
CONTAINERD_RELEASE_BRANCH = "release/1.6"

# Helm release semver limit
#
# - None for main branch
# - Version("3.14") (e.g. for release candidate builds)
HELM_BRANCH, HELM_RELEASE_SEMVER = "main", None

# MetalLB Helm repository and chart version
METALLB_REPO = "https://metallb.github.io/metallb"
METALLB_CHART_VERSION = "0.14.8"


def is_valid_version(pinned_ver: None | Version) -> Callable[[None | Version], bool]:
    """filter function to check if version is valid

    Valid version is defined as:

    1.  Check if version is not None
    2.  Check if version is not a pre-release or dev-release
    3.  Check if version is not outside the scope of the optionally pinned version
    """

    def _validate(version: None | Version) -> bool:
        return (
            version is not None
            and not version.is_prerelease
            and not version.is_devrelease
            and (
                not pinned_ver
                or (version.major, version.minor)
                == (pinned_ver.major, pinned_ver.minor)
            )
        )

    return _validate


def parse_version(version: str) -> Version | None:
    try:
        return Version(version.removeprefix("v"))
    except ValueError:
        return None


def get_latest_upstream_version(repo_url: str, current_version: str) -> Optional[str]:
    """Get the latest upstream patch version for a component.

    This function fetches all tags from the upstream repository and returns
    the latest patch version that matches the current major.minor version.
    """
    try:
        # Use a temporary directory and fetch all tags without checking out a specific branch
        with tempfile.TemporaryDirectory() as tmpdir:
            # Clone just the tags, no specific branch needed
            subprocess.run(
                ["git", "clone", "--bare", "--filter=blob:none", repo_url, tmpdir],
                check=True,
                capture_output=True,
            )
            tags = (
                subprocess.run(
                    ["git", "tag"], cwd=tmpdir, check=True, capture_output=True
                )
                .stdout.decode()
                .strip()
                .split()
            )

            current = parse_version(current_version)
            if not current:
                return None

            # Filter to same major.minor, get latest patch
            same_minor = is_valid_version(Version(f"{current.major}.{current.minor}.0"))
            releases = sorted(filter(same_minor, map(parse_version, tags)))
            if releases:
                return f"v{releases[-1]}"
    except Exception as e:
        LOG.warning("Failed to fetch upstream version for %s: %s", repo_url, e)
    return None


def get_kubernetes_version() -> str:
    """Update Kubernetes version based on the specified marker file"""
    LOG.info("Checking latest Kubernetes version from %s", KUBERNETES_VERSION_MARKER)
    return util.read_url(KUBERNETES_VERSION_MARKER)


def get_cni_upstream_version() -> str:
    """Get latest version of CNI from upstream repository"""
    cni_repo = util.read_file(COMPONENTS / "cni/repository")
    current_version = util.read_file(COMPONENTS / "cni/version")

    latest_version = get_latest_upstream_version(cni_repo, current_version)
    if latest_version is None:
        raise RuntimeError("Failed to determine latest upstream CNI version")

    return latest_version


def get_cni_kubernetes_version() -> str:
    """Update CNI version to match the CNI version used in $kubernetes/build/dependencies.yaml"""
    kube_repo = util.read_file(COMPONENTS / "kubernetes/repository")
    kube_version = util.read_file(COMPONENTS / "kubernetes/version")

    with util.git_repo(kube_repo, kube_version) as dir:
        deps_file = dir / "build/dependencies.yaml"
        deps = yaml.safe_load(util.read_file(deps_file))

        for dep in deps["dependencies"]:
            if dep["name"] == "cni":
                ersion = dep["version"]
                return f"v{ersion.lstrip('v')}"

        raise Exception(f"Failed to find cni dependency in {deps_file}")


def get_containerd_version() -> str:
    """Update containerd version using latest tag of specified branch"""
    containerd_repo = util.read_file(COMPONENTS / "containerd/repository")

    with util.git_repo(
        containerd_repo, CONTAINERD_RELEASE_BRANCH, shallow=False
    ) as dir:
        # Get the latest tagged release from the current branch
        return util.parse_output(["git", "describe", "--tags", "--abbrev=0"], cwd=dir)


def get_runc_upstream_version() -> str:
    """Get latest version of runc from upstream repository"""
    runc_repo = util.read_file(COMPONENTS / "runc/repository")
    current_version = util.read_file(COMPONENTS / "runc/version")

    latest_version = get_latest_upstream_version(runc_repo, current_version)
    if latest_version is None:
        raise RuntimeError("Failed to determine latest upstream runc version")

    return latest_version


def get_runc_containerd_version() -> str:
    """Update runc version based on containerd"""
    containerd_repo = util.read_file(COMPONENTS / "containerd/repository")
    containerd_version = util.read_file(COMPONENTS / "containerd/version")

    with util.git_repo(containerd_repo, containerd_version) as dir:
        # See https://github.com/containerd/containerd/blob/main/docs/RUNC.md
        return util.read_file(dir / "script/setup/runc-version")


def get_helm_version() -> str:
    """Get latest version of helm"""

    helm_repo = util.read_file(COMPONENTS / "helm/repository")
    with util.git_repo(helm_repo, HELM_BRANCH, shallow=False) as dir:
        tags = util.parse_output(["git", "tag"], cwd=dir).split()
        # Parse tag strings to Version objects, then use by_helm_releases
        # to filter conditionally.
        by_helm_releases = is_valid_version(HELM_RELEASE_SEMVER)
        releases = sorted(filter(by_helm_releases, map(parse_version, tags)))
        if not releases:
            raise ValueError("No valid helm releases found")

        return f"v{releases[-1]}"


def pull_metallb_chart() -> None:
    LOG.info("Pulling MetalLB chart @ %s", METALLB_CHART_VERSION)
    util.helm_pull("metallb", METALLB_REPO, METALLB_CHART_VERSION, CHARTS)


def update_go_version(dry_run: bool):
    k8s_version = (COMPONENTS / "kubernetes/version").read_text().strip()
    url = f"https://raw.githubusercontent.com/kubernetes/kubernetes/refs/tags/{k8s_version}/.go-version"
    with urllib.request.urlopen(url) as response:
        go_version = response.read().decode("utf-8").strip()

    LOG.info("Upstream go version is %s", go_version)
    go_snap = f"go/{'.'.join(go_version.split('.')[:2])}/stable"
    snapcraft_yaml = SNAPCRAFT.read_text()
    if f"- {go_snap}" in snapcraft_yaml:
        LOG.info("snapcraft.yaml already contains go version %s", go_snap)
        return

    LOG.info("Update go snap version to %s in %s", go_snap, SNAPCRAFT)
    if not dry_run:
        updated = re.sub(r"- go/\d+\.\d+/stable", f"- {go_snap}", snapcraft_yaml)
        SNAPCRAFT.write_text(updated)


def collect_and_apply_component_updates(dry_run: bool) -> dict:
    """Collect component updates, apply them to version files, and return PR metadata.

    Returns a dict with:
    - title: PR title listing all updated components
    - description: PR description with version bumps and warnings
    """
    updates = []
    independent_warnings = []

    # Check all components for updates (with parent component checks where applicable)
    components_to_check = [
        ("kubernetes", get_kubernetes_version, None, None),
        ("cni", get_cni_upstream_version, "kubernetes", get_cni_kubernetes_version),
        ("containerd", get_containerd_version, None, None),
        ("runc", get_runc_upstream_version, "containerd", get_runc_containerd_version),
        ("helm", get_helm_version, None, None),
    ]

    for (
        component,
        get_version_func,
        parent_component,
        parent_version_requirement_func,
    ) in components_to_check:
        try:
            current_version = util.read_file(COMPONENTS / component / "version")
            new_upstream_version = get_version_func()
            parent_version_requirement = (
                parent_version_requirement_func() if parent_component else None
            )

            # Check if there's an update needed
            if current_version.strip() != new_upstream_version.strip():
                LOG.info(
                    "Found %s update: %s -> %s",
                    component,
                    current_version.strip(),
                    new_upstream_version.strip(),
                )

                if (
                    parent_version_requirement is not None
                    and new_upstream_version != parent_version_requirement
                ):
                    LOG.info(
                        "Requirement deviation detected for %s: upstream=%s, parent requires=%s",
                        component,
                        new_upstream_version,
                        parent_version_requirement,
                    )

                    independent_warnings.append(
                        f"- **{component}**: {current_version.strip()} → {new_upstream_version} "
                        f"(upstream has newer patches than parent {parent_component} expects {parent_version_requirement})"
                    )

                updates.append(
                    {
                        "component": component,
                        "old_version": current_version.strip(),
                        "new_version": new_upstream_version.strip(),
                    }
                )

                # Apply the update
                if not dry_run:
                    Path(COMPONENTS / component / "version").write_text(
                        new_upstream_version.strip() + "\n"
                    )
            else:
                LOG.info(
                    "No update needed for %s (current: %s)",
                    component,
                    current_version.strip(),
                )
        except Exception as e:
            LOG.warning("Failed to check/update %s: %s", component, e)

    # Also update go version if kubernetes was updated
    try:
        update_go_version(dry_run)
    except Exception as e:
        LOG.warning("Failed to update go version: %s", e)

    # Pull helm charts
    try:
        if not dry_run:
            pull_metallb_chart()
    except Exception as e:
        LOG.warning("Failed to pull helm charts: %s", e)

    # Generate PR title and description
    if not updates:
        return {"title": "", "description": ""}

    # Build title
    component_names = [u["component"] for u in updates]
    if len(component_names) == 1:
        title = f"Update {component_names[0]}"
    elif len(component_names) == 2:
        title = f"Update {component_names[0]} and {component_names[1]}"
    else:
        title = f"Update {', '.join(component_names[:-1])}, and {component_names[-1]}"

    # Build description
    description_parts = []
    description_parts.append("## Component Version Updates\n")

    for update in updates:
        description_parts.append(
            f"- **{update['component']}**: {update['old_version']} → {update['new_version']}"
        )

    # Add warnings for independent updates
    if independent_warnings:
        description_parts.append("\n## ⚠️ Independent Patch Updates\n")
        description_parts.append(
            "The following updates include patches newer than what parent components require. "
            "This should only be considered for critical CVE fixes of EoL versions. Please verify compatibility before merging:\n"
        )
        description_parts.extend(independent_warnings)

    return {"title": title, "description": "\n".join(description_parts)}


def main():
    parser = argparse.ArgumentParser(
        "update-component-versions.py", usage=USAGE, description=DESCRIPTION
    )
    parser.add_argument(
        "--dry-run",
        default=False,
        action="store_true",
        help="Don't write changes to disk",
    )
    parser.add_argument(
        "--json-output",
        default=False,
        action="store_true",
        help="Output JSON for PR creation instead of updating files",
    )
    args = parser.parse_args(sys.argv[1:])

    pr_metadata = collect_and_apply_component_updates(args.dry_run)
    if args.json_output:
        print(json.dumps(pr_metadata))
    else:
        print(pr_metadata["title"])
        print(pr_metadata["description"])


if __name__ == "__main__":
    main()
