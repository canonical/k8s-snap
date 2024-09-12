#
# Copyright 2024 Canonical, Ltd.
#
from pathlib import Path

import pytest
import requests
import semver


@pytest.fixture
def upstream_release() -> semver.VersionInfo:
    """Return the latest stable k8s in the release series"""
    release_url = "https://dl.k8s.io/release/stable.txt"
    r = requests.get(release_url)
    r.raise_for_status()
    return semver.Version.parse(r.content.decode().lstrip("v"))


@pytest.fixture
def current_release() -> semver.VersionInfo:
    """Return the current branch k8s version"""
    ver_file = (
        Path(__file__).parent / "../../../build-scripts/components/kubernetes/version"
    )
    version = ver_file.read_text().strip()
    return semver.Version.parse(version.lstrip("v"))
