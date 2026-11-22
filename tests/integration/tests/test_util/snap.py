#
# Copyright 2025 Canonical, Ltd.
#
import json
import logging
import re
import urllib.error
from typing import List, Optional, Tuple

from test_util import util
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

        if version_key not in channel_map or RISK_LEVELS.index(
            risk
        ) < RISK_LEVELS.index(channel_map[version_key][1]):
            channel_map[version_key] = (channel, risk)

    # Sort channels by major and minor version
    if num_of_channels <= 0:
        sorted_versions = []
    elif reverse:
        sorted_versions = sorted(
            channel_map.keys(), key=lambda v: (v[0], v[1]), reverse=True
        )[:num_of_channels]
    else:
        sorted_versions = sorted(channel_map.keys(), key=lambda v: (v[0], v[1]))[
            -num_of_channels:
        ]

    # Extract only the channel names
    final_channels = [channel_map[v][0] for v in sorted_versions]

    if include_latest:
        latest_channel = f"latest/edge/{flavor}"
        if reverse:
            final_channels.insert(0, latest_channel)
        else:
            final_channels.append(latest_channel)
    return final_channels


def build_upgrade_channel_list(
    channels: List[str],
    flavor: str,
    arch: str,
    ref: Optional[str] = None,
    min_release: Optional[str] = None,
    snap_path: Optional[str] = None,
    local_snap_version: Optional[Tuple[int, int]] = None,
) -> List[str]:
    """Build the list of channels for version upgrades."""
    if isinstance(channels, str):
        channels = channels.strip().split()
    else:
        channels = list(channels)

    if channels and channels[0].lower() == "recent":
        if len(channels) != 2:
            raise ValueError(
                "'recent' requires the number of releases as second argument"
            )
        num_channels = int(channels[1])

        ref_name = ref.removeprefix("refs/heads/") if ref else ""
        if ref_name.startswith("release-"):
            max_release = ref_name.removeprefix("release-")
        elif local_snap_version is not None:
            max_release = f"{local_snap_version[0]}.{local_snap_version[1]}"
        else:
            max_release = None

        # Omit latest/edge when a local snap build is already under test.
        include_latest = ref_name == util.MAIN_BRANCH and not snap_path

        resolved_channels = get_most_stable_channels(
            num_of_channels=num_channels,
            flavor=flavor,
            arch=arch,
            include_latest=include_latest,
            min_release=min_release,
            max_release=max_release,
            reverse=False,
        )
    else:
        resolved_channels = list(channels)

    if snap_path and local_snap_version:
        added = False
        for i, ch in enumerate(resolved_channels):
            if "latest" in ch:
                continue
            chan_ver = major_minor(ch)
            if chan_ver is not None and local_snap_version < chan_ver:
                resolved_channels.insert(i, snap_path)
                added = True
                break

        if not added:
            resolved_channels.append(snap_path)

    return resolved_channels


def build_downgrade_channel_list(
    channels: List[str],
    flavor: str,
    arch: str,
    ref: Optional[str] = None,
    min_release: Optional[str] = None,
) -> List[str]:
    """Build the list of channels for version downgrades."""
    if isinstance(channels, str):
        channels = channels.strip().split()
    else:
        channels = list(channels)

    if channels and channels[0].lower() == "recent":
        if len(channels) != 2:
            raise ValueError(
                "'recent' requires the number of releases as second argument"
            )
        num_channels = int(channels[1])

        ref_name = ref.removeprefix("refs/heads/") if ref else ""
        max_release = (
            ref_name.removeprefix("release-")
            if ref_name.startswith("release-")
            else None
        )

        include_latest = ref_name == util.MAIN_BRANCH

        return get_most_stable_channels(
            num_of_channels=num_channels,
            flavor=flavor,
            arch=arch,
            include_latest=include_latest,
            min_release=min_release,
            max_release=max_release,
            reverse=True,
        )

    return channels


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
