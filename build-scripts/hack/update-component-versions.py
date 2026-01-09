#!/usr/bin/env python3

USAGE = "Update component versions for Canonical Kubernetes"

DESCRIPTION = """
Update the component versions that we use to build Canonical Kubernetes. This
script updates the individual `build-scripts/components/<component>/version`
files. The logic for each component is different, and is managed by configuration
options found in this script.
"""


import argparse
import logging
import sys
import yaml
from packaging.version import Version
from pathlib import Path
from typing import Callable
import util
from update_utils import update_go_version


logging.basicConfig(level=logging.INFO)

LOG = logging.getLogger(__name__)

DIR = Path(__file__).absolute().parent
COMPONENTS = DIR.parent / "components"
CHARTS = DIR.parent.parent / "k8s" / "manifests" / "charts"

# Version marker for latest Kubernetes version. Expected to be one of:
#
# - "https://dl.k8s.io/release/stable.txt"
# - "https://dl.k8s.io/release/stable-1.xx.txt"
# - "https://dl.k8s.io/release/latest-1.xx.txt" (e.g. for release candidate builds)
KUBERNETES_VERSION_MARKER = "https://dl.k8s.io/release/stable.txt"

# Containerd release branch to track. The most recent tag in the branch will be used.
CONTAINERD_RELEASE_BRANCH = "release/2.1"

# Helm release semver limit
#
# - None for main branch
# - Version("3.14") (e.g. for release candidate builds)
HELM_BRANCH, HELM_RELEASE_SEMVER = "main", None

# MetalLB Helm repository and chart version
METALLB_REPO = "https://metallb.github.io/metallb"
METALLB_CHART_VERSION = "0.15.3"


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


def pull_metallb_chart() -> None:
    LOG.info("Pulling MetalLB chart @ %s", METALLB_CHART_VERSION)
    util.helm_pull("metallb", METALLB_REPO, METALLB_CHART_VERSION, CHARTS)


def update_component_versions(dry_run: bool):
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
        existing = Path(path)
        existing_version_text = (
            existing.read_text().strip() if existing.exists() else None
        )
        upstream_version_text = version.strip()

        existing_parsed = (
            parse_version(existing_version_text) if existing_version_text else None
        )
        upstream_parsed = parse_version(upstream_version_text)

        # If both versions parse and the existing one is greater than upstream, skip update.
        if existing_parsed and upstream_parsed and existing_parsed > upstream_parsed:
            LOG.info(
                "Existing version %s is greater than upstream %s; keeping existing version",
                existing_version_text,
                upstream_version_text,
            )
            continue

        LOG.info("Update %s version to %s in %s", component, version, path)
        if not dry_run:
            Path(path).write_text(upstream_version_text + "\n")

    update_go_version(dry_run)

    for component, pull_helm_chart in [
        ("metallb", pull_metallb_chart),
    ]:
        LOG.info("Updating chart for %s", component)
        if not dry_run:
            pull_helm_chart()


def main():
    parser = argparse.ArgumentParser(
        "update_component_versions.py", usage=USAGE, description=DESCRIPTION
    )
    parser.add_argument("--dry-run", default=False, action="store_true")
    args = parser.parse_args(sys.argv[1:])

    return update_component_versions(args.dry_run)


if __name__ == "__main__":
    main()
