#
# Copyright 2026 Canonical, Ltd.
#
"""GitHub API helpers shared by ``k8s-ci`` subcommands."""

import json
import subprocess
import sys
from typing import Optional


def gh_api(endpoint: str) -> Optional[dict | list]:
    """Call the GitHub REST API via ``gh api``.  Returns parsed JSON or None.

    Paginated list endpoints are merged into a single list by ``gh``.
    """
    try:
        r = subprocess.run(
            ["gh", "api", endpoint, "--paginate"],
            capture_output=True,
            text=True,
            check=True,
            timeout=30,
        )
        return json.loads(r.stdout)
    except (
        subprocess.CalledProcessError,
        json.JSONDecodeError,
        subprocess.TimeoutExpired,
    ) as e:
        print(f"  WARNING: gh api {endpoint}: {e}", file=sys.stderr)
        return None
