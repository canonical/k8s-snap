# Copyright 2026 Canonical, Ltd.
# See LICENSE file for licensing details.

"""Read and write patch-metadata.json.

The file is git-tracked so that every contributor starts from the correct
last_documented_sha. Writes are atomic (write to a temp file, then rename)
to avoid corruption on error.
"""

from __future__ import annotations

import json
import os
import pathlib
import tempfile
from datetime import date
from typing import Any

METADATA_DIR = pathlib.Path(__file__).parent.parent / "metadata"
STATE_FILE = METADATA_DIR / "patch-metadata.json"


def load() -> dict[str, Any]:
    """Return the full contents of patch-metadata.json, or an empty dict."""
    if not STATE_FILE.exists():
        return {}
    return json.loads(STATE_FILE.read_text())


def update(track: str, latest_sha: str) -> None:
    """Update *track* with *latest_sha* and today's date. Atomic write."""
    data = load()
    data.setdefault("tracks", {})
    data["tracks"][track] = {
        "last_documented_sha": latest_sha,
        "last_documented_date": date.today().isoformat(),
    }
    _atomic_write(STATE_FILE, json.dumps(data, indent=2) + "\n")


def _atomic_write(path: pathlib.Path, content: str) -> None:
    """Write *content* to *path* atomically via a sibling temp file."""
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, tmp = tempfile.mkstemp(dir=path.parent, prefix=".tmp-patch-metadata-")
    try:
        os.write(fd, content.encode())
        os.close(fd)
        os.replace(tmp, path)
    except Exception:
        os.close(fd)
        os.unlink(tmp)
        raise