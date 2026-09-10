#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
import time
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
    """
    Verify that CoreDNS replicas end up spread after a late join, and that
    dnsrebalancer does not storm (it deletes one co-located pod rather than
    restarting the whole Deployment).
    """
    initial_node = instances[0]
    joining_cplane_node = instances[1]

    util.wait_until_k8s_ready(initial_node, [initial_node])
    util.wait_for_dns(initial_node)

    rebalance_log_since = initial_node.exec(
        ["date", "--iso-8601=seconds"],
        text=True,
        capture_output=True,
    ).stdout.strip()

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
            "json",
        ],
        text=True,
        capture_output=True,
    )
    initial_pods = [
        pod
        for pod in json.loads(result.stdout)["items"]
        if not pod["metadata"].get("deletionTimestamp")
    ]
    initial_nodes = {pod["spec"].get("nodeName") for pod in initial_pods}
    assert (
        len(initial_nodes) == 1
    ), f"Expected all CoreDNS pods on one node initially, got {initial_nodes}"

    join_token = util.get_join_token(initial_node, joining_cplane_node)
    util.join_cluster(joining_cplane_node, join_token)
    util.wait_until_k8s_ready(initial_node, instances)

    def _nodes_from_pod_json(stdout: str):
        pods = [
            pod
            for pod in json.loads(stdout)["items"]
            if not pod["metadata"].get("deletionTimestamp") and pod["spec"].get("nodeName")
        ]
        return {pod["spec"]["nodeName"] for pod in pods}

    def _active_coredns_nodes():
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
                "json",
            ],
            text=True,
            capture_output=True,
        )
        return _nodes_from_pod_json(result.stdout)

    def _rebalance_trigger_count():
        log_result = initial_node.exec(
            [
                "journalctl",
                "-u",
                "snap.k8s.k8sd",
                "--since",
                rebalance_log_since,
                "--no-pager",
            ],
            text=True,
            capture_output=True,
            check=False,
        )
        return log_result.stdout.count("CoreDNS pods need rebalancing")

    def pods_distributed(result):
        nodes = _nodes_from_pod_json(result.stdout)
        if len(nodes) > 1:
            LOG.info(f"CoreDNS pods distributed across nodes: {nodes}")
            return True
        LOG.debug(f"CoreDNS pods still on {len(nodes)} node(s), waiting...")
        return False

    util.stubbornly(retries=60, delay_s=2).on(initial_node).until(
        pods_distributed
    ).exec(
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
            "json",
        ],
        text=True,
    )

    trigger_count = _rebalance_trigger_count()
    LOG.info(f"dnsrebalancer triggered {trigger_count} time(s)")
    # One deletion per attempt, with a 30s settle window. A handful of retries
    # is OK if the first reschedule still co-locates; a join-time storm is not.
    assert trigger_count <= 3, (
        "dnsrebalancer triggered too many times, indicating a rebalancing loop "
        f"(triggered {trigger_count} times)"
    )

    time.sleep(30)
    final_trigger_count = _rebalance_trigger_count()
    assert final_trigger_count == trigger_count, (
        "dnsrebalancer kept triggering after pods were spread "
        f"(went from {trigger_count} to {final_trigger_count})"
    )

    nodes = _active_coredns_nodes()
    assert (
        len(nodes) > 1
    ), f"Expected CoreDNS pods to be spread across multiple nodes, got: {nodes}"


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_dns_cluster_dns_propagates_to_late_joiners(
    instances: List[harness.Instance],
):
    """
    Regression test for issue #2516.

    When a node joins after CoreDNS has already been reconciled, the
    kube-system/k8sd-config ConfigMap is already at its steady-state contents
    including cluster-dns=<CoreDNS ClusterIP>. A bare Watch established *after*
    the ConfigMap was created does not synthesise an ADDED event for the
    pre-existing object, so on the joining node NodeConfigurationController
    would never reconcile kubelet with --cluster-dns. The fix seeds WatchConfigMap
    with a Get before starting the Watch, and additionally writes kubelet args +
    restarts kubelet locally on the node that reconciles CoreDNS. This test joins
    a control-plane and a worker node *after* DNS is ready and asserts every node
    ends up with --cluster-dns pointing at the CoreDNS service IP.
    """
    initial_node = instances[0]
    joining_cplane_node = instances[1]
    joining_worker_node = instances[2]

    # Bring up the initial control plane and wait for CoreDNS to converge so
    # the k8sd-config ConfigMap is already in its steady state with
    # cluster-dns=<IP> before either other node joins.
    util.wait_until_k8s_ready(initial_node, [initial_node])
    util.wait_for_dns(initial_node)

    # Discover the CoreDNS service ClusterIP instead of hard-coding it so the
    # assertion survives CIDR changes.
    result = initial_node.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "service",
            "-n",
            "kube-system",
            "coredns",
            "-o",
            "jsonpath={.spec.clusterIP}",
        ],
        capture_output=True,
    )
    coredns_ip = result.stdout.decode().strip()
    assert coredns_ip, "CoreDNS service should have a ClusterIP"
    LOG.info("CoreDNS ClusterIP: %s", coredns_ip)

    cplane_token = util.get_join_token(initial_node, joining_cplane_node)
    worker_token = util.get_join_token(initial_node, joining_worker_node, "--worker")
    util.join_cluster(joining_cplane_node, cplane_token)
    util.join_cluster(joining_worker_node, worker_token)

    util.wait_until_k8s_ready(initial_node, instances)

    # Every node's kubelet args file must end up with --cluster-dns set to the
    # CoreDNS ClusterIP. Before the fix, the joining nodes'
    # NodeConfigurationController would miss the pre-existing ConfigMap's
    # initial state and this arg would never be written.
    expected_arg = f'--cluster-dns="{coredns_ip}"'
    for instance in instances:
        LOG.info("Asserting --cluster-dns on node %s", instance.id)
        util.stubbornly(retries=12, delay_s=5).on(instance).until(
            lambda p: expected_arg in p.stdout.decode()
        ).exec(["cat", "/var/snap/k8s/common/args/kubelet"])

    # End-to-end sanity: schedule a pod on the late-joined worker and verify
    # its /etc/resolv.conf points at the CoreDNS ClusterIP. This is the actual
    # user-visible symptom from the issue.
    worker_name = util.hostname(joining_worker_node)
    initial_node.exec(
        [
            "k8s",
            "kubectl",
            "run",
            "busybox-late",
            "--image=ghcr.io/containerd/busybox:1.28",
            "--restart=Never",
            "--overrides",
            f'{{"spec": {{"nodeName": "{worker_name}"}}}}',
            "--",
            "sleep",
            "3600",
        ],
    )

    util.stubbornly(retries=3, delay_s=1).on(initial_node).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "busybox-late",
            "--timeout",
            "180s",
        ]
    )

    resolv = initial_node.exec(
        [
            "k8s",
            "kubectl",
            "exec",
            "busybox-late",
            "--",
            "cat",
            "/etc/resolv.conf",
        ],
        capture_output=True,
    ).stdout.decode()
    assert (
        f"nameserver {coredns_ip}" in resolv
    ), f"expected CoreDNS ClusterIP in pod resolv.conf, got: {resolv}"
