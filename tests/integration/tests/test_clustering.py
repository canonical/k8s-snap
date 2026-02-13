#
# Copyright 2026 Canonical, Ltd.
#
import concurrent.futures
import datetime
import logging
import os
import subprocess
import time
from typing import List

import pytest
import tenacity
import yaml
from cryptography import x509
from cryptography.hazmat.backends import default_backend
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(4)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_control_plane_nodes(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node_1 = instances[1]
    joining_node_2 = instances[2]
    joining_node_3 = instances[3]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token = util.get_join_token(cluster_node, joining_node_1)
    util.join_cluster(joining_node_1, join_token)

    join_token = util.get_join_token(cluster_node, joining_node_2)
    util.join_cluster(joining_node_2, join_token)

    join_token = util.get_join_token(cluster_node, joining_node_3)
    util.join_cluster(joining_node_3, join_token)

    util.wait_until_k8s_ready(cluster_node, instances)
    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_node_1)
    assert "control-plane" in util.get_local_node_status(joining_node_2)
    assert "control-plane" in util.get_local_node_status(joining_node_3)

    # Verify that the initial node can be removed
    # Verify that the initial node can be removed
    util.remove_node_with_retry(joining_node_1, cluster_node.id)
    util.stubbornly(retries=5, delay_s=3).until(
        lambda _: not util.diverged_cluster_memberships(
            joining_node_1, [joining_node_1, joining_node_2, joining_node_3]
        )
    )

    # Verify that a node can remove itself
    util.remove_node_with_retry(joining_node_1, joining_node_1.id)
    util.stubbornly(retries=5, delay_s=3).until(
        lambda _: not util.diverged_cluster_memberships(
            joining_node_2, [joining_node_2, joining_node_3]
        )
    )


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
    assert "worker" in util.get_local_node_status(
        joining_node
    ), f"{joining_node.id} should be ready and in the cluster"
    assert "worker" in util.get_local_node_status(
        other_joining_node
    ), f"{other_joining_node.id} should be ready and in the cluster"

    util.remove_node_with_retry(cluster_node, joining_node.id)
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
    bootstrapFile = config.COMMON_ETC_DIR + "/configurations/disa-stig/bootstrap.yaml"
    cluster_node.exec(["sysctl", "-w", "vm.overcommit_memory=1"])
    cluster_node.exec(["sysctl", "-w", "kernel.panic=10"])
    cluster_node.exec(["sysctl", "-w", "kernel.panic_on_oops=1"])

    cluster_node.exec(["k8s", "bootstrap", "--file", bootstrapFile])
    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    util.setup_k8s_snap(joining_cp)
    cp_file = config.COMMON_ETC_DIR + "/configurations/disa-stig/control-plane.yaml"
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
    worker_file = config.COMMON_ETC_DIR + "/configurations/disa-stig/worker.yaml"
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
    assert "control-plane" in util.get_local_node_status(
        cluster_node
    ), f"{cluster_node.id} should be ready and in the cluster"
    assert "control-plane" in util.get_local_node_status(
        joining_cp
    ), f"{joining_cp.id} should be ready and in the cluster"
    assert "worker" in util.get_local_node_status(
        joining_worker
    ), f"{joining_worker.id} should be ready and in the cluster"


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_concurrent_cp_membership_operations(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp_A = instances[1]
    joining_cp_B = instances[2]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        future_A = executor.submit(join_node_with_retry, cluster_node, joining_cp_A)
        future_B = executor.submit(join_node_with_retry, cluster_node, joining_cp_B)
        concurrent.futures.wait([future_A, future_B])

    util.wait_until_k8s_ready(cluster_node, instances)

    for node in [cluster_node, joining_cp_A, joining_cp_B]:
        assert "control-plane" in util.get_local_node_status(node)

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        future_A = executor.submit(
            util.remove_node_with_retry, cluster_node, joining_cp_A.id
        )
        future_B = executor.submit(
            util.remove_node_with_retry, cluster_node, joining_cp_B.id
        )
        concurrent.futures.wait([future_A, future_B])

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    nodes = util.ready_nodes(cluster_node)
    assert (
        len(nodes) == 1
    ), "two control-plane nodes, should have been removed from cluster"

    assert cluster_node.id in [node["metadata"]["name"] for node in nodes]


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_concurrent_worker_membership_operations(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_worker_A = instances[1]
    joining_worker_B = instances[2]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        future_A = executor.submit(
            join_node_with_retry, cluster_node, joining_worker_A, worker=True
        )
        future_B = executor.submit(
            join_node_with_retry, cluster_node, joining_worker_B, worker=True
        )
        concurrent.futures.wait([future_A, future_B])

    util.wait_until_k8s_ready(cluster_node, instances)

    for node in [joining_worker_A, joining_worker_B]:
        assert "worker" in util.get_local_node_status(node)

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        future_A = executor.submit(
            util.remove_node_with_retry, cluster_node, joining_worker_A.id
        )
        future_B = executor.submit(
            util.remove_node_with_retry, cluster_node, joining_worker_B.id
        )
        concurrent.futures.wait([future_A, future_B])

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 1, "two worker nodes should have been removed from cluster"

    assert cluster_node.id in [node["metadata"]["name"] for node in nodes]


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_mixed_concurrent_membership_operations(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp_A = instances[1]
    joining_cp_B = instances[2]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token_A = util.get_join_token(cluster_node, joining_cp_A)

    util.join_cluster(joining_cp_A, join_token_A)
    util.wait_until_k8s_ready(cluster_node, [cluster_node, joining_cp_A])

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_cp_A)

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        future_A = executor.submit(
            util.remove_node_with_retry, cluster_node, joining_cp_A.id
        )
        future_B = executor.submit(join_node_with_retry, cluster_node, joining_cp_B)
        concurrent.futures.wait([future_A, future_B])

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

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        future_join = executor.submit(join_node_with_retry, cluster_node, joining_cp_A)
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
    """Test that joining a permanently down control plane node can
    completely be removed from the cluster and does not affect future joins."""

    cluster_node = instances[0]
    joining_cp_A = instances[1]
    joining_cp_B = instances[2]
    joining_cp_C = instances[3]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token_A = util.get_join_token(cluster_node, joining_cp_A)
    join_token_B = util.get_join_token(cluster_node, joining_cp_B)

    util.join_cluster(joining_cp_A, join_token_A)
    util.join_cluster(joining_cp_B, join_token_B)

    util.wait_until_k8s_ready(cluster_node, [cluster_node, joining_cp_A, joining_cp_B])

    for node in [cluster_node, joining_cp_A, joining_cp_B]:
        assert "control-plane" in util.get_local_node_status(node)

    cluster_node_id = cluster_node.id
    cluster_node.delete()

    util.stubbornly(retries=15, delay_s=10).on(joining_cp_A).until(
        lambda p: "NotReady" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "node", cluster_node_id])

    util.remove_node_with_retry(joining_cp_A, cluster_node_id, force=True)

    # Now join a new node to verify that the cluster is still functional.
    join_token_C = util.get_join_token(joining_cp_A, joining_cp_C)
    util.join_cluster(joining_cp_C, join_token_C)

    util.wait_until_k8s_ready(joining_cp_A, [joining_cp_A, joining_cp_B, joining_cp_C])

    nodes = util.ready_nodes(joining_cp_A)
    assert (
        len(nodes) == 3
    ), "three control plane nodes should be ready, original node is removed"

    node_names = {node["metadata"]["name"] for node in nodes}
    assert {joining_cp_A.id, joining_cp_B.id, joining_cp_C.id}.issubset(
        node_names
    ), f"{joining_cp_A.id}, {joining_cp_B.id}, and {joining_cp_C.id} should be ready and in the cluster"


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_node_removal_during_concurrent_join(
    instances: List[harness.Instance],
):
    cluster_node = instances[0]
    joining_cp_A = instances[1]
    joining_cp_B = instances[2]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token_A = util.get_join_token(cluster_node, joining_cp_A)
    join_token_B = util.get_join_token(cluster_node, joining_cp_B)
    assert join_token_A != join_token_B, "Join tokens should be different"

    util.join_cluster(joining_cp_A, join_token_A)
    util.wait_until_k8s_ready(cluster_node, [cluster_node, joining_cp_A])

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        future_remove = executor.submit(
            util.remove_node_with_retry, cluster_node, joining_cp_A.id
        )
        future_join = executor.submit(join_node_with_retry, cluster_node, joining_cp_B)
        concurrent.futures.wait([future_remove, future_join])

    util.wait_until_k8s_ready(cluster_node, [cluster_node, joining_cp_B])

    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "There should be two control-plane nodes in the cluster"

    node_names = {node["metadata"]["name"] for node in nodes}
    assert {cluster_node.id, joining_cp_B.id}.issubset(
        node_names
    ), f"{cluster_node.id} and {joining_cp_B.id} should be ready and in the cluster"


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

    util.remove_node_with_retry(cluster_node, "my-node")
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "cp node should be removed from the cluster"

    util.remove_node_with_retry(cluster_node, joining_cp_with_hostname.id)
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
    assert "control-plane" in util.get_local_node_status(
        cluster_node
    ), f"{cluster_node.id} should be ready and in the cluster"
    assert "worker" in util.get_local_node_status(
        joining_worker
    ), f"{joining_worker.id} should be ready and in the cluster"

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


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_join_cp_with_duplicate_name_rejected(instances: List[harness.Instance]):
    """Tests that a CP node joining with the same name as a worker node in the cluster is rejected"""
    cluster_node = instances[0]
    joined_worker = instances[1]
    joining_cp = instances[2]
    shared_node_name = "nutella"

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token_worker = util.get_join_token(
        cluster_node, joined_worker, "--worker", name=shared_node_name
    )
    util.join_cluster(joined_worker, join_token_worker, name=shared_node_name)
    util.wait_until_k8s_ready(
        cluster_node,
        [cluster_node, joined_worker],
        node_names={joined_worker.id: shared_node_name},
    )

    LOG.info("Worker node joined successfully")

    # Try to get join token with duplicate name - should fail
    try:
        util.get_join_token(cluster_node, joining_cp, name=shared_node_name)
        assert False, "get-join-token should have failed due to duplicate node name"
    except tenacity.RetryError as e:
        LOG.info("get-join-token failed as expected")
        # Extract the underlying exception
        cause = e.last_attempt.exception()
        if not isinstance(cause, subprocess.CalledProcessError):
            raise e
        error_output = (
            cause.stderr if cause.stderr else cause.stdout if cause.stdout else ""
        )
        if isinstance(error_output, bytes):
            error_output = error_output.decode()
        assert (
            f'a node with this name is already part of the cluster: "{shared_node_name}"'
            in error_output
        ), f"Join error message should indicate duplicate node name. Got: {error_output}"

    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "Only master node and worker should be in the cluster"
    node_names = [node["metadata"]["name"] for node in nodes]
    assert cluster_node.id in node_names
    assert shared_node_name in node_names

    LOG.info("Successfully prevented joining of CP node with same name as worker")


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_join_worker_with_duplicate_name_rejected(instances: List[harness.Instance]):
    """Tests that a worker node joining with the same name as a control plane node in the cluster is rejected"""
    cluster_node = instances[0]
    joining_worker = instances[1]

    # Use the cluster node's name as the duplicate name
    shared_node_name = cluster_node.id

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    LOG.info("Cluster initialized")

    try:
        util.get_join_token(
            cluster_node, joining_worker, "--worker", name=shared_node_name
        )
        assert False, "get-join-token should have failed due to duplicate node name"
    except tenacity.RetryError as e:
        LOG.info("get-join-token failed as expected")
        cause = e.last_attempt.exception()
        if not isinstance(cause, subprocess.CalledProcessError):
            raise e
        error_output = (
            cause.stderr if cause.stderr else cause.stdout if cause.stdout else ""
        )
        if isinstance(error_output, bytes):
            error_output = error_output.decode()
        assert (
            f'a node with this name is already part of the cluster: "{shared_node_name}"'
            in error_output
        ), f"Join error message should indicate duplicate node name. Got: {error_output}"

    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 1, "Only original cluster node should be in the cluster"

    node_names = [node["metadata"]["name"] for node in nodes]
    assert cluster_node.id in node_names
    assert shared_node_name in node_names

    LOG.info(
        "Successfully prevented joining of worker node with same name as control plane"
    )


def join_node_with_retry(
    cluster_node, joining_node, retries=25, delay_s=1, worker=False
):
    """Join cluster with retry, generating a new token on each attempt"""
    for attempt in range(retries):
        try:
            if worker:
                join_token = util.get_join_token(cluster_node, joining_node, "--worker")
            else:
                join_token = util.get_join_token(cluster_node, joining_node)
            util.join_cluster(joining_node, join_token)
            break
        except Exception as e:
            if attempt == retries - 1:  # Last attempt
                raise
            LOG.info(
                f"Join attempt {attempt + 1} failed, retrying in {delay_s} second(s): {e}"
            )
            time.sleep(delay_s)


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
