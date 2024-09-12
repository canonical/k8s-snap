#
# Copyright 2024 Canonical, Ltd.
#
from pathlib import Path
from subprocess import check_output

import requests


def _get_max_minor(major):
    """Get the latest minor release of the provided major.
    For example if you use 1 as major you will get back X where X gives you latest 1.X release.
    """
    minor = 0
    while _upstream_release_exists(major, minor):
        minor += 1
    return minor - 1


def _upstream_release_exists(major, minor):
    """Return true if the major.minor release exists"""
    release_url = "https://dl.k8s.io/release/stable-{}.{}.txt".format(major, minor)
    r = requests.get(release_url)
    return r.status_code == 200


def _confirm_branch_exists(branch):
    cmd = f"git ls-remote --heads https://github.com/canonical/k8s-snap.git/ {branch}"
    output = check_output(cmd.split()).decode("utf-8")
    assert branch in output, f"Branch {branch} does not exist"


def _branch_flavours(branch: str = None):
    patch_dir = Path("build-scripts/patches")
    branch = "HEAD" if branch is None else branch
    cmd = f"git ls-tree --full-tree -r --name-only {branch} {patch_dir}"
    output = check_output(cmd.split()).decode("utf-8")
    patches = set(
        Path(f).relative_to(patch_dir).parents[0] for f in output.splitlines()
    )
    return [p.name for p in patches]


def _confirm_recipe(track, flavour):
    recipe = f"https://launchpad.net/~containers/k8s/+snap/k8s-snap-{track}-{flavour}"
    r = requests.get(recipe)
    return r.status_code == 200


def test_branches(upstream_release):
    """Ensures git branches exist for prior releases.

    We need to make sure the LP builders pointing to the main github branch are only pushing
    to the latest and current k8s edge snap tracks. An indication that this is not enforced is
    that we do not have a branch for the k8s release for the previous stable release. Let me
    clarify with an example.

    Assuming upstream stable k8s release is v1.12.x, there has to be a 1.11 github branch used
    by the respective LP builders for building the v1.11.y.
    """
    if upstream_release.minor != 0:
        major = upstream_release.major
        minor = upstream_release.minor - 1
    else:
        major = int(upstream_release.major) - 1
        minor = _get_max_minor(major)

    prior_branch = f"release-{major}.{minor}"
    print(f"Current stable is {upstream_release}")
    print(f"Checking {prior_branch} branch exists")
    _confirm_branch_exists(prior_branch)
    flavours = _branch_flavours(prior_branch)
    for flavour in flavours:
        prior_branch = f"autoupdate/{prior_branch}-{flavour}"
        print(f"Checking {prior_branch} branch exists")
        _confirm_branch_exists(prior_branch)


def test_launchpad_recipe(current_release):
    """Ensures the current recipes are available.

    We should ensure that a launchpad recipe exists for this release to be build with
    """
    track = f"{current_release.major}.{current_release.minor}"
    print(f"Checking {track} recipe exists")
    flavours = ["classic"] + _branch_flavours()
    recipe_exists = {flavour: _confirm_recipe(track, flavour) for flavour in flavours}
    if missing_recipes := [
        flavour for flavour, exists in recipe_exists.items() if not exists
    ]:
        assert (
            not missing_recipes
        ), f"LP Recipes do not exist for {track} {missing_recipes}"
