#
# Copyright 2026 Canonical, Ltd.
#
from email.message import Message
from http.client import RemoteDisconnected
from unittest.mock import MagicMock
from urllib.error import HTTPError, URLError

import pytest
import semver
import test_branches


def _store_payload(*channels: str) -> dict:
    entries = []
    for ch in channels:
        track, risk = ch.split("/", 1)
        entries.append(
            {
                "channel": {
                    "architecture": "amd64",
                    "name": ch,
                    "released-at": "2026-01-15T10:00:00Z",
                    "risk": risk,
                    "track": track,
                }
            }
        )
    return {"channel-map": entries}


@pytest.fixture(autouse=True)
def _clear_store_cache():
    test_branches._store_info.cache_clear()
    yield
    test_branches._store_info.cache_clear()


def test_k8s_version_skips_when_component_not_on_stable(monkeypatch):
    payload = _store_payload("1.35-classic/stable", "1.37-classic/candidate")
    monkeypatch.setattr(
        test_branches, "_get_k8s_component_version", lambda: semver.Version(1, 37, 1)
    )
    monkeypatch.setattr(
        test_branches, "_get_k8s_docs_version", lambda: semver.Version(1, 35, 0)
    )
    monkeypatch.setattr(test_branches, "query_info", lambda kind, name: payload)

    with pytest.raises(pytest.skip.Exception):
        test_branches.test_k8s_version()


def test_k8s_version_fails_when_component_is_on_stable(monkeypatch):
    payload = _store_payload("1.35-classic/stable", "1.37-classic/stable")
    monkeypatch.setattr(
        test_branches, "_get_k8s_component_version", lambda: semver.Version(1, 37, 1)
    )
    monkeypatch.setattr(
        test_branches, "_get_k8s_docs_version", lambda: semver.Version(1, 35, 0)
    )
    monkeypatch.setattr(test_branches, "query_info", lambda kind, name: payload)

    with pytest.raises(AssertionError, match="Channel mismatch"):
        test_branches.test_k8s_version()


def test_snap_version_in_files_fails_when_snap_is_on_stable(monkeypatch):
    payload = _store_payload("1.35-classic/stable", "1.37-classic/stable")
    monkeypatch.setattr(
        test_branches, "_get_k8s_component_version", lambda: semver.Version(1, 37, 1)
    )
    monkeypatch.setattr(test_branches, "query_info", lambda kind, name: payload)

    with pytest.raises(AssertionError, match="Channel mismatch"):
        test_branches.test_k8s_version_in_files(
            "docs/canonicalk8s/_parts/install.md:2",
            "sudo snap install k8s --classic --channel=1.35-classic/stable\n",
            "1.35",
        )


def test_version_checks_skip_and_cache_when_store_returns_429(monkeypatch):
    rate_limited = HTTPError(
        "https://api.snapcraft.io/v2/snaps/info/k8s?fields=channel-map",
        429,
        "Too Many Requests",
        Message(),
        None,
    )
    query_mock = MagicMock(side_effect=rate_limited)
    monkeypatch.setattr(
        test_branches, "_get_k8s_component_version", lambda: semver.Version(1, 37, 1)
    )
    monkeypatch.setattr(
        test_branches, "_get_k8s_docs_version", lambda: semver.Version(1, 35, 0)
    )
    monkeypatch.setattr(test_branches, "query_info", query_mock)

    with pytest.raises(pytest.skip.Exception):
        test_branches.test_k8s_version()
    with pytest.raises(pytest.skip.Exception):
        test_branches.test_k8s_version_in_files(
            "docs/canonicalk8s/_parts/install.md:2",
            "sudo snap install k8s --classic --channel=1.35-classic/stable\n",
            "1.35",
        )

    assert query_mock.call_count == 1


@pytest.mark.parametrize(
    "exc",
    [
        URLError("TLS EOF"),
        RemoteDisconnected("Remote end closed connection"),
    ],
)
def test_juju_version_check_skips_when_charm_store_unreachable(monkeypatch, exc):
    monkeypatch.setattr(
        test_branches, "_get_k8s_component_version", lambda: semver.Version(1, 37, 1)
    )
    monkeypatch.setattr(test_branches, "query_info", MagicMock(side_effect=exc))

    with pytest.raises(pytest.skip.Exception):
        test_branches.test_k8s_version_in_files(
            "docs/canonicalk8s/_parts/install.md:11",
            'juju deploy k8s --channel=1.35/stable --base="ubuntu@24.04"\n',
            "1.35",
        )


def test_juju_version_check_fails_when_charm_is_on_stable(monkeypatch):
    payload = _store_payload("1.35/stable", "1.37/stable")
    monkeypatch.setattr(
        test_branches, "_get_k8s_component_version", lambda: semver.Version(1, 37, 1)
    )
    monkeypatch.setattr(test_branches, "query_info", lambda kind, name: payload)

    with pytest.raises(AssertionError, match="Channel mismatch"):
        test_branches.test_k8s_version_in_files(
            "docs/canonicalk8s/_parts/install.md:11",
            'juju deploy k8s --channel=1.35/stable --base="ubuntu@24.04"\n',
            "1.35",
        )
