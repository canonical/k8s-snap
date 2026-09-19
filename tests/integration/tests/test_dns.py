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
    Verify that a late join spreads CoreDNS without a rollout restart storm.
    """
    initial_node = instances[0]
    joining_cplane_node = instances[1]

    util.wait_until_k8s_ready(initial_node, [initial_node])
    util.wait_for_dns(initial_node)

    get_coredns = [
        "k8s",
        "kubectl",
        "get",
        "deployments,pods",
        "-n",
        "kube-system",
        "-l",
        "k8s-app=coredns",
        "-o",
        "json",
    ]

    def _settled_pods(result):
        items = json.loads(result.stdout)["items"]
        deployments = [item for item in items if item["kind"] == "Deployment"]
        pods = [item for item in items if item["kind"] == "Pod"]
        if len(deployments) != 1:
            return []
        deployment = deployments[0]
        desired = deployment["spec"].get("replicas", 1)
        status = deployment.get("status", {})
        if (
            desired < 2
            or status.get("observedGeneration", 0)
            < deployment["metadata"]["generation"]
            or any(
                status.get(field, 0) != desired
                for field in (
                    "replicas",
                    "updatedReplicas",
                    "readyReplicas",
                    "availableReplicas",
                )
            )
            or len(pods) != desired
        ):
            return []
        for pod in pods:
            if (
                pod["metadata"].get("deletionTimestamp")
                or not pod["spec"].get("nodeName")
                or not any(
                    condition["type"] == "Ready" and condition["status"] == "True"
                    for condition in pod.get("status", {}).get("conditions", [])
                )
            ):
                return []
        hashes = {pod["metadata"]["labels"].get("pod-template-hash") for pod in pods}
        return pods if len(hashes) == 1 and None not in hashes else []

    result = (
        util.stubbornly(retries=36, delay_s=5)
        .on(initial_node)
        .until(_settled_pods)
        .exec(get_coredns, text=True)
    )
    initial_pods = _settled_pods(result)
    initial_hash = initial_pods[0]["metadata"]["labels"]["pod-template-hash"]
    initial_nodes = {pod["spec"]["nodeName"] for pod in initial_pods}
    assert (
        len(initial_nodes) == 1
    ), f"Expected all CoreDNS pods on one node initially, got {initial_nodes}"

    rebalance_log_since = initial_node.exec(
        ["date", "+@%s"],
        text=True,
        capture_output=True,
    ).stdout.strip()

    join_token = util.get_join_token(initial_node, joining_cplane_node)
    util.join_cluster(joining_cplane_node, join_token)
    util.wait_until_k8s_ready(initial_node, instances)

    def _rebalance_trigger_count():
        count = 0
        for instance in instances:
            log_result = instance.exec(
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
            )
            count += log_result.stdout.count("CoreDNS pods need rebalancing")
        return count

    def _is_rebalanced(result):
        pods = _settled_pods(result)
        return (
            bool(pods)
            and pods[0]["metadata"]["labels"]["pod-template-hash"] != initial_hash
            and len({pod["spec"]["nodeName"] for pod in pods}) > 1
        )

    result = (
        util.stubbornly(retries=60, delay_s=5)
        .on(initial_node)
        .until(_is_rebalanced)
        .exec(get_coredns, text=True)
    )
    settled_pods = _settled_pods(result)
    settled_uids = {pod["metadata"]["uid"] for pod in settled_pods}
    LOG.info(
        "CoreDNS pod placement after late join: %s",
        {pod["spec"]["nodeName"] for pod in settled_pods},
    )
    util.wait_for_dns(initial_node)

    trigger_count = _rebalance_trigger_count()
    LOG.info(f"dnsrebalancer triggered {trigger_count} time(s)")
    assert trigger_count == 1, (
        "Expected exactly one dnsrebalancer trigger "
        f"(triggered {trigger_count} times)"
    )

    for _ in range(6):
        time.sleep(10)
        result = initial_node.exec(get_coredns, text=True, capture_output=True)
        assert _is_rebalanced(
            result
        ), f"CoreDNS did not remain spread and ready: {result.stdout}"
        assert {
            pod["metadata"]["uid"] for pod in _settled_pods(result)
        } == settled_uids, "CoreDNS pods were replaced after the rollout settled"
        assert (
            _rebalance_trigger_count() == trigger_count
        ), "dnsrebalancer kept triggering after the rollout settled"
    util.wait_for_dns(initial_node)


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
