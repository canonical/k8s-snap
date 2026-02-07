#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
import re
from ipaddress import IPv4Address, IPv6Address, ip_address
from typing import List

import pytest
from test_util import config, harness, tags, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-dualstack.yaml").read_text()
)
@pytest.mark.dualstack()
@pytest.mark.tags(tags.NIGHTLY, tags.PROMOTE_CANDIDATE)
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
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.network_type("dualstack")
@pytest.mark.tags(tags.NIGHTLY, tags.PROMOTE_CANDIDATE)
@pytest.mark.skipif(
    config.SUBSTRATE == "multipass", reason="QUEMU does not properly support IPv6"
)
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

    util.wait_until_k8s_ready(main, instances)


@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.network_type("jumbo")
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.skipif(
    config.SUBSTRATE == "multipass", reason="Not implemented for multipass"
)
def test_jumbo(instances: List[harness.Instance]):
    cp_instance = instances[0]
    worker_instance = instances[1]

    cp_instance.exec(["k8s", "bootstrap"])

    join_token_worker = util.get_join_token(cp_instance, worker_instance, "--worker")
    worker_instance.exec(["k8s", "join-cluster", join_token_worker])

    util.wait_until_k8s_ready(cp_instance, instances)
    util.wait_for_network(cp_instance)

    util.set_node_labels(cp_instance, cp_instance.id, {"kubernetes.io/role": "master"})
    util.set_node_labels(
        cp_instance, worker_instance.id, {"kubernetes.io/role": "worker"}
    )

    manifest = MANIFESTS_DIR / "nginx-sticky-pod.yaml"
    cp_instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=manifest.read_bytes(),
    )

    util.stubbornly(retries=3, delay_s=1).on(cp_instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "nginx",
            "--timeout",
            "180s",
        ]
    )

    # make sure the netshoot pod is scheduled on the control plane node
    # while the nginx is scheduled on the worker node
    cp_instance.exec(
        [
            "k8s",
            "kubectl",
            "run",
            "netshoot",
            "--image=ghcr.io/nicolaka/netshoot:v0.14",
            "--restart=Never",
            "--overrides",
            '{"spec": {"nodeSelector": {"kubernetes.io/role": "master"}}}',
            "--",
            "sleep",
            "3600",
        ],
    )

    util.stubbornly(retries=3, delay_s=1).on(cp_instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "run=netshoot",
            "--timeout",
            "180s",
        ]
    )

    nginx_pod_ip = get_pod_ip(cp_instance, "nginx")
    # Exec into netshoot and ping nginx pod IP with 8000 byte packets without fragmentation..
    # The packets must be routed over the second nic (eth1) which has MTU of 9000
    result = cp_instance.exec(
        [
            "k8s",
            "kubectl",
            "exec",
            "netshoot",
            "--",
            "ping",
            "-M",
            "do",
            "-c",
            "50",
            "-s",
            "8000",
            f"{nginx_pod_ip}",
        ],
        capture_output=True,
    )

    # Sample output:
    # 50 packets transmitted, 40 received, 20% packet loss, time 9109ms
    assert "50 packets transmitted" in result.stdout.decode()
    packets_received = int(result.stdout.decode().split(", ")[1].split(" ")[0])
    assert (
        packets_received > 45
    ), "Expected at least 45 packets out of 50 to be received in running ping"


def get_pod_ip(instance: harness.Instance, pod_name, namespace="default"):
    result = instance.exec(
        ["k8s", "kubectl", "get", "pod", pod_name, "-n", namespace, "-o", "json"],
        capture_output=True,
        text=True,
        check=True,
    )
    pod_info = json.loads(result.stdout)
    return pod_info["status"]["podIP"]


@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.network_type("dualnic")
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.skipif(
    config.SUBSTRATE == "multipass", reason="Not implemented for multipass"
)
def test_dual_nic(instances: List[harness.Instance]):
    cp_instance = instances[0]
    worker_instance = instances[1]

    cp_instance.exec(["k8s", "bootstrap"])

    join_token_worker = util.get_join_token(cp_instance, worker_instance, "--worker")
    worker_instance.exec(["k8s", "join-cluster", join_token_worker])

    util.wait_until_k8s_ready(cp_instance, instances)
    util.wait_for_network(cp_instance)

    # set up multus-cni and wait for it to be ready
    manifest = MANIFESTS_DIR / "multus-cni-setup.yaml"
    cp_instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=manifest.read_bytes(),
    )

    util.stubbornly(retries=3, delay_s=1).on(cp_instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "name=multus",
            "--timeout",
            "180s",
            "-n",
            "kube-system",
        ]
    )

    # define a network attachment for the second nic
    manifest = MANIFESTS_DIR / "multus-network-attachment.yaml"
    cp_instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=manifest.read_bytes(),
    )

    manifest = MANIFESTS_DIR / "nginx-dual-nic.yaml"
    cp_instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=manifest.read_bytes(),
    )

    util.stubbornly(retries=3, delay_s=1).on(cp_instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "app=nginx",
            "--timeout",
            "180s",
        ]
    )

    number_of_devices = cp_instance.exec(
        [
            "k8s",
            "kubectl",
            "exec",
            "nginx",
            "--",
            "/bin/bash",
            "-c",
            "ls /sys/class/net | grep -v lo | wc -l",
        ],
        text=True,
        capture_output=True,
    )

    assert number_of_devices.stdout.strip() == "2"


@pytest.mark.node_count(1)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.network_type("fan")
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.skipif(
    config.SUBSTRATE == "multipass", reason="Not implemented for multipass"
)
def test_with_fan_networking(instances: List[harness.Instance]):
    main = instances[0]

    main.exec(["k8s", "bootstrap"])

    util.stubbornly(retries=5, delay_s=60).on(main).until(
        lambda p: "Please consider changing the Cilium tunnel port" in p.stdout.decode()
    ).exec(["snap", "logs", "k8s.k8sd"])

    main.exec(["k8s", "set", "annotations=k8sd/v1alpha1/cilium/tunnel-port=8473"])

    util.wait_until_k8s_ready(main, instances)
