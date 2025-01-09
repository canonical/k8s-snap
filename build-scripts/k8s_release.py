#!/usr/bin/env python3

import argparse
import json
import logging
import os
import re
import subprocess
from typing import List, Optional

import requests
from packaging.version import Version

K8S_TAGS_URL = "https://api.github.com/repos/kubernetes/kubernetes/tags"
EXEC_TIMEOUT = 60

LOG = logging.getLogger(__name__)


def _url_get(url: str) -> str:
    r = requests.get(url, timeout=5)
    r.raise_for_status()
    return r.text


def get_k8s_tags() -> List[str]:
    """Retrieve semantically ordered k8s releases, newest to oldest."""
    response = _url_get(K8S_TAGS_URL)
    tags_json = json.loads(response)
    if len(tags_json) == 0:
        raise ValueError("No k8s tags retrieved.")
    tag_names = [tag["name"] for tag in tags_json]
    # Github already sorts the tags semantically but let's not rely on that.
    tag_names.sort(key=lambda x: Version(x), reverse=True)
    return tag_names


# k8s release naming:
# * alpha:  v{major}.{minor}.{patch}-alpha.{version}
# * beta:   v{major}.{minor}.{patch}-beta.{version}
# * rc:     v{major}.{minor}.{patch}-rc.{version}
# * stable: v{major}.{minor}.{patch}
def is_stable_release(release: str):
    return "-" not in release


def get_latest_stable() -> str:
    k8s_tags = get_k8s_tags()
    for tag in k8s_tags:
        if is_stable_release(tag):
            return tag
    raise ValueError("Couldn't find stable release, received tags: %s" % k8s_tags)


def get_latest_release() -> str:
    k8s_tags = get_k8s_tags()
    return k8s_tags[0]


def get_outstanding_prerelease() -> Optional[str]:
    latest_release = get_latest_release()
    if not is_stable_release(latest_release):
        return latest_release
    # The latest release is a stable release, no outstanding pre-release.
    return None


def get_obsolete_prereleases() -> List[str]:
    """Return obsolete K8s pre-releases.

    We only keep the latest pre-release if there is no corresponding stable
    release. All previous pre-releases are discarded.
    """
    k8s_tags = get_k8s_tags()
    if not is_stable_release(k8s_tags[0]):
        # Valid pre-release
        k8s_tags = k8s_tags[1:]
    # Discard all other pre-releases.
    return [tag for tag in k8s_tags if not is_stable_release(tag)]


def _exec(cmd: List[str], check=True, timeout=EXEC_TIMEOUT, cwd=None):
    """Run the specified command and return the stdout/stderr output as a tuple."""
    LOG.debug("Executing: %s, cwd: %s.", cmd, cwd)
    proc = subprocess.run(
        cmd, check=check, timeout=timeout, cwd=cwd, capture_output=True, text=True
    )
    return proc.stdout, proc.stderr


def _branch_exists(
    branch_name: str, remote=True, project_basedir: Optional[str] = None
):
    cmd = ["git", "branch"]
    if remote:
        cmd += ["-r"]

    stdout, stderr = _exec(cmd, cwd=project_basedir)
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


def prepare_prerelease_git_branch(project_basedir: str, remote: str = "origin"):
    prerelease = get_outstanding_prerelease()
    if not prerelease:
        LOG.info("No outstanding k8s pre-release.")
        return

    _update_prerelease_k8s_component(project_basedir, str(prerelease))

    _exec(
        ["git", "add", "./build-scripts/components/kubernetes/version"],
        cwd=project_basedir,
    )
    _exec(
        ["git", "commit", "-m", f"Update k8s version to {prerelease}"],
        cwd=project_basedir,
    )

    branch = get_prerelease_git_branch(str(prerelease))
    _exec(["git", "checkout", "-b", branch])
    _exec(["git", "push", remote, branch, "--force"])


def clean_obsolete_git_branches(project_basedir: str, remote="origin"):
    """Remove obsolete pre-release git branches."""
    latest_release = get_latest_release()
    LOG.info("Latest k8s release: %s", latest_release)

    _exec(["git", "fetch", remote], cwd=project_basedir)

    obsolete_prereleases = get_obsolete_prereleases()
    for outstanding_prerelease in obsolete_prereleases:
        branch = get_prerelease_git_branch(outstanding_prerelease)

        if _branch_exists(
            f"{remote}/{branch}", remote=True, project_basedir=project_basedir
        ):
            LOG.info("Cleaning up obsolete pre-release branch: %s", branch)
            _exec(["git", "push", remote, "--delete", branch], cwd=project_basedir)
        else:
            LOG.info("Obsolete branch not found, skpping: %s", branch)


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

    cmd = subparsers.add_parser("prepare_prerelease_git_branch")
    cmd.add_argument(
        "--project-basedir",
        dest="project_basedir",
        help="The k8s-snap project base directory.",
        default=os.getcwd(),
    )
    cmd.add_argument("--remote", dest="remote", help="Git remote.", default="origin")

    subparsers.add_parser("get_outstanding_prerelease")
    subparsers.add_parser("remove_obsolete_prereleases")

    kwargs = vars(parser.parse_args())
    f = locals()[kwargs.pop("subparser")]
    out = f(**kwargs)
    if isinstance(out, (list, tuple)):
        for item in out:
            print(item)
    else:
        print(out or "")
