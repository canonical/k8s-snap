#
# Copyright 2025 Canonical, Ltd.
#
from pathlib import Path
from typing import Optional

import pytest
import util
import semver

STABLE_URL = "https://dl.k8s.io/release/stable.txt"
RELEASE_URL = "https://dl.k8s.io/release/stable-{}.{}.txt"


def _upstream_release(ver: semver.Version) -> Optional[semver.Version]:
    """Semver of the major.minor release if it exists"""
    resp = (
        util.stubbornly(retries=10, delay=6)
        .exec(
            ["curl", "-f", "-L", RELEASE_URL.format(ver.major, ver.minor)],
            text=True,
            capture_output=True,
        )
        .stdout
    )
    return semver.Version.parse(resp.lstrip("v"))


def _get_max_minor(ver: semver.Version) -> semver.Version:
    """
    Get the latest patch release based on the provided major.

    e.g. 1.<any>.<any> could yield 1.31.4 if 1.31 is the latest stable release on that maj channel
    e.g. 2.<any>.<any> could yield 2.12.1 if 2.12 is the latest stable release on that maj channel
    """
    out = semver.Version(ver.major, 0, 0)
    while ver := _upstream_release(ver):
        out, ver = ver, semver.Version(ver.major, ver.minor + 1, 0)
    return out


def _previous_release(ver: semver.Version) -> semver.Version:
    """Return the prior release version based on the provided version ignoring patch"""
    if ver.minor != 0:
        return _upstream_release(semver.Version(ver.major, ver.minor - 1, 0))
    return _get_max_minor(semver.Version(ver.major, 0, 0))


@pytest.fixture(scope="session")
def stable_release() -> semver.Version:
    """Return the latest stable k8s in the release series"""
    resp = (
        util.stubbornly(retries=10, delay=6)
        .exec(
            ["curl", "-L", STABLE_URL],
            text=True,
            capture_output=True,
        )
        .stdout
    )
    return semver.Version.parse(resp.lstrip("v"))


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
