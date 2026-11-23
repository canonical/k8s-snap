#
# Copyright 2026 Canonical, Ltd.
#
"""Snap Store and Charmhub channel helpers shared by ``k8s-ci`` and CI tests."""

import json
import random
import time
from datetime import timezone
from email.utils import parsedate_to_datetime
from http.client import HTTPException
from typing import Optional
from urllib.error import HTTPError
from urllib.request import Request, urlopen

_INFO_URLS = {
    "charm": "https://api.charmhub.io/v2/charms/info/{}?fields=channel-map",
    "snap": "https://api.snapcraft.io/v2/snaps/info/{}?fields=channel-map",
}
# Store APIs occasionally drop connections or return 5xx/429; retry transient
# failures so one blip does not fail CI.
_ATTEMPTS = 5
_RETRY_STATUSES = {429, 500, 502, 503, 504}
_MAX_RETRY_DELAY = 30.0


def _parse_retry_after(retry_after: str) -> Optional[float]:
    try:
        return float(retry_after)
    except ValueError:
        pass
    try:
        parsed = parsedate_to_datetime(retry_after)
    except (TypeError, ValueError):
        return None
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed.timestamp() - time.time()


def _retry_delay(attempt: int, retry_after: Optional[str] = None) -> float:
    delay = min(
        float(2**attempt) + random.uniform(0, 1), _MAX_RETRY_DELAY
    )  # nosec B311
    if retry_after:
        requested = _parse_retry_after(retry_after)
        if requested is not None:
            delay = min(max(delay, requested), _MAX_RETRY_DELAY)
    return delay


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
    headers = {
        "Accept": "application/json",
        "Snap-Device-Series": "16",
        "User-Agent": "Mozilla/5.0",
    }
    if auth_token:
        headers["Authorization"] = f"Bearer {auth_token}"

    req = Request(_INFO_URLS[kind].format(name), headers=headers)
    for attempt in range(1, _ATTEMPTS + 1):
        retry_after = None
        try:
            with urlopen(req, timeout=20) as resp:  # nosec B310
                return json.load(resp)
        except HTTPError as e:
            try:
                if e.code not in _RETRY_STATUSES or attempt == _ATTEMPTS:
                    raise
                retry_after = e.headers.get("Retry-After") if e.headers else None
            finally:
                e.close()
        except (OSError, HTTPException, ValueError):
            if attempt == _ATTEMPTS:
                raise
        time.sleep(_retry_delay(attempt, retry_after))


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
