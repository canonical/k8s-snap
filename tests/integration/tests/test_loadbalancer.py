#
# Copyright 2026 Canonical, Ltd.
#
import logging
import os
from enum import Enum
from pathlib import Path
from typing import List, Tuple

import pytest
from test_util import harness, tags, util
from test_util.config import MANIFESTS_DIR, SUBSTRATE

LOG = logging.getLogger(__name__)


def _iface_addr(instance: harness.Instance, iface: str) -> Tuple[str, str]:
    p = instance.exec(
        ["ip", "-o", "-4", "addr", "show", "dev", iface], capture_output=True
    )
    addr = [i for i in p.stdout.decode().split() if "/" in i][0]
    ip, cidr = addr.split("/")
    return ip, cidr


class K8sNetType(Enum):
    ipv4 = "ipv4"
    ipv6 = "ipv6"
    dualstack = "dualstack"


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
@pytest.mark.disable_k8s_bootstrapping()
# For loadbalancer communication
@pytest.mark.required_ports(80)
def test_loadbalancer_ipv4(instances: List[harness.Instance]):
    _test_loadbalancer(instances, k8s_net_type=K8sNetType.ipv4)


@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.PULL_REQUEST)
@pytest.mark.skipif(
    SUBSTRATE == "multipass", reason="QUEMU does not properly support IPv6"
)
def test_loadbalancer_ipv6_only(instances: List[harness.Instance]):
    _test_loadbalancer(instances, k8s_net_type=K8sNetType.ipv6)


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.network_type("dualstack")
@pytest.mark.skipif(
    SUBSTRATE == "multipass", reason="QUEMU does not properly support IPv6"
)
def test_loadbalancer_ipv6_dualstack(instances: List[harness.Instance]):
    _test_loadbalancer(instances, k8s_net_type=K8sNetType.dualstack)


def _test_loadbalancer(instances: List[harness.Instance], k8s_net_type: K8sNetType):
    instance = instances[0]
    tester_instance = instances[1]

    if k8s_net_type == K8sNetType.ipv6:
        bootstrap_config = (MANIFESTS_DIR / "bootstrap-ipv6-only.yaml").read_text()
        instance.exec(
            ["k8s", "bootstrap", "--file", "-", "--address", "::/0"],
            input=str.encode(bootstrap_config),
        )
    elif k8s_net_type == K8sNetType.dualstack:
        bootstrap_config = (MANIFESTS_DIR / "bootstrap-dualstack.yaml").read_text()
        instance.exec(
            ["k8s", "bootstrap", "--file", "-"],
            input=str.encode(bootstrap_config),
        )
    else:
        instance.exec(["k8s", "bootstrap"])

    lb_cidrs = []

    def get_lb_cidr(ipv6_cidr: bool):
        instance_default_ip = util.get_default_ip(instance, ipv6=ipv6_cidr)
        tester_instance_default_ip = util.get_default_ip(
            tester_instance, ipv6=ipv6_cidr
        )
        instance_default_cidr = util.get_default_cidr(instance, instance_default_ip)
        lb_cidr = util.find_suitable_cidr(
            parent_cidr=instance_default_cidr,
            excluded_ips=[instance_default_ip, tester_instance_default_ip],
            instance=instance,
        )
        return lb_cidr

    if k8s_net_type in (K8sNetType.ipv4, K8sNetType.dualstack):
        lb_cidrs.append(get_lb_cidr(ipv6_cidr=False))
    if k8s_net_type in (K8sNetType.ipv6, K8sNetType.dualstack):
        lb_cidrs.append(get_lb_cidr(ipv6_cidr=True))
    lb_cidr_str = ",".join(lb_cidrs)

    util.wait_for_network(instance)
    util.wait_for_dns(instance)

    instance.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(instance)
    instance.exec(
        [
            "k8s",
            "set",
            f"load-balancer.cidrs={lb_cidr_str}",
            "load-balancer.l2-mode=true",
        ]
    )

    manifest = MANIFESTS_DIR / "loadbalancer-test.yaml"
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    LOG.info("Waiting for nginx pod to show up...")
    util.stubbornly(retries=5, delay_s=10).on(instance).until(
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Nginx pod showed up.")

    util.stubbornly(retries=3, delay_s=5).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "run=my-nginx",
            "--timeout",
            "180s",
        ]
    )

    util.stubbornly(retries=10, delay_s=5).on(instance).until(
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "service", "-o", "json"])

    p = (
        util.stubbornly(retries=10, delay_s=5)
        .on(instance)
        .until(lambda p: len(p.stdout.decode().replace("'", "")) > 0)
        .exec(
            [
                "k8s",
                "kubectl",
                "get",
                "service",
                "my-nginx",
                "-o=jsonpath='{.status.loadBalancer.ingress[0].ip}'",
            ],
        )
    )
    service_ip = p.stdout.decode().replace("'", "")
    if ":" in service_ip:
        service_ip = "[" + service_ip + "]"

    LOG.info(f"Reaching out to service with service_ip = {service_ip}")
    util.stubbornly(retries=40, delay_s=5).on(tester_instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", service_ip])


@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.network_type("dualnic")
@pytest.mark.required_ports(80)
@pytest.mark.tags(tags.LB_DATAPLANE)
@pytest.mark.skipif(SUBSTRATE == "multipass", reason="Not implemented for multipass")
def test_loadbalancer_client_via_non_cilium_device(instances: List[harness.Instance]):
    """A LoadBalancer service must be reachable from a client whose return path
    leaves the node through the default-route interface.

    Triple-NIC layout mirroring sunbeam deployments: eth0 is OAM (default route,
    NOT a cilium device), eth1 is the internal network (cilium
    direct-routing-device) and eth2 is the public network carrying the MetalLB
    L2 announcements. With kube-proxy replacement, replies from a backend pod
    that is local to the announcing node must still carry the LoadBalancer IP
    as source when they egress an interface outside the cilium devices list
    (the default-route/OAM iface); a regression drops them on the client
    because the SYN-ACK leaves with the pod IP un-rev-NATted.

    Traffic path of the external check (the leg under test):

        host (src <external_ip>) --[dualstack-br0 L2]-->
          node eth2 (public, cilium device; cilium BPF DNAT VIP -> local pod)
          --> pod veth --> backend pod (colocated on the announcing node)
          --> pod SYN-ACK (punted to host stack; BUG: not rev-NATted here,
              carries the raw pod IP as source)
          --> node routing: client is off-subnet -> default route -> eth0
              (OAM, outside cilium devices -> no egress BPF -> source is not
              fixed up)
          --> client discards the SYN-ACK (source is not the VIP) -> timeout.

    The on-subnet control check takes the same path except the reply egresses
    eth2 (a cilium device -> egress BPF applies rev-NAT -> source is the VIP),
    which is why it succeeds.
    """
    node, tester = instances[0], instances[1]

    node_eth2_ip, node_eth2_prefix = _iface_addr(node, "eth2")
    tester_eth2_ip, _ = _iface_addr(tester, "eth2")

    lb_cidr = util.find_suitable_cidr(
        parent_cidr=f"{node_eth2_ip}/{node_eth2_prefix}",
        excluded_ips=[node_eth2_ip, tester_eth2_ip],
        instance=node,
    )

    node.exec(["k8s", "bootstrap"])
    util.wait_for_network(node)
    util.wait_for_dns(node)

    cilium_devices = os.getenv("TEST_CILIUM_DEVICES", "eth1,eth2")
    node.exec(
        [
            "k8s",
            "set",
            f"annotations=k8sd/v1alpha1/cilium/devices: {cilium_devices}\n"
            "k8sd/v1alpha1/cilium/direct-routing-device: eth1",
        ]
    )
    util.wait_until_k8s_ready(node, [node])
    util.wait_for_network(node)

    node.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(node)
    node.exec(
        [
            "k8s",
            "set",
            f"load-balancer.cidrs={lb_cidr}",
            "load-balancer.l2-mode=true",
            "load-balancer.l2-interfaces=eth2",
        ]
    )

    manifest = MANIFESTS_DIR / "loadbalancer-test.yaml"
    node.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    util.stubbornly(retries=5, delay_s=10).on(node).until(
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    util.stubbornly(retries=3, delay_s=5).on(node).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "run=my-nginx",
            "--timeout",
            "180s",
        ]
    )

    p = (
        util.stubbornly(retries=10, delay_s=5)
        .on(node)
        .until(lambda p: len(p.stdout.decode().replace("'", "")) > 0)
        .exec(
            [
                "k8s",
                "kubectl",
                "get",
                "service",
                "my-nginx",
                "-o=jsonpath='{.status.loadBalancer.ingress[0].ip}'",
            ],
        )
    )
    service_ip = p.stdout.decode().replace("'", "")

    LOG.info(f"On-public-subnet control check with service_ip = {service_ip}")
    util.stubbornly(retries=6, delay_s=5).on(tester).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", "--interface", tester_eth2_ip, "--max-time", "10", service_ip])

    for conf in ("all", "eth0", "eth1", "eth2"):
        node.exec(["sysctl", "-qw", f"net.ipv4.conf.{conf}.rp_filter=2"])

    # The external client is the host itself, bound to its default-route source
    # IP: off the LB subnet, so the node's reply egresses its default-route
    # interface (eth0, a non-cilium device), and outside LXD's managed subnets,
    # so the source is preserved on the wire (no host masquerade in transit).
    external_ip = (
        util.run(
            ["ip", "-o", "-4", "route", "get", "1.1.1.1"],
            capture_output=True,
            text=True,
            check=True,
        )
        .stdout.split(" src ")[1]
        .split()[0]
    )
    LOG.info(
        f"External check from host via {external_ip} with service_ip = {service_ip}"
    )
    util.stubbornly(retries=6, delay_s=5).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", "--interface", external_ip, "--max-time", "10", service_ip])
