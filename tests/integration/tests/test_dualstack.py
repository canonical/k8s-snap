#
# Copyright 2024 Canonical, Ltd.
#
import logging
from ipaddress import IPv4Address, IPv6Address, ip_address
from pathlib import Path

import pytest
import yaml
from test_util import config, harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
def test_dualstack(h: harness.Harness, tmp_path: Path):
    snap_path = (tmp_path / "k8s.snap").as_posix()
    main = h.new_instance(dualstack=True)
    util.setup_k8s_snap(main, snap_path)

    with open(config.MANIFESTS_DIR / "bootstrap-dualstack.yaml", "r") as file:
        bootstrap_config = yaml.safe_load(file)

    main.exec(
        ["k8s", "bootstrap", "--file", "-"],
        input=str.encode(bootstrap_config),
    )
    util.wait_until_k8s_ready(main, [main])

    with open(config.MANIFESTS_DIR / "nginx-dualstack.yaml", "r") as file:
        dualstack_config = yaml.safe_load(file)

    # Deploy nginx with dualstack service
    main.exec(
        ["k8s", "kubectl", "apply", "-f", "-"], input=str.encode(dualstack_config)
    )
    addresses = (
        util.stubbornly(retries=5, delay_s=3)
        .on(main)
        .exec(
            [
                "k8s",
                "kubectl",
                "get",
                "svc",
                "nginx-dualstack",
                "-o",
                "jsonpath='{.spec.clusterIPs[*]}'",
            ],
            text=True,
            capture_output=True,
        )
        .stdout
    )

    for ip in addresses.split():
        addr = ip_address(ip.strip("'"))
        if isinstance(addr, IPv6Address):
            address = f"http://[{str(addr)}]"
        elif isinstance(addr, IPv4Address):
            address = f"http://{str(addr)}"
        else:
            pytest.fail(f"Unknown IP address type: {addr}")

        # need to shell out otherwise this runs into permission errors
        util.stubbornly(retries=3, delay_s=1).on(main).exec(
            ["curl", address], shell=True
        )
