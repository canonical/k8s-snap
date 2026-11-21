#
# Copyright 2026 Canonical, Ltd.
#
"""Snap Store and Charmhub channel helpers shared by ``k8s-ci`` and CI tests."""

from typing import Optional

import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

_INFO_URLS = {
    "charm": "https://api.charmhub.io/v2/charms/info/{}?fields=channel-map",
    "snap": "https://api.snapcraft.io/v2/snaps/info/{}?fields=channel-map",
}


def query_info(kind: str, name: str, auth_token: Optional[str] = None) -> dict:
    """
    Query the store for a charm or snap's published channels.

    Args:
        kind: Either "charm" (Charmhub) or "snap" (Snap Store).
        name: The name of the charm or snap.
        auth_token: Optional authentication token for private charms.

    Returns:
        JSON response from the store API.
    """
    headers = {"Content-Type": "application/json", "Snap-Device-Series": "16"}
    if auth_token:
        headers["Authorization"] = f"Bearer {auth_token}"

    # Store APIs occasionally drop connections or return 5xx; retry transient
    # failures so one blip does not fail CI.
    retry = Retry(
        total=3,
        backoff_factor=2,
        status_forcelist=(429, 500, 502, 503, 504),
        raise_on_status=False,
    )
    with requests.Session() as session:
        session.mount("https://", HTTPAdapter(max_retries=retry))
        resp = session.get(_INFO_URLS[kind].format(name), headers=headers, timeout=20)
    resp.raise_for_status()
    return resp.json()


def channel_available(info: dict, channel: str) -> bool:
    """
    Check if a channel is published in a store info response.

    Args:
        info: The response from ``query_info``.
        channel: The full channel name (e.g. '1.33/stable', '1.33-classic/stable').

    Returns:
        True if the channel exists, False otherwise.
    """
    return any(
        entry.get("channel", {}).get("name") == channel
        for entry in info.get("channel-map", [])
    )
