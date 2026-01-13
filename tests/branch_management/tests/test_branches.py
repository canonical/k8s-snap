#
# Copyright 2026 Canonical, Ltd.
#
import functools
import logging
import re
import subprocess
from pathlib import Path

import pytest
import requests
import semver
import yaml

log = logging.getLogger(__name__)
K8S_GH_REPO = "https://github.com/canonical/k8s-snap.git/"
K8S_LP_REPO = " https://git.launchpad.net/k8s"

DIR = Path(__file__).absolute().parent
PROJECT_BASE_DIR = DIR / ".." / ".." / ".."
COMPONENTS_DIR = PROJECT_BASE_DIR / "build-scripts" / "components"


def _sh(*args, **kwargs):
    default = {"text": True, "stderr": subprocess.PIPE}
    try:
        return subprocess.check_output(*args, **{**default, **kwargs})
    except subprocess.CalledProcessError as e:
        log.error("stdout: %s", e.stdout)
        log.error("stderr: %s", e.stderr)
        raise e


def _branch_flavours(branch: str = None):
    patch_dir = Path("build-scripts/patches")
    branch = "HEAD" if not branch else branch
    cmd = f"git ls-tree --full-tree -r --name-only origin/{branch} {patch_dir}"
    output = _sh(cmd.split())
    patches = set(
        Path(f).relative_to(patch_dir).parents[0] for f in output.splitlines()
    )
    return [p.name for p in patches]


@functools.lru_cache
def _confirm_branch_exists(repo, branch):
    log.info(f"Checking {branch} branch exists in {repo}")
    cmd = f"git ls-remote --heads {repo} {branch}"
    output = _sh(cmd.split())
    return branch in output


def _confirm_all_branches_exist(leader):
    assert _confirm_branch_exists(
        K8S_GH_REPO, leader
    ), f"GH Branch {leader} does not exist"
    branches = [leader]
    branches += [f"autoupdate/{leader}-{fl}" for fl in _branch_flavours(leader)]
    if missing := [b for b in branches if not _confirm_branch_exists(K8S_GH_REPO, b)]:
        assert missing, f"GH Branches do not exist {missing}"
    if missing := [b for b in branches if not _confirm_branch_exists(K8S_LP_REPO, b)]:
        assert missing, f"LP Branches do not exist {missing}"


@functools.lru_cache
def _confirm_recipe_exist(track, flavour):
    recipe = f"https://launchpad.net/~containers/k8s/+snap/k8s-snap-{track}-{flavour}"
    r = requests.get(recipe)
    return r.status_code == 200


def _confirm_all_recipes_exist(track, branch):
    log.info(f"Checking {track} recipe exists")
    assert _confirm_branch_exists(
        K8S_GH_REPO, branch
    ), f"GH Branch {branch} does not exist"
    flavours = ["classic"] + _branch_flavours(branch)
    recipes = {flavour: _confirm_recipe_exist(track, flavour) for flavour in flavours}
    if missing := [fl for fl, exists in recipes.items() if not exists]:
        assert missing, f"LP Recipes do not exist for {track} {missing}"


def test_prior_branches(prior_stable_release):
    """Ensures git branches exist for prior stable releases.

    This is to ensure that the prior release branches exist in the k8s-snap repository
    before we can proceed to build the next release. For example, if the current stable
    k8s release is v1.31.0, there must be a release-1.30 branch before updating main.
    """
    branch = f"release-{prior_stable_release.major}.{prior_stable_release.minor}"
    _confirm_all_branches_exist(branch)


def test_prior_recipes(prior_stable_release):
    """Ensures the recipes exist for prior stable releases.

    This is to ensure that the prior release recipes exist in launchpad before we can proceed
    to build the next release. For example, if the current stable k8s release is v1.31.0, there
    must be a k8s-snap-1.30-classic recipe before updating main.
    """
    track = f"{prior_stable_release.major}.{prior_stable_release.minor}"
    branch = f"release-{track}"
    _confirm_all_recipes_exist(track, branch)


def test_branches(current_release):
    """Ensures the current release has a release branch.

    This is to ensure that the current release branches exist in the k8s-snap repository
    before we can proceed to build it. For example, if the current stable
    k8s release is v1.31.0, there must be a release-1.31 branch.
    """
    branch = f"release-{current_release.major}.{current_release.minor}"
    _confirm_all_branches_exist(branch)


def test_recipes(current_release):
    """Ensures the current recipes are available.

    We should ensure that a launchpad recipes exist for this release to be build with

    This can fail when a new minor release (e.g. 1.32) is detected and its release branch
    is yet to be created from main.
    """
    track = f"{current_release.major}.{current_release.minor}"
    branch = f"release-{track}"
    _confirm_all_recipes_exist(track, branch)


def test_tip_recipes():
    """Ensures the tip recipes are available.

    We should ensure that a launchpad recipes always exist for tip to be build with
    """
    _confirm_all_recipes_exist("latest", "main")


def _get_k8s_component_version() -> semver.Version:
    k8s_version_file = COMPONENTS_DIR / "kubernetes" / "version"
    with open(k8s_version_file, "r") as f:
        version_str = f.read().strip("\n ").lstrip("v")
        return semver.Version.parse(version_str, optional_minor_and_patch=True)


def _get_k8s_docs_version() -> semver.Version:
    substitutions_file = (
        PROJECT_BASE_DIR / "docs" / "canonicalk8s" / "reuse" / "substitutions.yaml"
    )
    with open(substitutions_file, "r") as f:
        substitutions = yaml.safe_load(f.read())
        assert (
            "version" in substitutions
        ), "substitutions.yaml doesn't contain the k8s version"

        version_str = substitutions["version"].lstrip("v")
        return semver.Version.parse(version_str, optional_minor_and_patch=True)


def test_k8s_version():
    """Ensure that the k8s component version matches the one from the docs."""
    component_version = _get_k8s_component_version()
    docs_version = _get_k8s_docs_version()

    assert (
        component_version[:2] == docs_version[:2]
    ), f"Channel mismatch between {component_version=} and {docs_version=}"


def _find_k8s_channel_version(path: Path):
    """Yields (path:line_no, line, found_version) for each line containing a k8s channel version in the file."""
    rel = path.relative_to(PROJECT_BASE_DIR)
    with open(path, "r") as f:
        channel_re = r"channel[ =](\d+\.\d+)"
        for i, line in enumerate(f.readlines()):
            matches = re.findall(channel_re, line)
            for found in matches:
                yield f"{rel}:{i + 1}", line, found


def _all_version_lines():
    _files_mentioning_k8s_version = (
        PROJECT_BASE_DIR / "README.md",
        PROJECT_BASE_DIR / "docs/canonicalk8s/_parts/install.md",
    )
    for f in _files_mentioning_k8s_version:
        yield from _find_k8s_channel_version(f)


def _stable_channel_exists(channel: str) -> bool:
    url = f"https://charmhub.io/k8s?channel={channel}/stable"
    headers = {"Accept": "application/json", "Snap-Device-Series": "16"}
    r = requests.get(url, headers=headers)
    return r.status_code == 200


@pytest.mark.parametrize("path_line_no, line, found", _all_version_lines())
def test_k8s_version_in_files(path_line_no, line, found):
    """Ensure that the k8s component version is used in all relevant files."""
    ver = _get_k8s_component_version()
    ver_major_minor = f"{ver.major}.{ver.minor}"

    if "juju" in line and not _stable_channel_exists(ver_major_minor):
        # A k8s-operator release might be released weeks after a k8s-snap release.
        # In that case, we skip the test for this particular line until the channel
        # is released.
        pytest.skip(
            f"Skipping check for {ver_major_minor=} in {path_line_no} as k8s-operator "
            f"is not yet released for this version."
        )

    found = tuple(map(int, found.split(".")))
    assert len(found) == 2, f"Invalid channel {found=} in {path_line_no}"
    assert (
        ver[:2] == found[:2]
    ), f"Channel mismatch: expected={ver[:2]} but {found=} on {path_line_no}"
