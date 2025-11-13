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
CONTAINERD_RELEASE_BRANCH = "release/1.7"

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


def get_kubernetes_version() -> str:
    """Update Kubernetes version based on the specified marker file"""
    LOG.info("Checking latest Kubernetes version from %s", KUBERNETES_VERSION_MARKER)
    return util.read_url(KUBERNETES_VERSION_MARKER)


def get_cni_version() -> str:
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


def get_runc_version() -> str:
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
                capture_output=True
            )
            tags = subprocess.run(
                ["git", "tag"],
                cwd=tmpdir,
                check=True,
                capture_output=True
            ).stdout.decode().strip().split()
            
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


def check_independent_update(
    dependency: str,
    current_version: str,
    parent_required_version: str,
    upstream_latest_version: Optional[str]
) -> tuple[Optional[str], bool, str]:
    """Check if dependency has an independent update available.
    
    Returns:
        tuple of (new_version, is_independent_update, description)
    """
    if not upstream_latest_version:
        return None, False, ""
    
    current = parse_version(current_version)
    parent_required = parse_version(parent_required_version)
    upstream_latest = parse_version(upstream_latest_version)
    
    if not current or not upstream_latest:
        return None, False, ""
    
    # Skip if upstream is older or same as current
    if upstream_latest <= current:
        return None, False, ""
    
    # Check if this is an independent update (diverges from parent requirement)
    is_independent = parent_required and upstream_latest > parent_required
    
    description = ""
    if is_independent:
        description = f"""⚠️ EoL Dependency Patch Update

The following patch version bump deviates from parent requirements:
- {dependency}: {current_version} → {upstream_latest_version}

The parent does not yet require this version. Please verify compatibility and approve manually."""
    
    return upstream_latest_version, is_independent, description


def pull_metallb_chart() -> None:
    LOG.info("Pulling MetalLB chart @ %s", METALLB_CHART_VERSION)
    util.helm_pull("metallb", METALLB_REPO, METALLB_CHART_VERSION, CHARTS)


def collect_component_updates() -> list[dict]:
    """Collect all component updates and return JSON-ready data structure.
    
    Returns a list of update entries, where each entry contains:
    - dependency: component name
    - current_version: version currently in use
    - new_version: proposed new version
    - upstream_parent_version: version required by parent (if applicable)
    - upstream_latest_version: latest upstream version
    - independent_update: True if update diverges from parent requirement
    - title: PR title
    - description: PR description
    """
    updates = []
    
    # Check runc for independent updates
    try:
        LOG.info("Checking runc for independent updates")
        current_runc = util.read_file(COMPONENTS / "runc/version")
        parent_required_runc = get_runc_version()  # What containerd requires
        runc_repo = util.read_file(COMPONENTS / "runc/repository")
        upstream_runc = get_latest_upstream_version(runc_repo, current_runc)
        
        new_version, is_independent, description = check_independent_update(
            "runc",
            current_runc,
            parent_required_runc,
            upstream_runc
        )
        
        if new_version:
            containerd_version = util.read_file(COMPONENTS / "containerd/version")
            if is_independent:
                title = f"Update runc to {new_version} (independent patch release)"
                description = f"""⚠️ EoL Dependency Patch Update

The following patch version bump deviates from parent requirements:
- runc: {current_runc} → {new_version}

The parent (containerd {containerd_version}) does not yet require this version. Please verify compatibility and approve manually."""
            else:
                title = f"Update runc to {new_version}"
                description = f"Update runc from {current_runc} to {new_version}"
            
            updates.append({
                "dependency": "runc",
                "current_version": current_runc,
                "new_version": new_version,
                "upstream_parent_version": parent_required_runc,
                "upstream_latest_version": upstream_runc or new_version,
                "independent_update": is_independent,
                "title": title,
                "description": description
            })
            LOG.info("Found runc update: %s -> %s (independent: %s)", 
                    current_runc, new_version, is_independent)
    except Exception as e:
        LOG.warning("Failed to check runc updates: %s", e)
    
    return updates


def update_component_versions(dry_run: bool, json_output: bool = False):
    """Update component versions or output JSON for PR creation.
    
    Args:
        dry_run: If True, don't write changes to disk
        json_output: If True, output JSON instead of updating files
    """
    if json_output:
        # Generate JSON output for workflow to create PRs
        updates = collect_component_updates()
        print(json.dumps(updates, indent=2))
        return
    
    # Original behavior: update files directly
    for component, get_version in [
        ("kubernetes", get_kubernetes_version),
        ("cni", get_cni_version),
        ("containerd", get_containerd_version),
        ("runc", get_runc_version),
        ("helm", get_helm_version),
    ]:
        LOG.info("Updating version for %s", component)
        version: str = get_version()
        path = COMPONENTS / component / "version"
        LOG.info("Update %s version to %s in %s", component, version, path)
        if not dry_run:
            Path(path).write_text(version.strip() + "\n")

    update_go_version(dry_run)

    for component, pull_helm_chart in [
        ("metallb", pull_metallb_chart),
    ]:
        LOG.info("Updating chart for %s", component)
        if not dry_run:
            pull_helm_chart()


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


def main():
    parser = argparse.ArgumentParser(
        "update-component-versions.py", usage=USAGE, description=DESCRIPTION
    )
    parser.add_argument("--dry-run", default=False, action="store_true",
                       help="Don't write changes to disk")
    parser.add_argument("--json-output", default=False, action="store_true",
                       help="Output JSON for PR creation instead of updating files")
    args = parser.parse_args(sys.argv[1:])

    return update_component_versions(args.dry_run, args.json_output)


if __name__ == "__main__":
    main()
