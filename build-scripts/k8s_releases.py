#!/usr/bin/env python3

import json
import sys
from typing import List, Optional

import requests
from packaging.version import Version

K8S_TAGS_URL = "https://api.github.com/repos/kubernetes/kubernetes/tags"


def _url_get(url: str) -> str:
    r = requests.get(url, timeout=5)
    r.raise_for_status()
    return r.text


def get_k8s_tags() -> List[str]:
    """Retrieve semantically ordered k8s releases, newest to oldest."""
    response = _url_get(K8S_TAGS_URL)
    tags_json = json.loads(response)
    if len(tags_json) == 0:
        raise Exception("No k8s tags retrieved.")
    tag_names = [tag['name'] for tag in tags_json]
    # Github already sorts the tags semantically but let's not rely on that.
    tag_names.sort(key=lambda x: Version(x), reverse=True)
    return tag_names


# k8s release naming:
# * alpha:  v{major}.{minor}.{patch}-alpha.{version}
# * beta:   v{major}.{minor}.{patch}-beta.{version}
# * rc:     v{major}.{minor}.{patch}-rc.{version}
# * stable: v{major}.{minor}.{patch}
def is_stable_release(release: str):
    return "-" not in release


def get_latest_stable() -> str:
    k8s_tags = get_k8s_tags()
    for tag in k8s_tags:
        if is_stable_release(tag):
            return tag
    raise Exception(
        "Couldn't find stable release, received tags: %s" % k8s_tags)


def get_latest_release() -> str:
    k8s_tags = get_k8s_tags()
    return k8s_tags[0]


def get_outstanding_prerelease() -> Optional[str]:
    latest_release = get_latest_release()
    if not is_stable_release(latest_release):
        return latest_release
    # The latest release is a stable release, no outstanding pre-release.
    return None


def get_obsolete_prereleases() -> List[str]:
    """Return obsolete K8s pre-releases.

    We only keep the latest pre-release if there is no corresponding stable
    release. All previous pre-releases are discarded.
    """
    k8s_tags = get_k8s_tags()
    if not is_stable_release(k8s_tags[0]):
        # Valid pre-release
        k8s_tags = k8s_tags[1:]
    # Discard all other pre-releases.
    return [tag for tag in k8s_tags if not is_stable_release(tag)]


# Rudimentary CLI that exposes these functions to shell scripts or GH actions.
if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.stderr.write(f"Usage: {sys.argv[0]} <function>\n")
        sys.exit(1)
    f = locals()[sys.argv[1]]
    out = f()
    if isinstance(out, (list, tuple)):
        for item in out:
            print(item)
    else:
        print(out or "")
