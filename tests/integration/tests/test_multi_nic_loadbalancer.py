#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

# Routing table and rule priority for the source based policy routing configured on the
# nodes. This mirrors what the advanced-routing charm sets up on multi-homed clusters.
POLICY_TABLE = "100"
POLICY_RULE_PRIORITY = "101"


def _address_of(instance: harness.Instance, device: str) -> str:
    """Return the IPv4 address configured on device."""
    process = instance.exec(
        ["ip", "-4", "-json", "addr", "show", "dev", device], capture_output=True
    )
    return json.loads(process.stdout.decode())[0]["addr_info"][0]["local"]


def _configure_policy_routing(node: harness.Instance, router_address: str):
    """Route replies sourced from the public network through the router.

    The default route stays on the management NIC, so a reply that is routed while its
    source is still the backend pod address cannot reach the client.
    """
    public_gateway = str(config.LXD_MULTI_NIC_PUBLIC_NETWORK_GATEWAY)
    for args in (
        [
            "route",
            "add",
            config.LXD_MULTI_NIC_PUBLIC_CIDR,
            "dev",
            "eth2",
            "table",
            POLICY_TABLE,
        ],
        [
            "route",
            "add",
            "default",
            "via",
            public_gateway,
            "dev",
            "eth2",
            "table",
            POLICY_TABLE,
        ],
        [
            "route",
            "add",
            config.LXD_MULTI_NIC_CLIENT_CIDR,
            "via",
            router_address,
            "dev",
            "eth2",
            "table",
            POLICY_TABLE,
        ],
        [
            "rule",
            "add",
            "from",
            config.LXD_MULTI_NIC_PUBLIC_CIDR,
            "table",
            POLICY_TABLE,
            "priority",
            POLICY_RULE_PRIORITY,
        ],
    ):
        node.exec(["ip", *args])


# The devices annotation must cover the interface holding the default route, otherwise
# replies served by a pod on the ingress node never get reverse-NATed.
BOOTSTRAP_CONFIG_TEMPLATE = """
cluster-config:
  network:
    enabled: true
  dns:
    enabled: true
  load-balancer:
    enabled: true
    l2-mode: true
    cidrs:
    - {lb_pool}
  annotations:
    k8sd/v1alpha1/cilium/devices: eth+
"""


@pytest.mark.node_count(3)
@pytest.mark.network_type("multinic")
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.skipif(
    config.SUBSTRATE != "lxd",
    reason="the multi-nic topology is only provisioned on the LXD substrate",
)
def test_loadbalancer_multi_nic_policy_routing(
    h: harness.Harness, instances: List[harness.Instance]
):
    """External LoadBalancer traffic must work when the backend runs on the ingress node.

    The nodes have three NICs: the default route is on the management NIC, the node
    addresses are on the cluster NIC and the LoadBalancer VIP is announced on the public
    NIC. Source based policy routing only steers replies whose source is on the public
    network towards the client.

    Cilium reverse-NATs the reply of a LoadBalancer connection served by a pod on the
    ingress node in the eBPF program of the egress device, while the kernel selects that
    device with the pod IP as source. If the default route device is not managed by
    Cilium, such replies leave the node untranslated and through the wrong NIC.
    """
    router = h.new_instance(network_type="multinicrouter", name_suffix="-router")
    client = h.new_instance(network_type="multinicclient", name_suffix="-client")

    router.exec(["sysctl", "-w", "net.ipv4.ip_forward=1"])
    router_address = _address_of(router, "eth1")
    LOG.info("Router reaches the public network on %s", router_address)

    for node in instances:
        _configure_policy_routing(node, router_address)

    # The client is off-cluster: it only reaches the public network through the router.
    util.stubbornly(retries=5, delay_s=2).on(client).exec(
        ["ping", "-c", "1", "-W", "2", _address_of(instances[0], "eth2")]
    )

    # Put the node addresses on the cluster network instead of the interface that holds
    # the default route, as on the affected deployments.
    cluster_addresses = [_address_of(node, "eth1") for node in instances]
    instances[0].exec(
        ["k8s", "bootstrap", "--address", cluster_addresses[0], "--file", "-"],
        input=BOOTSTRAP_CONFIG_TEMPLATE.format(
            lb_pool=config.LXD_MULTI_NIC_LB_POOL
        ).encode(),
    )
    util.wait_until_k8s_ready(instances[0], instances[:1])

    for node, address in zip(instances[1:], cluster_addresses[1:]):
        token = util.get_join_token(instances[0], node)
        node.exec(["k8s", "join-cluster", token, "--address", address])
    util.wait_until_k8s_ready(instances[0], instances)

    for node, address in zip(instances, cluster_addresses):
        process = instances[0].exec(
            ["k8s", "kubectl", "get", "node", node.id, "-o", "json"],
            capture_output=True,
        )
        internal_ips = [
            entry["address"]
            for entry in json.loads(process.stdout.decode())["status"]["addresses"]
            if entry["type"] == "InternalIP"
        ]
        assert address in internal_ips, (
            f"node {node.id} reports {internal_ips} instead of the cluster network "
            f"address {address}"
        )

    util.wait_for_network(instances[0])
    util.wait_for_load_balancer(instances[0])

    manifest = config.MANIFESTS_DIR / "loadbalancer-multi-nic-test.yaml"
    instances[0].exec(
        ["k8s", "kubectl", "apply", "-f", "-"], input=manifest.read_bytes()
    )
    instances[0].exec(
        [
            "k8s",
            "kubectl",
            "rollout",
            "status",
            "deploy/my-nginx-multi-nic",
            "--timeout",
            "5m",
        ]
    )

    process = instances[0].exec(
        ["k8s", "kubectl", "get", "svc", "my-nginx-multi-nic", "-o", "json"],
        capture_output=True,
    )
    vip = json.loads(process.stdout.decode())["status"]["loadBalancer"]["ingress"][0][
        "ip"
    ]
    LOG.info("LoadBalancer VIP is %s", vip)

    served_by = set()
    failures = []
    for attempt in range(45):
        process = client.exec(
            ["curl", "--silent", "--show-error", "--max-time", "5", f"http://{vip}/"],
            capture_output=True,
            check=False,
        )
        if process.returncode != 0:
            failures.append(f"attempt {attempt} failed with exit {process.returncode}")
            continue
        served_by.add(process.stdout.decode().strip())

    assert not failures, f"requests to the LoadBalancer VIP {vip} failed: {failures}"
    # Every node runs a replica, so the node announcing the VIP serves part of the
    # requests itself. Without responses from all backends the locally served reply path
    # would not have been exercised at all.
    assert len(served_by) == 3, (
        f"only {sorted(served_by)} answered, the reply path of a locally served request "
        "was not exercised"
    )
