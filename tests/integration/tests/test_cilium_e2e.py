#
# Copyright 2024 Canonical, Ltd.
#
import logging
import os
import platform
from typing import List

import pytest
from test_util import config, harness, util, tags

LOG = logging.getLogger(__name__)

ARCH = platform.machine()
CILIUM_CLI_ARCH_MAP = {"aarch64": "arm64", "x86_64": "amd64"}
CILIUM_CLI_VERSION = "v0.16.3"
CILIUM_CLI_TAR_GZ = f"https://github.com/cilium/cilium-cli/releases/download/{CILIUM_CLI_VERSION}/cilium-linux-{CILIUM_CLI_ARCH_MAP.get(ARCH)}.tar.gz"  # noqa


@pytest.mark.skipif(
    ARCH not in CILIUM_CLI_ARCH_MAP, reason=f"Platform {ARCH} not supported"
)
@pytest.mark.skipif(
    os.getenv("TEST_CILIUM_E2E") in ["false", None],
    reason="Test is known to be flaky on GitHub Actions",
)
@pytest.mark.tags(tags.WEEKLY)
def test_cilium_e2e(instances: List[harness.Instance]):
    instance = instances[0]
    instance.exec(["bash", "-c", "mkdir -p ~/.kube"])
    instance.exec(["bash", "-c", "k8s config > ~/.kube/config"])

    # Download cilium-cli
    instance.exec(["curl", "-L", CILIUM_CLI_TAR_GZ, "-o", "cilium.tar.gz"])
    instance.exec(["tar", "xvzf", "cilium.tar.gz"])
    instance.exec(["./cilium", "version", "--client"])

    instance.exec(["k8s", "status", "--wait-ready"])

    util.wait_for_dns(instance)
    util.wait_for_network(instance)

    # Run cilium e2e tests
    e2e_args = []
    if config.SUBSTRATE == "lxd":
        # NOTE(neoaggelos): disable "no-unexpected-packet-drops" on LXD as it fails:
        # [=] Test [no-unexpected-packet-drops] [1/61]
        #   [-] Scenario [no-unexpected-packet-drops/no-unexpected-packet-drops]
        #       Found unexpected packet drops:
        # {
        #   "labels": {
        #     "direction": "INGRESS",
        #     "reason": "VLAN traffic disallowed by VLAN filter"
        #   },
        #   "name": "cilium_drop_count_total",
        #   "value": 4
        # }
        e2e_args.extend(["--test", "!no-unexpected-packet-drops"])

    instance.exec(["./cilium", "connectivity", "test", *e2e_args])
