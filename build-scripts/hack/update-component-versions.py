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


def pull_metallb_chart() -> None:
    LOG.info("Pulling MetalLB chart @ %s", METALLB_CHART_VERSION)
    util.helm_pull("metallb", METALLB_REPO, METALLB_CHART_VERSION, CHARTS)


def collect_and_apply_component_updates(dry_run: bool) -> dict:
    """Collect component updates, apply them to version files, and return PR metadata.
    
    Returns a dict with:
    - title: PR title listing all updated components
    - description: PR description with version bumps and warnings
    """
    updates = []
    independent_warnings = []
    
    # Check all components for updates
    components_to_check = [
        ("kubernetes", get_kubernetes_version, None, None),
        ("cni", get_cni_version, None, None),
        ("containerd", get_containerd_version, None, None),
        ("runc", get_runc_version, "containerd", None),
        ("helm", get_helm_version, None, None),
    ]
    
    for component, get_version_func, parent_component, check_upstream in components_to_check:
        try:
            current_version = util.read_file(COMPONENTS / component / "version")
            new_version = get_version_func()
            
            # Check if there's an update needed
            if current_version.strip() != new_version.strip():
                LOG.info("Found %s update: %s -> %s", component, current_version.strip(), new_version.strip())
                
                # For runc, check if this is an independent update
                is_independent = False
                if component == "runc":
                    try:
                        runc_repo = util.read_file(COMPONENTS / "runc/repository")
                        upstream_runc = get_latest_upstream_version(runc_repo, current_version.strip())
                        parent_required = new_version  # This is what containerd requires
                        
                        if upstream_runc:
                            upstream_parsed = parse_version(upstream_runc)
                            parent_parsed = parse_version(parent_required)
                            
                            if upstream_parsed and parent_parsed and upstream_parsed > parent_parsed:
                                is_independent = True
                                new_version = upstream_runc
                                containerd_version = util.read_file(COMPONENTS / "containerd/version").strip()
                                independent_warnings.append(
                                    f"- **{component}**: {current_version.strip()} → {new_version} "
                                    f"(upstream has newer patches than parent {parent_component} {containerd_version} requires)"
                                )
                                LOG.info("Independent update detected for %s: upstream=%s, parent requires=%s",
                                        component, upstream_runc, parent_required)
                    except Exception as e:
                        LOG.warning("Failed to check upstream version for %s: %s", component, e)
                
                updates.append({
                    "component": component,
                    "old_version": current_version.strip(),
                    "new_version": new_version.strip(),
                    "independent": is_independent
                })
                
                # Apply the update
                if not dry_run:
                    Path(COMPONENTS / component / "version").write_text(new_version.strip() + "\n")
            else:
                LOG.info("No update needed for %s (current: %s)", component, current_version.strip())
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
        description_parts.append(f"- **{update['component']}**: {update['old_version']} → {update['new_version']}")
    
    # Add warnings for independent updates
    if independent_warnings:
        description_parts.append("\n## ⚠️ Independent Patch Updates\n")
        description_parts.append("The following updates include patches newer than what parent components require. "
                                "Please verify compatibility before merging:\n")
        description_parts.extend(independent_warnings)
    
    return {
        "title": title,
        "description": "\n".join(description_parts)
    }


def update_component_versions(dry_run: bool, json_output: bool = False):
    """Update component versions or output JSON for PR creation.
    
    Args:
        dry_run: If True, don't write changes to disk
        json_output: If True, output JSON with PR metadata instead of updating files directly
    """
    if json_output:
        # Check for updates, apply them, and generate PR metadata
        pr_metadata = collect_and_apply_component_updates(dry_run)
        
        if pr_metadata["title"]:
            LOG.info("Updates found: %s", pr_metadata["title"])
        else:
            LOG.info("No updates found")
        
        print(json.dumps(pr_metadata, indent=2))
        return
    
    # Original behavior: update files directly without JSON output
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
