#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, snap, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.no_setup()
@pytest.mark.skipif(
    not config.STRICT_INTERFACE_CHANNELS, reason="No strict channels configured"
)
@pytest.mark.tags(tags.WEEKLY)
def test_strict_interfaces(instances: List[harness.Instance], tmp_path):
    channels = config.STRICT_INTERFACE_CHANNELS
    cp = instances[0]
    current_channel = channels[0]

    if current_channel.lower() == "recent":
        if len(channels) != 3:
            pytest.fail(
                "'recent' requires the number of releases as second argument and the flavour as third argument"
            )
        _, num_channels, flavour = channels
        channels = snap.get_channels(int(num_channels), flavour, cp.arch, "edge", True)

    for channel in channels:
        util.setup_k8s_snap(cp, tmp_path, channel, connect_interfaces=False)

        # Log the current snap version on the node.
        out = cp.exec(["snap", "list", config.SNAP_NAME], capture_output=True)
        LOG.info(f"Current snap version: {out.stdout.decode().strip()}")

        check_snap_interfaces(cp, config.SNAP_NAME)

        cp.exec(["snap", "remove", config.SNAP_NAME, "--purge"])


def check_snap_interfaces(cp, snap_name):
    """Check the strict snap interfaces."""
    interfaces = [
        "docker-privileged",
        "kubernetes-support",
        "network",
        "network-bind",
        "network-control",
        "network-observe",
        "firewall-control",
        "process-control",
        "kernel-module-observe",
        "cilium-module-load",
        "mount-observe",
        "hardware-observe",
        "system-observe",
        "home",
        "opengl",
        "home-read-all",
        "login-session-observe",
        "log-observe",
    ]
    for interface in interfaces:
        cp.exec(
            [
                "snap",
                "run",
                "--shell",
                snap_name,
                "-c",
                f"snapctl is-connected {interface}",
            ],
        )
