#
# Copyright 2026 Canonical, Ltd.
#
from unittest.mock import patch

import pytest
from test_util import snap, tags


@pytest.mark.tags(tags.PULL_REQUEST)
@patch("test_util.snap.get_snap_info")
def test_get_most_stable_channels_order_with_latest(mock_get_snap_info):
    mock_get_snap_info.return_value = {
        "channel-map": [
            {"channel": {"name": "1.32-classic/stable", "architecture": "amd64"}},
            {"channel": {"name": "1.33-classic/stable", "architecture": "amd64"}},
            {"channel": {"name": "1.34-classic/stable", "architecture": "amd64"}},
        ]
    }

    # Ascending order (reverse=False, for upgrades) - latest should be at the end
    ascending = snap.get_most_stable_channels(
        3, "classic", "amd64", reverse=False, include_latest=True
    )
    assert ascending == [
        "1.32-classic/stable",
        "1.33-classic/stable",
        "1.34-classic/stable",
        "latest/edge/classic",
    ]

    # Descending order (reverse=True, for downgrades) - latest should be at the beginning
    descending = snap.get_most_stable_channels(
        3, "classic", "amd64", reverse=True, include_latest=True
    )
    assert descending == [
        "latest/edge/classic",
        "1.34-classic/stable",
        "1.33-classic/stable",
        "1.32-classic/stable",
    ]

    # Without latest, ascending
    assert snap.get_most_stable_channels(
        3, "classic", "amd64", reverse=False, include_latest=False
    ) == [
        "1.32-classic/stable",
        "1.33-classic/stable",
        "1.34-classic/stable",
    ]

    # Without latest, descending
    assert snap.get_most_stable_channels(
        3, "classic", "amd64", reverse=True, include_latest=False
    ) == [
        "1.34-classic/stable",
        "1.33-classic/stable",
        "1.32-classic/stable",
    ]


MOCK_CHANNEL_MAP_6_TRACKS = {
    "channel-map": [
        {
            "channel": {
                "name": "1.32-classic/stable",
                "architecture": "amd64",
                "track": "1.32-classic",
                "risk": "stable",
            },
            "revision": 5831,
            "version": "v1.32.10",
        },
        {
            "channel": {
                "name": "1.33-classic/stable",
                "architecture": "amd64",
                "track": "1.33-classic",
                "risk": "stable",
            },
            "revision": 5835,
            "version": "v1.33.8",
        },
        {
            "channel": {
                "name": "1.34-classic/stable",
                "architecture": "amd64",
                "track": "1.34-classic",
                "risk": "stable",
            },
            "revision": 5836,
            "version": "v1.34.6",
        },
        {
            "channel": {
                "name": "1.35-classic/stable",
                "architecture": "amd64",
                "track": "1.35-classic",
                "risk": "stable",
            },
            "revision": 5832,
            "version": "v1.35.4",
        },
        {
            "channel": {
                "name": "1.36-classic/candidate",
                "architecture": "amd64",
                "track": "1.36-classic",
                "risk": "candidate",
            },
            "revision": 5834,
            "version": "v1.36.2",
        },
        {
            "channel": {
                "name": "1.37-classic/candidate",
                "architecture": "amd64",
                "track": "1.37-classic",
                "risk": "candidate",
            },
            "revision": 5822,
            "version": "v1.37.1",
        },
        {
            "channel": {
                "name": "edge",
                "architecture": "amd64",
                "track": "latest",
                "risk": "edge",
            },
            "revision": 5968,
            "version": "v1.37.1",
        },
    ]
}


@pytest.mark.tags(tags.PULL_REQUEST)
@patch("test_util.snap.get_snap_info")
def test_get_most_stable_channels_recent_2(mock_get_snap_info):
    mock_get_snap_info.return_value = MOCK_CHANNEL_MAP_6_TRACKS

    # Ascending order (reverse=False, for upgrades) - should return the 2 MOST RECENT tracks
    ascending = snap.get_most_stable_channels(
        2, "classic", "amd64", reverse=False, include_latest=False
    )
    assert ascending == [
        "1.36-classic/candidate",
        "1.37-classic/candidate",
    ]

    # Descending order (reverse=True, for downgrades) - should return the 2 MOST RECENT tracks reversed
    descending = snap.get_most_stable_channels(
        2, "classic", "amd64", reverse=True, include_latest=False
    )
    assert descending == [
        "1.37-classic/candidate",
        "1.36-classic/candidate",
    ]


@pytest.mark.tags(tags.PULL_REQUEST)
@patch("test_util.snap.get_snap_info")
def test_get_most_stable_channels_dedupe_latest_edge(mock_get_snap_info):
    mock_get_snap_info.return_value = MOCK_CHANNEL_MAP_6_TRACKS

    # latest/edge has version v1.37.1, matching 1.37-classic/candidate (same minor 1.37).
    # It should be omitted to avoid redundant 1.37 <-> latest/edge hops.
    channels = snap.get_most_stable_channels(
        2, "classic", "amd64", reverse=False, include_latest=True
    )
    assert channels == [
        "1.36-classic/candidate",
        "1.37-classic/candidate",
    ]

    # When latest/edge has a newer version (e.g. v1.38.0), it SHOULD still be included
    info_with_newer_edge = {
        "channel-map": [
            *MOCK_CHANNEL_MAP_6_TRACKS["channel-map"][:-1],
            {
                "channel": {
                    "name": "edge",
                    "architecture": "amd64",
                    "track": "latest",
                    "risk": "edge",
                },
                "revision": 9999,
                "version": "v1.38.0",
            },
        ]
    }
    mock_get_snap_info.return_value = info_with_newer_edge
    channels_with_newer = snap.get_most_stable_channels(
        2, "classic", "amd64", reverse=False, include_latest=True
    )
    assert channels_with_newer == [
        "1.36-classic/candidate",
        "1.37-classic/candidate",
        "latest/edge/classic",
    ]


@pytest.mark.tags(tags.PULL_REQUEST)
def test_deduplicate_channels_by_revision():
    mock_info = {
        "channel-map": [
            {
                "channel": {
                    "name": "1.37-classic/candidate",
                    "architecture": "amd64",
                    "track": "1.37-classic",
                    "risk": "candidate",
                },
                "revision": 5822,
            },
            {
                "channel": {
                    "name": "1.37-classic/beta",
                    "architecture": "amd64",
                    "track": "1.37-classic",
                    "risk": "beta",
                },
                "revision": 5822,
            },
            {
                "channel": {
                    "name": "1.36-classic/candidate",
                    "architecture": "amd64",
                    "track": "1.36-classic",
                    "risk": "candidate",
                },
                "revision": 5834,
            },
        ]
    }

    # Collapses adjacent channels with same revision 5822
    raw = [
        "1.36-classic/candidate",
        "1.37-classic/candidate",
        "1.37-classic/beta",
    ]
    assert snap.deduplicate_channels_by_revision(raw, "amd64", mock_info) == [
        "1.36-classic/candidate",
        "1.37-classic/candidate",
    ]

    # Local snap path is kept intact
    raw_with_path = ["1.36-classic/candidate", "/tmp/k8s.snap"]
    assert snap.deduplicate_channels_by_revision(raw_with_path, "amd64", mock_info) == [
        "1.36-classic/candidate",
        "/tmp/k8s.snap",
    ]

    # channel@rev syntax collapses only if adjacent revisions match
    raw_explicit_revs = [
        "1.36-classic/candidate@5834",
        "1.36-classic/beta@5834",
        "1.37-classic/candidate@5822",
    ]
    assert snap.deduplicate_channels_by_revision(
        raw_explicit_revs, "amd64", mock_info
    ) == [
        "1.36-classic/candidate@5834",
        "1.37-classic/candidate@5822",
    ]


@pytest.mark.tags(tags.PULL_REQUEST)
@patch("test_util.snap.get_snap_info")
def test_build_upgrade_channel_list(mock_get_snap_info):
    mock_get_snap_info.return_value = MOCK_CHANNEL_MAP_6_TRACKS
    snap_path = "/tmp/k8s.snap"

    # Case 1: main PR (recent 2 + local 1.37 snap)
    # -> ["1.36-classic/candidate", snap_path], replacing same-minor 1.37-classic/candidate
    c1 = snap.build_upgrade_channel_list(
        channels=["recent", "2"],
        flavor="classic",
        arch="amd64",
        ref="main",
        snap_path=snap_path,
        local_snap_version=(1, 37),
    )
    assert c1 == ["1.36-classic/candidate", snap_path]

    # Case 2: Nightly latest/edge (recent 6 + downloaded 1.37 snap)
    # -> ["1.32-classic/stable", ..., "1.36-classic/candidate", snap_path]
    # No duplicate latest/edge + snap_path
    c2 = snap.build_upgrade_channel_list(
        channels=["recent", "6"],
        flavor="classic",
        arch="amd64",
        ref="main",
        snap_path=snap_path,
        local_snap_version=(1, 37),
    )
    assert c2 == [
        "1.32-classic/stable",
        "1.33-classic/stable",
        "1.34-classic/stable",
        "1.35-classic/stable",
        "1.36-classic/candidate",
        snap_path,
    ]

    # Case 3: Nightly 1.34-classic/edge (recent 6, ref="main", local 1.34 snap)
    # -> caps max_release at 1.34 -> ["1.32-classic/stable", "1.33-classic/stable", snap_path]
    c3 = snap.build_upgrade_channel_list(
        channels=["recent", "6"],
        flavor="classic",
        arch="amd64",
        ref="main",
        snap_path=snap_path,
        local_snap_version=(1, 34),
    )
    assert c3 == [
        "1.32-classic/stable",
        "1.33-classic/stable",
        snap_path,
    ]

    # Case 4: release-1.32 PR (recent 2, min_release="1.32", ref="release-1.32", local 1.32 snap)
    # -> since only 1 store track exists, appends snap_path -> ["1.32-classic/stable", snap_path]
    c4 = snap.build_upgrade_channel_list(
        channels=["recent", "2"],
        flavor="classic",
        arch="amd64",
        ref="release-1.32",
        min_release="1.32",
        snap_path=snap_path,
        local_snap_version=(1, 32),
    )
    assert c4 == ["1.32-classic/stable", snap_path]

    # Case 5: Unreleased next minor (recent 2 + local 1.38 snap when store max is 1.37)
    # -> ["1.36-classic/candidate", "1.37-classic/candidate", snap_path]
    c5 = snap.build_upgrade_channel_list(
        channels=["recent", "2"],
        flavor="classic",
        arch="amd64",
        ref="main",
        snap_path=snap_path,
        local_snap_version=(1, 38),
    )
    assert c5 == [
        "1.36-classic/candidate",
        "1.37-classic/candidate",
        snap_path,
    ]
