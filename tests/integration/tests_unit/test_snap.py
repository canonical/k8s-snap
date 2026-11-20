#
# Copyright 2026 Canonical, Ltd.
#
from unittest.mock import patch

from test_util import snap


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
