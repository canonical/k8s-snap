#
# Copyright 2025 Canonical, Ltd.
#
import concurrent.futures
import datetime
import logging
import os
import subprocess
from typing import List

import pytest
import yaml
from cryptography import x509
from cryptography.hazmat.backends import default_backend
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_control_plane_nodes(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node = instances[1]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token = util.get_join_token(cluster_node, joining_node)
    util.join_cluster(joining_node, join_token)

    util.wait_until_k8s_ready(cluster_node, instances)
    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_node)

    cluster_node.exec(["k8s", "remove-node", joining_node.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 1, "node should have been removed from cluster"
    assert (
        nodes[0]["metadata"]["name"] == cluster_node.id
    ), f"only {cluster_node.id} should be left in cluster"


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_worker_nodes(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node = instances[1]
    other_joining_node = instances[2]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token = util.get_join_token(cluster_node, joining_node, "--worker")
    join_token_2 = util.get_join_token(cluster_node, other_joining_node, "--worker")

    assert join_token != join_token_2

    util.join_cluster(joining_node, join_token)

    util.join_cluster(other_joining_node, join_token_2)

    util.wait_until_k8s_ready(cluster_node, instances)

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "worker" in util.get_local_node_status(joining_node)
    assert "worker" in util.get_local_node_status(other_joining_node)

    cluster_node.exec(["k8s", "remove-node", joining_node.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "worker should have been removed from cluster"
    assert cluster_node.id in [
        node["metadata"]["name"] for node in nodes
    ] and other_joining_node.id in [
        node["metadata"]["name"] for node in nodes
    ], f"only {cluster_node.id} should be left in cluster"


@pytest.mark.node_count(3)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.PULL_REQUEST)
def test_disa_stig_clustering(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp = instances[1]
    joining_worker = instances[2]

    util.setup_k8s_snap(cluster_node)
    bootstrapFile = config.COMMON_ETC_DIR + "/templates/disa-stig/bootstrap.yaml"
    cluster_node.exec(["sysctl", "-w", "vm.overcommit_memory=1"])
    cluster_node.exec(["sysctl", "-w", "kernel.panic=10"])
    cluster_node.exec(["sysctl", "-w", "kernel.panic_on_oops=1"])

    cluster_node.exec(["k8s", "bootstrap", "--file", bootstrapFile])
    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    util.setup_k8s_snap(joining_cp)
    cp_file = config.COMMON_ETC_DIR + "/templates/disa-stig/control-plane.yaml"
    join_token_cp = util.get_join_token(cluster_node, joining_cp)

    cp_file_content = joining_cp.exec(
        ["cat", cp_file], capture_output=True, text=True
    ).stdout
    cp_data = yaml.safe_load(cp_file_content)
    joining_cp.exec(["sysctl", "-w", "vm.overcommit_memory=1"])
    joining_cp.exec(["sysctl", "-w", "kernel.panic=10"])
    joining_cp.exec(["sysctl", "-w", "kernel.panic_on_oops=1"])
    util.join_cluster(joining_cp, join_token_cp, yaml.dump(cp_data))

    util.setup_k8s_snap(joining_worker)
    worker_file = config.COMMON_ETC_DIR + "/templates/disa-stig/worker.yaml"
    join_token_worker = util.get_join_token(cluster_node, joining_worker, "--worker")

    worker_file_content = joining_worker.exec(
        ["cat", worker_file], capture_output=True, text=True
    ).stdout
    worker_data = yaml.safe_load(worker_file_content)
    joining_worker.exec(["sysctl", "-w", "vm.overcommit_memory=1"])
    joining_worker.exec(["sysctl", "-w", "kernel.panic=10"])
    joining_worker.exec(["sysctl", "-w", "kernel.panic_on_oops=1"])
    util.join_cluster(joining_worker, join_token_worker, yaml.dump(worker_data))

    util.wait_until_k8s_ready(cluster_node, instances)
    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_cp)
    assert "worker" in util.get_local_node_status(joining_worker)


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_concurrent_membership_operations(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp_A = instances[1]
    joining_cp_B = instances[2]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token_A = util.get_join_token(cluster_node, joining_cp_A)
    join_token_B = util.get_join_token(cluster_node, joining_cp_B)

    assert join_token_A != join_token_B, "Join tokens should be different"

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        future_A = executor.submit(
            util.stubbornly(retries=5, delay_s=1).on(joining_cp_A).exec,
            ["k8s", "join-cluster", join_token_A],
        )
        future_B = executor.submit(
            util.stubbornly(retries=5, delay_s=1).on(joining_cp_B).exec,
            ["k8s", "join-cluster", join_token_B],
        )
        # Optionally, wait for both to complete
        concurrent.futures.wait([future_A, future_B])

    util.wait_until_k8s_ready(cluster_node, instances)

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_cp_A)
    assert "control-plane" in util.get_local_node_status(joining_cp_B)

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        util.stubbornly(retries=5, delay_s=1).on(cluster_node).execute(
            executor.submit(cluster_node.exec, ["k8s", "remove-node", joining_cp_A.id])
        )
        util.stubbornly(retries=5, delay_s=1).on(cluster_node).execute(
            executor.submit(cluster_node.exec, ["k8s", "remove-node", joining_cp_B.id])
        )

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    nodes = util.ready_nodes(cluster_node)
    assert (
        len(nodes) == 1
    ), "two control-plane nodes should have been removed from cluster"

    assert cluster_node.id in [node["metadata"]["name"] for node in nodes]


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_mixed_concurrent_membership_operations(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp_A = instances[1]
    joining_cp_B = instances[2]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token_A = util.get_join_token(cluster_node, joining_cp_A)
    join_token_B = util.get_join_token(cluster_node, joining_cp_B)

    assert join_token_A != join_token_B, "Join tokens should be different"

    util.join_cluster(joining_cp_A, join_token_A)
    util.wait_until_k8s_ready(cluster_node, [cluster_node, joining_cp_A])

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_cp_A)

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        executor.submit(cluster_node.exec, ["k8s", "remove-node", joining_cp_A.id])
        executor.submit(util.join_cluster, joining_cp_B, join_token_B)

    util.wait_until_k8s_ready(cluster_node, [cluster_node, joining_cp_B])

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_cp_B)

    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "two control-plane nodes should be in the cluster"

    assert cluster_node.id in [
        node["metadata"]["name"] for node in nodes
    ] and joining_cp_B.id in [
        node["metadata"]["name"] for node in nodes
    ], f"only {cluster_node.id} and {joining_cp_B.id} should be left in cluster"


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.NIGHTLY)
def test_concurrent_membership_restart_operations(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp_A = instances[1]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token_A = util.get_join_token(cluster_node, joining_cp_A)

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        future_join = executor.submit(
            util.stubbornly(retries=5, delay_s=1).on(joining_cp_A).exec,
            ["k8s", "join-cluster", join_token_A],
        )
        future_restart = executor.submit(
            util.stubbornly(retries=5, delay_s=1).on(cluster_node).exec,
            ["snap", "restart", config.SNAP_NAME],
        )
        concurrent.futures.wait([future_join, future_restart])

    util.wait_until_k8s_ready(cluster_node, instances)

    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "two control-plane nodes should be in the cluster"

    assert cluster_node.id in [
        node["metadata"]["name"] for node in nodes
    ] and joining_cp_A.id in [
        node["metadata"]["name"] for node in nodes
    ], f"{cluster_node.id} and {joining_cp_A.id} should be in the cluster"

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_cp_A)


@pytest.mark.node_count(4)
@pytest.mark.tags(tags.NIGHTLY)
def test_node_join_succeeds_when_original_control_plane_is_down(
    instances: List[harness.Instance],
):
    cluster_node = instances[0]
    joining_cp_A = instances[1]
    joining_cp_B = instances[2]
    joining_cp_C = instances[3]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token_A = util.get_join_token(cluster_node, joining_cp_A)
    join_token_B = util.get_join_token(cluster_node, joining_cp_B)
    join_token_C = util.get_join_token(cluster_node, joining_cp_C)

    assert (
        join_token_A != join_token_B and join_token_B != join_token_C
    ), "Join tokens should be different"

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        util.stubbornly(retries=5, delay_s=1).on(cluster_node).execute(
            executor.submit(util.join_cluster, joining_cp_A, join_token_A)
        )
        util.stubbornly(retries=5, delay_s=1).on(cluster_node).execute(
            executor.submit(util.join_cluster, joining_cp_B, join_token_B)
        )

    util.wait_until_k8s_ready(cluster_node, [cluster_node, joining_cp_A, joining_cp_B])

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_cp_A)
    assert "control-plane" in util.get_local_node_status(joining_cp_B)

    cluster_node.exec(["snap", "stop", config.SNAP_NAME])

    util.join_cluster(joining_cp_C, join_token_C)

    util.wait_until_k8s_ready(joining_cp_A, [joining_cp_A, joining_cp_B, joining_cp_C])

    nodes = util.ready_nodes(joining_cp_A)
    assert (
        len(nodes) == 3
    ), "three control plane nodes should be ready, original node is still down"

    assert (
        joining_cp_A.id in [node["metadata"]["name"] for node in nodes]
        and joining_cp_B.id in [node["metadata"]["name"] for node in nodes]
        and joining_cp_C.id in [node["metadata"]["name"] for node in nodes]
    ), f"{joining_cp_A.id}, {joining_cp_B.id}, and {joining_cp_C.id} should be ready and in the cluster"

    joining_cp_C.exec(["k8s", "remove-node", cluster_node.id])
    nodes = util.ready_nodes(joining_cp_C)
    assert (
        len(nodes) == 3
    ), "three control plane nodes should be ready, original node is removed"

    assert (
        joining_cp_A.id in [node["metadata"]["name"] for node in nodes]
        and joining_cp_B.id in [node["metadata"]["name"] for node in nodes]
        and joining_cp_C.id in [node["metadata"]["name"] for node in nodes]
    ), f"{joining_cp_A.id}, {joining_cp_B.id}, and {joining_cp_C.id} should be ready and in the cluster"


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_join_with_custom_token_name(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp = instances[1]
    joining_cp_with_hostname = instances[2]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    out = cluster_node.exec(
        ["k8s", "get-join-token", "my-token"],
        capture_output=True,
        text=True,
    )
    join_token = out.stdout.strip()

    join_config = """
extra-sans:
- my-token
"""
    joining_cp.exec(
        ["k8s", "join-cluster", join_token, "--name", "my-node", "--file", "-"],
        input=join_config,
        text=True,
    )

    out = cluster_node.exec(
        ["k8s", "get-join-token", "my-token-2"],
        capture_output=True,
        text=True,
    )
    join_token_2 = out.stdout.strip()

    join_config_2 = """
extra-sans:
- my-token-2
"""
    joining_cp_with_hostname.exec(
        ["k8s", "join-cluster", join_token_2, "--file", "-"],
        input=join_config_2,
        text=True,
    )

    util.wait_until_k8s_ready(
        cluster_node, instances, node_names={joining_cp.id: "my-node"}
    )

    cluster_node.exec(["k8s", "remove-node", "my-node"])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "cp node should be removed from the cluster"

    cluster_node.exec(["k8s", "remove-node", joining_cp_with_hostname.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 1, "cp node with hostname should be removed from the cluster"


@pytest.mark.node_count(2)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-csr-auto-approve.yaml").read_text()
)
@pytest.mark.tags(tags.NIGHTLY)
def test_cert_refresh(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_worker = instances[1]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token_worker = util.get_join_token(cluster_node, joining_worker, "--worker")
    util.join_cluster(joining_worker, join_token_worker)

    util.wait_until_k8s_ready(cluster_node, instances)
    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "worker" in util.get_local_node_status(joining_worker)

    extra_san = "test_san.local"

    def _check_cert(instance, cert_fname):
        # Ensure that the certificate was refreshed, having the right expiry date
        # and extra SAN.
        cert_dir = _get_k8s_cert_dir(instance)
        cert_path = os.path.join(cert_dir, cert_fname)

        cert = _get_instance_cert(instance, cert_path)
        date = datetime.datetime.now()
        assert (cert.not_valid_after - date).days in (364, 365)

        san = cert.extensions.get_extension_for_class(x509.SubjectAlternativeName)
        san_dns_names = san.value.get_values_for_type(x509.DNSName)
        assert extra_san in san_dns_names

    joining_worker.exec(
        ["k8s", "refresh-certs", "--expires-in", "1y", "--extra-sans", extra_san]
    )

    _check_cert(joining_worker, "kubelet.crt")

    cluster_node.exec(
        ["k8s", "refresh-certs", "--expires-in", "1y", "--extra-sans", extra_san]
    )

    _check_cert(cluster_node, "kubelet.crt")
    _check_cert(cluster_node, "apiserver.crt")

    # Ensure that the services come back online after refreshing the certificates.
    util.wait_until_k8s_ready(cluster_node, instances)


def _get_k8s_cert_dir(instance: harness.Instance):
    tested_paths = [
        "/etc/kubernetes/pki/",
        "/var/snap/k8s/common/etc/kubernetes/pki/",
    ]
    for path in tested_paths:
        if _instance_path_exists(instance, path):
            return path

    raise Exception("Could not find k8s certificates dir.")


def _instance_path_exists(instance: harness.Instance, remote_path: str):
    try:
        instance.exec(["ls", remote_path])
        return True
    except subprocess.CalledProcessError:
        return False


def _get_instance_cert(
    instance: harness.Instance, remote_path: str
) -> x509.Certificate:
    result = instance.exec(["cat", remote_path], capture_output=True)
    pem = result.stdout
    cert = x509.load_pem_x509_certificate(pem, default_backend())
    return cert
