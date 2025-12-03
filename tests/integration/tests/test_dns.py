#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.PULL_REQUEST)
def test_dns(instances: List[harness.Instance]):
    instance = instances[0]
    util.wait_until_k8s_ready(instance, [instance])
    util.wait_for_network(instance)
    util.wait_for_dns(instance)

    instance.exec(
        [
            "k8s",
            "kubectl",
            "run",
            "busybox",
            "--image=ghcr.io/containerd/busybox:1.28",
            "--restart=Never",
            "--",
            "sleep",
            "3600",
        ],
    )

    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "run=busybox",
            "--timeout",
            "180s",
        ]
    )

    result = instance.exec(
        ["k8s", "kubectl", "exec", "busybox", "--", "nslookup", "kubernetes.default"],
        capture_output=True,
    )

    assert "10.152.183.1 kubernetes.default.svc.cluster.local" in result.stdout.decode()

    result = instance.exec(
        ["k8s", "kubectl", "exec", "busybox", "--", "nslookup", "canonical.com"],
        capture_output=True,
        check=False,
    )

    assert "can't resolve" not in result.stdout.decode()

    # Assert that coredns is not using the default service account name.
    result = instance.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "-n",
            "kube-system",
            "deployment.apps/coredns",
            "-o",
            "jsonpath='{.spec.template.spec.serviceAccount}'",
        ],
        text=True,
        capture_output=True,
    )
    assert (
        "'coredns'" == result.stdout
    ), "Expected coredns serviceaccount to be 'coredns', not {result.stdout}"


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_dns_ha_rebalancing(instances: List[harness.Instance]):
    initial_node = instances[0]
    joining_cplane_node = instances[1]

    # Wait for initial cluster to be ready
    util.wait_until_k8s_ready(initial_node, [initial_node])

    # Verify initial state: all CoreDNS pods should be on the first node
    result = initial_node.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "pods",
            "-n",
            "kube-system",
            "-l",
            "k8s-app=coredns",
            "-o",
            "jsonpath='{.items[*].spec.nodeName}'",
        ],
        text=True,
        capture_output=True,
    )
    initial_nodes = set(result.stdout.replace("'", "").split())
    LOG.info(f"Initial CoreDNS pod distribution: {initial_nodes}")
    assert (
        len(initial_nodes) == 1
    ), f"Expected all CoreDNS pods on one node initially, got {initial_nodes}"

    # Join additional control plane nodes
    join_token = util.get_join_token(initial_node, joining_cplane_node)

    util.join_cluster(joining_cplane_node, join_token)

    util.wait_until_k8s_ready(initial_node, instances)

    # Wait for the DNS rebalancer controller to trigger and complete the rollout
    # Allow more time for image pulls/scheduling across new nodes.
    util.stubbornly(retries=60, delay_s=2).on(initial_node).exec(
        [
            "k8s",
            "kubectl",
            "rollout",
            "status",
            "-n",
            "kube-system",
            "deployment/coredns",
            "--timeout=60s",
        ]
    )

    # Get the actual nodes where CoreDNS pods are now running
    result = initial_node.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "pods",
            "-n",
            "kube-system",
            "-l",
            "k8s-app=coredns",
            "-o",
            "jsonpath='{.items[*].spec.nodeName}'",
        ],
        text=True,
        capture_output=True,
    )

    # Verify CoreDNS pods are now distributed across multiple nodes
    node_names = set(result.stdout.replace("'", "").split())
    LOG.info(f"Final CoreDNS pod distribution: {node_names}")

    assert (
        len(node_names) > 1
    ), f"CoreDNS pods not distributed after rebalancing: {node_names}"
