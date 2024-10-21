#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
import re
import urllib.error
import urllib.request
from typing import List

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


def get_latest_channels(
    num_of_channels: int, flavor: str, arch: str, include_latest=True, minimum_risk=False
) -> List[str]:
    """Get an ascending list of latest channels based on the number of channels and flavour.

    e.g. get_latest_release_channels(3, "classic") -> ['1.31-classic/candidate', '1.30-classic/candidate']
    if there are less than num_of_channels available, return all available channels.
    Only the most stable risk level is returned for each major.minor version.
    By default, the `latest/edge/<flavor>` channel is included in the list.
    If the `minimum_risk` parameter is set to True, the edge channel is included in the list.
    """
    snap_info = get_snap_info()

    # Extract channel information
    channels = snap_info.get("channel-map", [])
    available_channels = [
        ch["channel"]["name"]
        for ch in channels
        if ch["channel"]["architecture"] == arch
    ]

    # Define regex pattern to match channels in the format 'major.minor-flavour'
    if flavor == "strict":
        pattern = re.compile(r"(\d+)\.(\d+)\/(" + "|".join(RISK_LEVELS) + ")")
    else:
        pattern = re.compile(
            r"(\d+)\.(\d+)-" + re.escape(flavor) + r"\/(" + "|".join(RISK_LEVELS) + ")"
        )

    # Dictionary to store the highest risk level for each major.minor
    channel_map = {}

    for channel in available_channels:
        match = pattern.match(channel)
        if match:
            major, minor, risk = match.groups()
            major_minor = (int(major), int(minor))

            # Add edge channels if minimum_risk is True
            if risk == "edge" and minimum_risk:
                channel_map[major_minor] = (channel, risk)
                continue

            # Store only the highest risk level channel for each major.minor
            if major_minor not in channel_map or RISK_LEVELS.index(
                risk
            ) < RISK_LEVELS.index(channel_map[major_minor][1]):
                channel_map[major_minor] = (channel, risk)

    # Sort channels by major and minor version in descending order
    sorted_channels = sorted(channel_map.keys(), reverse=False)

    # Prepare final channel list
    final_channels = [channel_map[mm][0] for mm in sorted_channels[:num_of_channels]]

    if include_latest:
        latest_channel = f"latest/edge/{flavor}"
        final_channels.append(latest_channel)

    return final_channels
