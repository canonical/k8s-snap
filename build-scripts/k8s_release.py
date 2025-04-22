#!/usr/bin/env python3

import argparse
import json
import logging
import os
import re
import subprocess
from typing import List, Optional, Dict

import requests
from packaging.version import Version

K8S_TAGS_URL = "https://api.github.com/repos/kubernetes/kubernetes/tags"
EXEC_TIMEOUT = 60

LOG = logging.getLogger(__name__)


def _url_get(url: str) -> str:
    """Make a GET request to the given URL and return the response text."""
    response = requests.get(url, timeout=5)
    response.raise_for_status()
    return response.text


def is_stable_release(release: str) -> bool:
    """Check if a Kubernetes release tag is stable (no pre-release suffix).

    Args:
        release: A Kubernetes release tag (e.g. 'v1.30.1', 'v1.30.0-alpha.1').

    Returns:
        True if the release is stable, False otherwise.
    """
    return "-" not in release


def get_k8s_tags() -> List[str]:
    """Retrieve semantically ordered Kubernetes release tags from GitHub.

    Returns:
        A list of release tag strings sorted from newest to oldest.

    Raises:
        ValueError: If no tags are retrieved.
    """
    response = _url_get(K8S_TAGS_URL)
    tags_json = json.loads(response)
    if not tags_json:
        raise ValueError("No k8s tags retrieved.")
    tag_names = [tag["name"] for tag in tags_json]
    tag_names.sort(key=lambda x: Version(x), reverse=True)
    return tag_names


def get_latest_stable() -> str:
    """Get the latest stable Kubernetes release tag.

    Returns:
        The latest stable release tag string (e.g., 'v1.30.1').

    Raises:
        ValueError: If no stable release is found.
    """
    for tag in get_k8s_tags():
        if is_stable_release(tag):
            return tag
    raise ValueError("Couldn't find a stable release.")


def get_latest_releases_by_minor() -> Dict[str, str]:
    """Map each minor Kubernetes version to its latest release tag.

    Returns:
        A dictionary mapping minor versions (e.g. '1.30') to the
        latest (pre-)release tag (e.g. 'v1.30.1').
    """
    latest_by_minor: Dict[str, str] = {}

    for tag in get_k8s_tags():
        # Strip leading 'v' if present since Version expects numeric first char
        version = Version(tag.lstrip("v"))
        key = f"{version.major}.{version.minor}"
        if key not in latest_by_minor:
            latest_by_minor[key] = tag

    return latest_by_minor


def get_outstanding_prereleases(as_git_branch: bool = False) -> List[str]:
    """Return outstanding K8s pre-releases.

    Args:
        as_git_branch: If True, return the git branch name for the pre-release.
    """
    latest_release = get_latest_releases_by_minor()
    prereleases = []
    for tag in latest_release.values():
        if not is_stable_release(tag):
            prereleases.append(tag)

    if as_git_branch:
        return [get_prerelease_git_branch(tag) for tag in prereleases]

    return prereleases


def get_obsolete_prereleases() -> List[str]:
    """Return obsolete K8s pre-releases.

    We only keep the latest pre-release(s) if there is no corresponding stable
    release. All previous pre-releases are discarded.
    """
    k8s_tags = get_k8s_tags()
    seen_stable_minors = set()
    obsolete = []

    for tag in k8s_tags:
        if is_stable_release(tag):
            version = Version(tag.lstrip("v"))
            seen_stable_minors.add((version.major, version.minor))
        else:
            version = Version(tag.lstrip("v").split("-")[0])
            if (version.major, version.minor) in seen_stable_minors:
                obsolete.append(tag)

    return obsolete


def _exec(*args, **kwargs) -> tuple[str, str]:
    """Run the specified command and return the stdout/stderr output as a tuple."""
    LOG.debug("Executing: %s, args: %s, kwargs: %s", cmd, args, kwargs)
    kwargs["text"] = True
    proc = subprocess.run(*args, **kwargs)
    return proc.stdout, proc.stderr


def _branch_exists(
    branch_name: str, remote=True, project_basedir: Optional[str] = None
):
    cmd = ["git", "branch"]
    if remote:
        cmd += ["-r"]

    stdout, _ = _exec(cmd, cwd=project_basedir)
    return branch_name in stdout


def get_prerelease_git_branch(prerelease: str):
    """Retrieve the name of the k8s-snap git branch for a given k8s pre-release."""
    prerelease_re = r"v\d+\.\d+\.\d-(?:alpha|beta|rc)\.\d+"
    if not re.match(prerelease_re, prerelease):
        raise ValueError("Unexpected k8s pre-release name: %s", prerelease)

    # Use a single branch for all pre-releases of a given risk level,
    # e.g. v1.33.0-alpha.0 -> autoupdate/v1.33.0-alpha
    branch = f"autoupdate/{prerelease}"
    return re.sub(r"(-[a-zA-Z]+)\.[0-9]+", r"\1", branch)


def _update_prerelease_k8s_component(project_basedir: str, k8s_version: str):
    if not project_basedir:
        raise ValueError("Project base directory unspecified.")
    k8s_component_path = os.path.join(
        project_basedir, "build-scripts", "components", "kubernetes", "version"
    )
    with open(k8s_component_path, "w") as f:
        f.write(k8s_version)


def prepare_prerelease_git_branches(project_basedir: str, remote: str = "origin"):
    prereleases = get_outstanding_prereleases()
    if not prereleases:
        LOG.info("No outstanding k8s pre-releases.")
        return

    for prerelease in prereleases:
        branch = get_prerelease_git_branch(str(prerelease))
        LOG.info("Preparing pre-release branch: %s", branch)

        # Reset branch to remote main
        _exec(
            ["git", "fetch", remote],
            cwd=project_basedir,
            capture_output=False,
        )
        _exec(
            ["git", "checkout", "-B", branch, f"{remote}/main"],
            cwd=project_basedir,
            capture_output=False,
        )

        # Update the k8s version and commit
        _update_prerelease_k8s_component(project_basedir, str(prerelease))
        _exec(
            ["git", "add", "./build-scripts/components/kubernetes/version"],
            cwd=project_basedir,
            capture_output=False,
        )

        # Only commit if there are actual changes
        result = _exec(
            ["git", "status", "--porcelain"],
            cwd=project_basedir,
            capture_output=True,
        )
        if result[0]:
            _exec(
                ["git", "commit", "-m", f"Update k8s version to {prerelease}"],
                cwd=project_basedir,
                capture_output=False,
            )
        else:
            LOG.info("Nothing to commit for %s", branch)

        # Force-push branch to remote
        _exec(
            ["git", "push", "-u", remote, branch, "--force"],
            cwd=project_basedir,
            capture_output=False,
        )


def clean_obsolete_git_branches(project_basedir: str, remote="origin"):
    """Remove obsolete pre-release git branches.

    All risk levels will be removed once the latest release is stable.
    """
    obsolete_prereleases = get_obsolete_prereleases()
    for prerelease in obsolete_prereleases:
        branch = get_prerelease_git_branch(prerelease)
        LOG.info("Checking for obsolete pre-release %s branch: %s", prerelease, branch)
        if _branch_exists(
            f"{remote}/{branch}", remote=True, project_basedir=project_basedir
        ):
            LOG.info("Cleaning up obsolete pre-release branch: %s", branch)
            _exec(["git", "push", remote, "--delete", branch], cwd=project_basedir)
        else:
            LOG.debug("Obsolete branch not found, skipping: %s", branch)


if __name__ == "__main__":
    logging.basicConfig(format="%(asctime)s %(message)s", level=logging.DEBUG)

    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest="subparser", required=True)

    cmd = subparsers.add_parser("clean_obsolete_git_branches")
    cmd.add_argument(
        "--project-basedir",
        dest="project_basedir",
        help="The k8s-snap project base directory.",
        default=os.getcwd(),
    )
    cmd.add_argument("--remote", dest="remote", help="Git remote.", default="origin")

    cmd = subparsers.add_parser("prepare_prerelease_git_branches")
    cmd.add_argument(
        "--project-basedir",
        dest="project_basedir",
        help="The k8s-snap project base directory.",
        default=os.getcwd(),
    )
    cmd.add_argument("--remote", dest="remote", help="Git remote.", default="origin")

    subparsers.add_parser("get_outstanding_prereleases")
    subparsers.add_parser("get_obsolete_prereleases")
    subparsers.add_parser("remove_obsolete_prereleases")

    kwargs = vars(parser.parse_args())
    f = locals()[kwargs.pop("subparser")]
    out = f(**kwargs)
    if isinstance(out, (list, tuple)):
        for item in out:
            print(item)
    else:
        print(out or "")
