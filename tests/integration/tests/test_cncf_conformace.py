#
# Copyright 2025 Canonical, Ltd.
#
import logging
import re
from typing import List

import pytest
import yaml
from test_util import etcd, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.CONFORMANCE)
def test_cncf_conformance(instances: List[harness.Instance]):
    cluster_node = cluster_setup(instances)

    _run_cncf_tests(cluster_node, "k8s-dqlite")


@pytest.mark.node_count(3)
@pytest.mark.etcd_count(3)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.CONFORMANCE)
def test_cncf_conformance_etcd(
    instances: List[harness.Instance], etcd_cluster: etcd.EtcdCluster
):
    cp_node = instances[0]

    bootstrap_conf = yaml.safe_dump(
        {
            "cluster-config": {"network": {"enabled": True}, "dns": {"enabled": True}},
            "datastore-type": "external",
            "datastore-servers": etcd_cluster.client_urls,
            "datastore-ca-crt": etcd_cluster.ca_cert,
            "datastore-client-crt": etcd_cluster.cert,
            "datastore-client-key": etcd_cluster.key,
        }
    )

    cp_node.exec(
        ["k8s", "bootstrap", "--file", "-"],
        input=str.encode(bootstrap_conf),
    )
    util.wait_for_dns(cp_node)
    util.wait_for_network(cp_node)

    cluster_setup(instances, skip_k8s_dqlite=True)

    _run_cncf_tests(cp_node, "etcd")


def _run_cncf_tests(instance: harness.Instance, suffix: str):
    install_sonobuoy(instance)

    label_focus = "validates resource limits of pods that are allowed to run"
    taint_focus = "removing taint cancels eviction|ConfigMap should be consumable from pods in volume"

    for i in range(200):
        LOG.info(f"Attempt {i} with {label_focus=}")
        _run_scenario(instance, label_focus, suffix)

    for i in range(200):
        LOG.info(f"Attempt {i} with {taint_focus=}")
        _run_scenario(instance, taint_focus, suffix)


def _run_scenario(instance: harness.Instance, focus: str, suffix: str):
    cmds = [
        ["./sonobuoy", "run", "--plugin", "e2e", "--e2e-focus", focus, "--wait"],
        ["./sonobuoy", "retrieve", "-f", "sonobuoy_e2e.tar.gz"],
        ["tar", "-xf", "sonobuoy_e2e.tar.gz", "--one-top-level"],
    ]
    for cmd in cmds:
        instance.exec(cmd)

    resp = instance.exec(
        ["./sonobuoy", "results", "sonobuoy_e2e.tar.gz"],
        capture_output=True,
    )
    output = resp.stdout.decode()

    # Clean up namespace, so it can be used again.
    instance.exec(["k8s", "kubectl", "delete", "ns", "sonobuoy"])

    failed_tests = int(re.search("Failed: (\\d+)", output).group(1))
    if not failed_tests:
        return

    instance.pull_file("/root/sonobuoy_e2e.tar.gz", f"sonobuoy_e2e_{suffix}.tar.gz")
    LOG.info(output)
    assert False, f"{focus=} test(s) failed"


def cluster_setup(
    instances: List[harness.Instance], skip_k8s_dqlite: bool = False
) -> harness.Instance:
    cluster_node = instances[0]
    joining_nodes = instances[1:]

    for joining_node in joining_nodes:
        join_token = util.get_join_token(cluster_node, joining_node)
        util.join_cluster(joining_node, join_token)

    skip_services = ["k8s-dqlite"] if skip_k8s_dqlite else []
    util.wait_until_k8s_ready(cluster_node, instances, skip_services=skip_services)

    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 3, "node should have joined cluster"
    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_nodes[0])
    assert "control-plane" in util.get_local_node_status(joining_nodes[1])

    config = cluster_node.exec(["k8s", "config"], capture_output=True)
    cluster_node.exec(["dd", "of=/root/.kube/config"], input=config.stdout)

    return cluster_node


def install_sonobuoy(instance: harness.Instance):
    instance.exec(
        ["curl", "-L", util.sonobuoy_tar_gz(instance.arch), "-o", "sonobuoy.tar.gz"]
    )
    instance.exec(["tar", "xvzf", "sonobuoy.tar.gz"])
    instance.exec(["./sonobuoy", "version"])
