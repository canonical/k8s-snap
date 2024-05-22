#!/usr/bin/env python3

USAGE = "Promote revisions for Canonical Kubernetes tracks"

DESCRIPTION = """
Promote revisions of the Canonical Kubernetes snap through the risk levels of each track.
The script only targets releases. The 'latest' track is ignored.
Each revision is promoted after being at a risk level for a certain amount of days.
The script will only promote a revision to stable if there is already another revision for this track at stable.
The first stable release for each track requires blessing from SolQA and is promoted manually.
"""

import argparse
import sys
import requests
import datetime
from dateutil import parser
import os

SNAPSTORE_API = "https://api.snapcraft.io/v2/snaps/info/"
PROMOTE_API_URL = "https://api.snapcraft.io/v2/snaps/revisions/promote"
SNAP_NAME = "k8s"
IGNORE_TRACKS = ["latest"]

# Headers for Snap Store API request
HEADERS = {
    "Snap-Device-Series": "16",
    "User-Agent": "Mozilla/5.0",
}

# Get the authorization token
AUTH_TOKEN = os.getenv("SNAPSTORE_AUTH_TOKEN")

# Headers for authenticated API requests
AUTH_HEADERS = {
    "Authorization": f"Bearer {AUTH_TOKEN}",
    "Content-Type": "application/json",
}

# The snap risk levels, used to find the next risk level for a revision.
RISK_LEVELS = ["edge", "beta", "candidate", "stable"]

# Revisions stay at a certain risk level for some days before being promoted.
DAYS_TO_STAY_IN_RISK = {"edge": 1, "beta": 3, "candidate": 5}


def get_snap_info(snap_name):
    response = requests.get(SNAPSTORE_API + snap_name, headers=HEADERS)
    response.raise_for_status()
    return response.json()


def promote_revision(revision, channel):
    payload = {
        "actions": [
            {"action": "release", "revision": revision, "channels": [channel]}
        ]
    }
    response = requests.post(PROMOTE_API_URL, headers=AUTH_HEADERS, json=payload)
    response.raise_for_status()
    print(f"Successfully promoted revision {revision} to {channel}")


def check_and_promote(snap_info, dry_run: bool):
    channels = {c["channel"]["name"]: c for c in snap_info["channel-map"]}

    for channel_info in snap_info["channel-map"]:
        channel = channel_info["channel"]
        track = channel["track"]
        risk = channel["risk"]
        next_risk = RISK_LEVELS[RISK_LEVELS.index(risk) + 1]
        revision = channel_info["revision"]

        if track in IGNORE_TRACKS:
            continue

        now = datetime.datetime.now(datetime.timezone.utc)

        if (now - parser.parse(channel["released-at"])).days > DAYS_TO_STAY_IN_RISK[
            risk
        ] and channels.get(f"{track}/{risk}", {}).get("revision") != channels.get(
            f"{track}/{next_risk}", {}
        ).get(
            "revision"
        ):
            if next_risk == "stable" and not f"{track}/stable" in channels.keys():
                # The track has not yet a stable release.
                # The first stable release requires blessing from SolQA and needs to be promoted manually.
                # Follow-up patches do not require this.
                print(
                    f"SolQA blessing required to promote first stable release for {track}. Skipping..."
                )
            else:
                print(f"Promoting {risk} to {next_risk} for track {track}")
                if not dry_run:
                    promote_revision(revision, f"{track}/{next_risk}")


def main():
    parser = argparse.ArgumentParser(
        "promote-tracks.py", usage=USAGE, description=DESCRIPTION
    )
    parser.add_argument("--dry-run", default=False, action="store_true")
    args = parser.parse_args(sys.argv[1:])

    snap_info = get_snap_info(SNAP_NAME)
    return check_and_promote(snap_info, args.dry_run)


if __name__ == "__main__":
    main()
