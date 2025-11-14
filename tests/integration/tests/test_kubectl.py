#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
def test_kubectl(instances: List[harness.Instance]):
    """
    Test kubectl behavior before bootstrap, after bootstrap, and when k8sd is stopped, on both
    control-plane and worker nodes.
    """
    cp, worker = instances

    # kubectl should fail when cluster is not bootstrapped
    LOG.info("Testing kubectl before bootstrap...")
    result = cp.exec(
        ["k8s", "kubectl", "get", "nodes"],
        capture_output=True,
        check=False,
        text=True,
    )
    assert result.returncode != 0, "kubectl should fail before bootstrap"
    LOG.info(f"kubectl failed as expected: {result.stderr}")

    # Bootstrap the cluster
    LOG.info("Bootstrapping the cluster and joining the worker...")
    cp.exec(["k8s", "bootstrap"])
    join_token = util.get_join_token(cp, worker, "--worker")
    util.join_cluster(worker, join_token)

    util.wait_until_k8s_ready(cp, instances)

    # kubectl should work after bootstrap on control-plane node
    LOG.info("Testing kubectl on control-plane node after bootstrap...")
    result = cp.exec(
        ["k8s", "kubectl", "get", "nodes"],
    )
    assert (
        result.returncode == 0
    ), "kubectl should work after bootstrap on control-plane"
    LOG.info("kubectl works after bootstrap on control-plane")

    # kubectl should not work on worker node without KUBECONFIG
    LOG.info("Testing kubectl on worker node after bootstrap without KUBECONFIG...")
    result = worker.exec(
        ["k8s", "kubectl", "get", "node", worker.id],
        capture_output=True,
        text=True,
        check=False,
    )
    assert (
        result.returncode != 0
    ), "kubectl should not work on worker without KUBECONFIG"
    LOG.info(f"kubectl fails as expected on worker without KUBECONFIG: {result.stderr}")

    # kubectl should work on worker node with KUBECONFIG
    LOG.info("Testing kubectl on worker node after bootstrap with KUBECONFIG...")
    result = worker.exec(
        [
            "bash",
            "-c",
            f"KUBECONFIG=/etc/kubernetes/kubelet.conf k8s kubectl get node {worker.id}",
        ],
    )
    assert result.returncode == 0, "kubectl should work on worker with KUBECONFIG"
    LOG.info("kubectl works on worker with KUBECONFIG")

    # Stop k8sd service
    LOG.info("Stopping k8sd service...")
    cp.exec(["snap", "stop", "k8s.k8sd"])
    worker.exec(["snap", "stop", "k8s.k8sd"])

    # Give it a moment to fully stop
    util.stubbornly(retries=10, delay_s=3).on(cp).until(
        lambda p: "inactive" in p.stdout.decode()
    ).exec(["snap", "services", "k8s.k8sd"])

    util.stubbornly(retries=10, delay_s=3).on(worker).until(
        lambda p: "inactive" in p.stdout.decode()
    ).exec(["snap", "services", "k8s.k8sd"])

    # kubectl should still work when k8sd is stopped
    LOG.info("Testing kubectl on control-plane node after k8sd is stopped...")
    result = cp.exec(
        ["k8s", "kubectl", "get", "nodes"],
    )
    assert (
        result.returncode == 0
    ), "kubectl should work on control-plane node even when k8sd is stopped"
    LOG.info("kubectl works on control-plane node with k8sd stopped")

    LOG.info("Testing kubectl on worker node after k8sd is stopped...")
    result = worker.exec(
        [
            "bash",
            "-c",
            f"KUBECONFIG=/etc/kubernetes/kubelet.conf k8s kubectl get node {worker.id}",
        ],
    )
    assert (
        result.returncode == 0
    ), "kubectl should work on worker node even when k8sd is stopped"
    LOG.info("kubectl works on worker node with k8sd stopped")
