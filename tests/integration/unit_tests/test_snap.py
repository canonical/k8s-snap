#
# Copyright 2026 Canonical, Ltd.
#
from unittest.mock import patch

from test_util import config, snap, util


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


@patch("test_util.snap.get_snap_info")
def test_get_most_stable_channels_keeps_latest_edge_on_same_minor(mock_get_snap_info):
    mock_get_snap_info.return_value = MOCK_CHANNEL_MAP_6_TRACKS

    # Even when latest/edge reports the same K8s minor (v1.37.1), it may carry
    # newer k8sd/chart/hook commits and must not be deduplicated by version.
    channels = snap.get_most_stable_channels(
        2, "classic", "amd64", reverse=False, include_latest=True
    )
    assert channels == [
        "1.36-classic/candidate",
        "1.37-classic/candidate",
        "latest/edge/classic",
    ]


@patch("test_util.snap.get_snap_info")
def test_get_most_stable_channels_strict_uses_edge_strict_branch(mock_get_snap_info):
    mock_get_snap_info.return_value = {
        "channel-map": [
            {"channel": {"name": "1.35-classic/stable", "architecture": "amd64"}},
            {"channel": {"name": "1.35-classic/edge", "architecture": "amd64"}},
            {"channel": {"name": "1.36-classic/edge", "architecture": "amd64"}},
            {"channel": {"name": "1.36-classic/edge", "architecture": "arm64"}},
        ]
    }

    assert snap.get_most_stable_channels(2, "strict", "amd64") == [
        "1.35-classic/edge/strict",
        "1.36-classic/edge/strict",
        "latest/edge/strict",
    ]


@patch("test_util.snap.get_snap_info")
def test_build_upgrade_channel_list(mock_get_snap_info):
    mock_get_snap_info.return_value = MOCK_CHANNEL_MAP_6_TRACKS
    snap_path = "/tmp/k8s.snap"

    # Case 1: main PR (recent 2 + local 1.37 snap)
    # Keeps 1.37-classic/candidate so in-track k8sd/hook upgrades are tested,
    # and drops latest/edge/classic when testing a local snap build.
    c1 = snap.build_upgrade_channel_list(
        channels=["recent", "2"],
        variant="classic",
        arch="amd64",
        ref="main",
        snap_path=snap_path,
        local_snap_version=(1, 37),
    )
    assert c1 == [
        "1.36-classic/candidate",
        "1.37-classic/candidate",
        snap_path,
    ]

    # Case 2: Nightly latest/edge (recent 6 + downloaded 1.37 snap)
    # Keeps all 6 tracks including 1.37-classic/candidate, omits latest/edge/classic,
    # and appends snap_path.
    c2 = snap.build_upgrade_channel_list(
        channels=["recent", "6"],
        variant="classic",
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
        "1.37-classic/candidate",
        snap_path,
    ]

    # Case 3: Nightly 1.34-classic/edge (recent 6, ref="main", local 1.34 snap)
    # Caps max_release at 1.34 and keeps 1.34-classic/stable -> snap_path hop.
    c3 = snap.build_upgrade_channel_list(
        channels=["recent", "6"],
        variant="classic",
        arch="amd64",
        ref="main",
        snap_path=snap_path,
        local_snap_version=(1, 34),
    )
    assert c3 == [
        "1.32-classic/stable",
        "1.33-classic/stable",
        "1.34-classic/stable",
        snap_path,
    ]

    # Case 4: release-1.32 PR (recent 2, min_release="1.32", ref="release-1.32", local 1.32 snap)
    c4 = snap.build_upgrade_channel_list(
        channels=["recent", "2"],
        variant="classic",
        arch="amd64",
        ref="release-1.32",
        min_release="1.32",
        snap_path=snap_path,
        local_snap_version=(1, 32),
    )
    assert c4 == ["1.32-classic/stable", snap_path]

    # Case 5: Unreleased next minor (recent 2 + local 1.38 snap when store max is 1.37)
    c5 = snap.build_upgrade_channel_list(
        channels=["recent", "2"],
        variant="classic",
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

    # Case 6: Channel-only run on main without local snap includes latest/edge/classic
    c6 = snap.build_upgrade_channel_list(
        channels=["recent", "2"],
        variant="classic",
        arch="amd64",
        ref="main",
    )
    assert c6 == [
        "1.36-classic/candidate",
        "1.37-classic/candidate",
        "latest/edge/classic",
    ]

    # Case 7: Explicit channel list inserts local snap in order without deduplication
    c7 = snap.build_upgrade_channel_list(
        channels=["1.34-classic/stable", "1.36-classic/candidate"],
        variant="classic",
        arch="amd64",
        snap_path=snap_path,
        local_snap_version=(1, 35),
    )
    assert c7 == ["1.34-classic/stable", snap_path, "1.36-classic/candidate"]


@patch("test_util.snap.get_snap_info")
def test_build_downgrade_channel_list(mock_get_snap_info):
    mock_get_snap_info.return_value = MOCK_CHANNEL_MAP_6_TRACKS

    # On main, includes latest/edge/classic followed by the 2 most recent tracks descending
    assert snap.build_downgrade_channel_list(
        channels=["recent", "2"],
        variant="classic",
        arch="amd64",
        ref="main",
    ) == [
        "latest/edge/classic",
        "1.37-classic/candidate",
        "1.36-classic/candidate",
    ]

    # On release-1.35, caps at 1.35 and omits latest/edge/classic
    assert snap.build_downgrade_channel_list(
        channels=["recent", "2"],
        variant="classic",
        arch="amd64",
        ref="release-1.35",
    ) == [
        "1.35-classic/stable",
        "1.34-classic/stable",
    ]

    # Explicit channel list is preserved verbatim without store lookup
    explicit = ["1.37-classic/candidate", "1.37-classic/beta", "1.36-classic/candidate"]
    mock_get_snap_info.reset_mock()
    assert (
        snap.build_downgrade_channel_list(
            channels=explicit,
            variant="classic",
            arch="amd64",
        )
        == explicit
    )
    assert mock_get_snap_info.call_count == 0


@patch("test_util.util._find_stable_track")
def test_previous_track_from_branch(mock_find_stable, monkeypatch):
    monkeypatch.setattr(config, "VARIANT", "classic")
    mock_find_stable.side_effect = lambda maj, min, variant: (
        f"{maj}.{min}-{variant}" if variant else f"{maj}.{min}"
    )

    assert util._previous_track_from_branch("release-1.34") == "1.33-classic"
    assert util._previous_track_from_branch("refs/heads/release-1.34") == "1.33-classic"
    assert util._previous_track_from_branch("release-1.32") == "1.31-classic"

    assert util._previous_track_from_branch("backport/foo-1.34") == "1.33-classic"
    assert (
        util._previous_track_from_branch("backport-upgrade-stability-release-1.36")
        == "1.35-classic"
    )

    with patch("test_util.util._get_local_snap_version", return_value=(1, 35)):
        assert (
            util._previous_track_from_branch("stacked-fix", snap_path="/tmp/test.snap")
            == "1.34-classic"
        )

    with patch("test_util.util._get_local_snap_version", return_value=None):
        with patch("test_util.util._get_repo_version", return_value=(1, 34)):
            assert util._previous_track_from_branch("stacked-fix") == "1.33-classic"


@patch("test_util.util._find_stable_track")
def test_previous_track_stacked_base_ref(mock_find_stable, monkeypatch):
    monkeypatch.setattr(config, "VARIANT", "classic")
    mock_find_stable.side_effect = lambda maj, min, variant: (
        f"{maj}.{min}-{variant}" if variant else f"{maj}.{min}"
    )
    monkeypatch.setattr(config, "GH_BASE_REF", "backport/fix-1.34")
    monkeypatch.setattr(config, "GH_REF", None)

    assert util.previous_track("/tmp/k8s.snap") == "1.33-classic"
