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
from pathlib import Path
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
KUBERNETES_VERSION_MARKER = "https://cdn.dl.k8s.io/release/latest-1.32.txt"

# Containerd release branch to track. The most recent tag in the branch will be used.
CONTAINERD_RELEASE_BRANCH = "release/1.6"

# Helm release branch to track. The most recent tag in the branch will be used.
HELM_RELEASE_BRANCH = "release-3.14"

# Contour Helm repository and chart version
CONTOUR_HELM_REPO = "https://charts.bitnami.com/bitnami"
CONTOUR_CHART_VERSION = "17.0.4"

# MetalLB Helm repository and chart version
METALLB_REPO = "https://metallb.github.io/metallb"
METALLB_CHART_VERSION = "0.14.8"


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


def pull_contour_chart() -> None:
    LOG.info(
        "Pulling Contour Helm chart from %s with version %s",
        CONTOUR_HELM_REPO,
        CONTOUR_CHART_VERSION,
    )
    util.helm_pull("contour", CONTOUR_HELM_REPO, CONTOUR_CHART_VERSION, CHARTS)


def get_containerd_version() -> str:
    """Update containerd version using latest tag of specified branch"""
    containerd_repo = util.read_file(COMPONENTS / "containerd/repository")

    with util.git_repo(containerd_repo, CONTAINERD_RELEASE_BRANCH, shallow=False) as dir:
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
    with util.git_repo(helm_repo, HELM_RELEASE_BRANCH, shallow=False) as dir:
        return util.parse_output(["git", "describe", "--tags", "--abbrev=0"], cwd=dir)


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
        LOG.info("Update %s version to %s in %s", component, version, path)
        if not dry_run:
            Path(path).write_text(version.strip() + "\n")

    update_go_version(dry_run)

    for component, pull_helm_chart in [
        ("bitnami/contour", pull_contour_chart),
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
    go_snap = f'go/{".".join(go_version.split(".")[:2])}/stable'
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
    parser.add_argument("--dry-run", default=False, action="store_true")
    args = parser.parse_args(sys.argv[1:])

    return update_component_versions(args.dry_run)


if __name__ == "__main__":
    main()
