#
# Copyright 2024 Canonical, Ltd.
#
import functools
import logging
import subprocess
from pathlib import Path

import requests

log = logging.getLogger(__name__)
K8S_GH_REPO = "https://github.com/canonical/k8s-snap.git/"
K8S_LP_REPO = " https://git.launchpad.net/k8s"


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
    branch = "HEAD" if branch is None else branch
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
