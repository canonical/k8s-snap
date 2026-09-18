#
# Copyright 2026 Canonical, Ltd.
#
"""Resolve dynamic test channel specifications into concrete channel lists.

The e2e-tests workflow fans out into a large parallel matrix where every job
that runs a version upgrade/downgrade or strict-interface test would otherwise
resolve 'recent <num> [flavour]' specifications independently, stampeding the
Snap Store info API (api.snapcraft.io) and triggering 429 rate limiting.

This script is run once per workflow run (in the single 'prepare' job) and
prints the resolved channel lists as KEY=VALUE lines on stdout, ready to be
appended to $GITHUB_OUTPUT. The matrix jobs then consume the already-resolved
values via needs.prepare.outputs.

If a specification does not use the 'recent' syntax it is passed through
unchanged. Must be run with the tox 'integration' environment Python so that
test_util is importable.
"""

import argparse
import logging
import sys
from pathlib import Path

sys.path.insert(0, (Path(__file__).parent / "tests").as_posix())

# NOTE: import harness before util/snap to match the conftest import order and
# avoid a partially-initialized circular import (util -> harness -> juju -> util).
from test_util import harness  # noqa: E402, F401
from test_util import snap, util  # noqa: E402

LOG = logging.getLogger(__name__)

UPGRADE_KEY = "TEST_VERSION_UPGRADE_CHANNELS"
DOWNGRADE_KEY = "TEST_VERSION_DOWNGRADE_CHANNELS"
STRICT_KEY = "TEST_STRICT_INTERFACE_CHANNELS"


def resolve_upgrade_channels(spec: str, args) -> str:
    """Mirror the 'recent' resolution in test_version_upgrades.py."""
    channels = spec.split()
    if not channels or channels[0].lower() != "recent":
        return spec
    if len(channels) != 2:
        raise ValueError(f"'recent' requires the number of releases: {spec!r}")
    resolved = snap.get_most_stable_channels(
        int(channels[1]),
        args.flavor,
        args.arch,
        min_release=args.min_release or None,
        # Include `latest/edge/<flavor>` only if this is not a release branch.
        include_latest=args.ref == util.MAIN_BRANCH,
    )
    return " ".join(resolved)


def resolve_downgrade_channels(spec: str, args) -> str:
    """Mirror the 'recent' resolution in test_version_upgrades.py."""
    channels = spec.split()
    if not channels or channels[0].lower() != "recent":
        return spec
    if len(channels) != 2:
        raise ValueError(f"'recent' requires the number of releases: {spec!r}")
    max_release = (
        args.ref.removeprefix("release-")
        if args.ref and args.ref.startswith("release-")
        else None
    )
    resolved = snap.get_most_stable_channels(
        int(channels[1]),
        args.flavor,
        args.arch,
        min_release=args.min_release or None,
        max_release=max_release,
        reverse=True,
        # Include `latest/edge/<flavor>` only if this is not a release branch.
        include_latest=args.ref == util.MAIN_BRANCH,
    )
    return " ".join(resolved)


def resolve_strict_channels(spec: str, args) -> str:
    """Mirror the 'recent' resolution in test_strict_interfaces.py."""
    channels = spec.split()
    if not channels or channels[0].lower() != "recent":
        return spec
    if len(channels) != 3:
        raise ValueError(
            f"'recent' requires the number of releases and the flavour: {spec!r}"
        )
    _, num_channels, flavour = channels
    resolved = snap.get_channels(int(num_channels), flavour, args.arch, "edge", True)
    return " ".join(resolved)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--arch", required=True, help="Snap architecture (e.g. amd64)")
    parser.add_argument(
        "--flavor", required=True, help="Snap flavor under test (e.g. classic)"
    )
    parser.add_argument(
        "--ref",
        default="",
        help="Git ref the tests run against (base ref for PRs, else ref name)",
    )
    parser.add_argument(
        "--min-release",
        default="",
        help="Minimum Kubernetes release to upgrade from (e.g. 1.32)",
    )
    parser.add_argument("--upgrade-channels", default="")
    parser.add_argument("--downgrade-channels", default="")
    parser.add_argument("--strict-channels", default="")
    args = parser.parse_args()

    # Mirror config.FLAVOR: an empty flavor means classic.
    args.flavor = args.flavor or "classic"

    logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")

    resolvers = (
        (UPGRADE_KEY, args.upgrade_channels, resolve_upgrade_channels),
        (DOWNGRADE_KEY, args.downgrade_channels, resolve_downgrade_channels),
        (STRICT_KEY, args.strict_channels, resolve_strict_channels),
    )
    for key, spec, resolve in resolvers:
        if not spec.strip():
            continue
        resolved = resolve(spec, args)
        LOG.info("Resolved %s=%r -> %r", key, spec, resolved)
        print(f"{key}={resolved}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
