#
# Copyright 2025 Canonical, Ltd.
#
import logging
from ipaddress import IPv4Address, IPv6Address, ip_address
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-dualstack.yaml").read_text()
)
@pytest.mark.dualstack()
@pytest.mark.tags(tags.NIGHTLY)
def test_dualstack(instances: List[harness.Instance]):
    main = instances[0]
    dualstack_config = (config.MANIFESTS_DIR / "nginx-dualstack.yaml").read_text()

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
        util.stubbornly(retries=10, delay_s=1).on(main).exec(
            ["curl", address], shell=True
        )


@pytest.mark.node_count(3)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.network_type("dualstack")
@pytest.mark.tags(tags.NIGHTLY)
def test_ipv6_only_on_dualstack_infra(instances: List[harness.Instance]):
    main = instances[0]
    joining_cp = instances[1]
    joining_worker = instances[2]

    ipv6_bootstrap_config = (
        config.MANIFESTS_DIR / "bootstrap-ipv6-only.yaml"
    ).read_text()

    main.exec(
        ["k8s", "bootstrap", "--file", "-", "--address", "::/0"],
        input=str.encode(ipv6_bootstrap_config),
    )

    join_token = util.get_join_token(main, joining_cp)
    joining_cp.exec(["k8s", "join-cluster", join_token, "--address", "::/0"])

    join_token_worker = util.get_join_token(main, joining_worker, "--worker")
    joining_worker.exec(["k8s", "join-cluster", join_token_worker, "--address", "::/0"])

    # Deploy nginx with ipv6 service
    ipv6_config = (config.MANIFESTS_DIR / "nginx-ipv6-only.yaml").read_text()
    main.exec(["k8s", "kubectl", "apply", "-f", "-"], input=str.encode(ipv6_config))
    addresses = (
        util.stubbornly(retries=5, delay_s=3)
        .on(main)
        .exec(
            [
                "k8s",
                "kubectl",
                "get",
                "svc",
                "nginx-ipv6",
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
            assert False, "IPv4 address found in IPv6-only cluster"
        else:
            pytest.fail(f"Unknown IP address type: {addr}")

        # need to shell out otherwise this runs into permission errors
        util.stubbornly(retries=10, delay_s=1).on(main).exec(
            ["curl", address], shell=True
        )

    # This might take a while
    util.stubbornly(retries=config.DEFAULT_WAIT_RETRIES, delay_s=20).until(
        util.ready_nodes(main) == 3
    )
