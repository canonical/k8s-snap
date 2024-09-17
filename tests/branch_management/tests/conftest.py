#
# Copyright 2024 Canonical, Ltd.
#
from pathlib import Path
from typing import Optional

import pytest
import requests
import semver

STABLE_URL = "https://dl.k8s.io/release/stable.txt"
RELEASE_URL = "https://dl.k8s.io/release/stable-{}.{}.txt"


def _upstream_release_exists(ver: semver.Version) -> Optional[semver.Version]:
    """Return true if the major.minor release exists"""
    r = requests.get(RELEASE_URL.format(ver.major, ver.minor))
    if r.status_code == 200:
        return semver.Version.parse(r.content.decode().lstrip("v"))


def _get_max_minor(rev: semver.Version) -> semver.Version:
    """
    Get the latest minor release of the provided major.

    For example if you use 1 as major you will get back X where X gives you latest 1.XX release.
    """
    out = semver.Version(rev.major, 0, 0)
    while rev := _upstream_release_exists(rev):
        out = rev
        rev = semver.Version(rev.major, rev.minor + 1, 0)
    return out


def _previous_release(ver: semver.Version) -> semver.Version:
    if ver.minor != 0:
        ver = semver.Version(ver.major, ver.minor - 1, 0)
        ver = _upstream_release_exists(ver)
    else:
        ver = semver.Version(ver.major, 0, 0)
        ver = _get_max_minor(ver)
    return ver


@pytest.fixture(scope="session")
def stable_release() -> semver.Version:
    """Return the latest stable k8s in the release series"""
    r = requests.get(STABLE_URL)
    r.raise_for_status()
    return semver.Version.parse(r.content.decode().lstrip("v"))


@pytest.fixture(scope="session")
def current_release() -> semver.Version:
    """Return the current branch k8s version"""
    ver_file = (
        Path(__file__).parent / "../../../build-scripts/components/kubernetes/version"
    )
    version = ver_file.read_text().strip()
    return semver.Version.parse(version.lstrip("v"))


@pytest.fixture
def prior_stable_release(stable_release) -> semver.Version:
    """Return the prior release to the upstream stable"""
    return _previous_release(stable_release)


@pytest.fixture
def prior_release(current_release) -> semver.Version:
    """Return the prior release to the current release"""
    return _previous_release(current_release)
