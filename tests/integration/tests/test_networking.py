#
# Copyright 2025 Canonical, Ltd.
#
import logging
import re
from ipaddress import IPv4Address, IPv6Address, ip_address
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-dualstack.yaml").read_text()
)
@pytest.mark.infra_network_type("Dualstack")
@pytest.mark.cluster_network_type("Dualstack")
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.skipif(
    config.SUBSTRATE == "multipass", reason="QUEMU does not properly support IPv6"
)
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

    def ips_available(p):
        # Look for --node-ip="ip1,ip2" pattern
        node_ip_match = re.search(r'--node-ip="([^"]+)"', p.stdout.decode())
        if not node_ip_match:
            LOG.warning("No --node-ip found in stdout")
            return False

        node_ip_value = node_ip_match.group(1)
        LOG.info("Found node-ip value: %s", node_ip_value)

        # Split by comma and validate IP addresses
        ips = [ip.strip() for ip in node_ip_value.split(",")]
        if len(ips) != 2:
            LOG.warning("Expected 2 IPs in node-ip, got %d: %s", len(ips), ips)
            return False

        # Check if we have one IPv4 and one IPv6
        ipv4_found = False
        ipv6_found = False

        for ip in ips:
            try:
                addr = ip_address(ip)
                if isinstance(addr, IPv4Address):
                    ipv4_found = True
                elif isinstance(addr, IPv6Address):
                    ipv6_found = True
                LOG.info("Parsed IP: %s (type: %s)", ip, type(addr).__name__)
            except ValueError as e:
                LOG.warning("Invalid IP address '%s': %s", ip, e)
                return False

        success = ipv4_found and ipv6_found
        LOG.info(
            "IP validation result - IPv4 found: %s, IPv6 found: %s, success: %s",
            ipv4_found,
            ipv6_found,
            success,
        )
        return success

    util.stubbornly(retries=10, delay_s=10).on(main).until(ips_available).exec(
        ["cat", "/var/snap/k8s/common/args/kubelet"]
    )


@pytest.mark.node_count(3)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-ipv6-only.yaml").read_text()
)
@pytest.mark.infra_network_type("Dualstack")
@pytest.mark.cluster_network_type("IPv6")
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.skipif(
    config.SUBSTRATE == "multipass", reason="QUEMU does not properly support IPv6"
)
def test_ipv6_only_on_dualstack_infra(instances: List[harness.Instance]):
    main = instances[0]
    joining_cp = instances[1]
    joining_worker = instances[2]

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

    util.wait_until_k8s_ready(main, instances)
