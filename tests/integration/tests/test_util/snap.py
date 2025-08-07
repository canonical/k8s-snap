#
# Copyright 2025 Canonical, Ltd.
#
import json
import logging
import re
import urllib.error
import urllib.request
from typing import List, Optional

from test_util.util import major_minor

LOG = logging.getLogger(__name__)

SNAP_NAME = "k8s"

# For Snap Store API request
SNAPSTORE_INFO_API = "https://api.snapcraft.io/v2/snaps/info/"
SNAPSTORE_HEADERS = {
    "Snap-Device-Series": "16",
    "User-Agent": "Mozilla/5.0",
}
RISK_LEVELS = ["stable", "candidate", "beta", "edge"]


def get_snap_info(snap_name=SNAP_NAME):
    """Get the snap info from the Snap Store API."""
    req = urllib.request.Request(
        SNAPSTORE_INFO_API + snap_name, headers=SNAPSTORE_HEADERS
    )
    try:
        with urllib.request.urlopen(req) as response:  # nosec
            return json.loads(response.read().decode())
    except urllib.error.HTTPError as e:
        LOG.exception("HTTPError ({%s}): {%s} {%s}", req.full_url, e.code, e.reason)
        raise
    except urllib.error.URLError as e:
        LOG.exception("URLError ({%s}): {%s}", req.full_url, e.reason)
        raise


def filter_arch_and_flavor(channels: List[dict], arch: str, flavor: str) -> List[tuple]:
    """Filter available channels by architecture and match them with a given regex pattern
    for a flavor."""
    if flavor == "strict":
        pattern = re.compile(r"(\d+)\.(\d+)\/(" + "|".join(RISK_LEVELS) + ")")
    else:
        pattern = re.compile(
            r"(\d+)\.(\d+)-" + re.escape(flavor) + r"\/(" + "|".join(RISK_LEVELS) + ")"
        )

    matched_channels = []
    for ch in channels:
        if ch["channel"]["architecture"] == arch:
            channel_name = ch["channel"]["name"]
            match = pattern.match(channel_name)
            if match:
                major, minor, risk = match.groups()
                matched_channels.append((channel_name, int(major), int(minor), risk))

    return matched_channels


def get_most_stable_channels(
    num_of_channels: int,
    flavor: str,
    arch: str,
    include_latest: bool = True,
    min_release: Optional[str] = None,
    max_release: Optional[str] = None,
    reverse: bool = False,
) -> List[str]:
    """Get an ascending list of latest channels based on the number of channels
    flavour and architecture."""
    snap_info = get_snap_info()

    # Extract channel information and filter by architecture and flavor
    arch_flavor_channels = filter_arch_and_flavor(
        snap_info.get("channel-map", []), arch, flavor
    )

    # Dictionary to store the most stable channels for each version
    channel_map = {}
    for channel, major, minor, risk in arch_flavor_channels:
        version_key = (int(major), int(minor))

        if min_release:
            _min_release = major_minor(min_release)
            if _min_release and version_key < _min_release:
                continue

        if max_release is not None:
            _max_release = major_minor(max_release)
            if _max_release is not None and version_key > _max_release:
                continue

        if version_key not in channel_map and risk == RISK_LEVELS[3]:
            channel_map[version_key] = (channel, "edge")

    # Sort channels by major and minor version (ascending order)
    sorted_versions = sorted(
        channel_map.keys(), key=lambda v: (v[0], v[1]), reverse=reverse
    )

    # Extract only the channel names
    final_channels = [channel_map[v][0] for v in sorted_versions[:num_of_channels]]

    if include_latest:
        final_channels.append(f"latest/edge/{flavor}")

    return final_channels


def get_channels(
    num_of_channels: int, flavor: str, arch: str, risk_level: str, include_latest=True
) -> List[str]:
    """Get channels based on the risk level, architecture and flavour."""
    snap_info = get_snap_info()
    arch_flavor_channels = filter_arch_and_flavor(
        snap_info.get("channel-map", []), arch, flavor
    )

    matching_channels = [ch[0] for ch in arch_flavor_channels if ch[3] == risk_level]
    matching_channels = matching_channels[:num_of_channels]
    if include_latest:
        latest_channel = f"latest/edge/{flavor}"
        matching_channels.append(latest_channel)

    return matching_channels
