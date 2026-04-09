#
# Copyright 2025 Canonical, Ltd.
#
import datetime
import json
import logging
import os
import subprocess
import threading
from typing import List

import pytest
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
    joining_node_1.exec(["k8s", "remove-node", cluster_node.id])
    util.stubbornly(retries=5, delay_s=3).until(
        lambda _: (
            not util.diverged_cluster_memberships(
                joining_node_1, [joining_node_1, joining_node_2, joining_node_3]
            )
        )
    )

    # Verify that a node can remove itself
    joining_node_2.exec(["k8s", "remove-node", joining_node_1.id])
    util.stubbornly(retries=5, delay_s=3).until(
        lambda _: (
            not util.diverged_cluster_memberships(
                joining_node_2, [joining_node_2, joining_node_3]
            )
        )
    )


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_etcd_learner_promotion(instances: List[harness.Instance]):
    """Test that joining a second node doesn't cause etcd quorum loss.

    When going from 1 to 2 etcd members, the new node must be added as a
    non-voting learner first. Without this, MemberAdd changes quorum from
    1/1 to 2/2 before the new etcd starts, causing a deadlock.

    This test creates continuous write pressure on etcd during the join to
    ensure any quorum loss is immediately detected as write timeouts.
    """
    cluster_node = instances[0]
    joining_node = instances[1]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    write_failures = []
    write_count = [0]
    stop_writer = threading.Event()

    def _continuous_etcd_writes():
        seq = 0
        while not stop_writer.is_set():
            try:
                result = cluster_node.exec(
                    [
                        "k8s",
                        "kubectl",
                        "create",
                        "configmap",
                        f"etcd-quorum-test-{seq}",
                        f"--from-literal=seq={seq}",
                        "-n",
                        "default",
                    ],
                    capture_output=True,
                    text=True,
                    timeout=30,
                )
                if result.returncode != 0:
                    write_failures.append(
                        f"write {seq}: exit {result.returncode}: {result.stderr.strip()}"
                    )
            except subprocess.TimeoutExpired:
                write_failures.append(f"write {seq}: timed out after 30s")
            except Exception as e:
                write_failures.append(f"write {seq}: {e}")
            seq += 1
            write_count[0] = seq
            stop_writer.wait(0.5)

    writer = threading.Thread(target=_continuous_etcd_writes, daemon=True)
    writer.start()

    try:
        join_token = util.get_join_token(cluster_node, joining_node)
        util.join_cluster(joining_node, join_token)
    finally:
        stop_writer.set()
        writer.join(timeout=60)

    LOG.info("Background writer completed %d writes during join", write_count[0])

    assert not write_failures, (
        f"etcd writes failed during node join, indicating quorum loss. "
        f"{len(write_failures)} failures out of {write_count[0]} writes: "
        + "; ".join(write_failures[:5])
    )

    util.wait_until_k8s_ready(cluster_node, instances)
    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_node)

    util.stubbornly(retries=5, delay_s=3).until(
        lambda _: (
            not util.diverged_cluster_memberships(
                cluster_node, [cluster_node, joining_node]
            )
        )
    )

    proc = cluster_node.exec(
        [
            "curl",
            "--cacert",
            "/etc/kubernetes/pki/etcd/ca.crt",
            "--cert",
            "/etc/kubernetes/pki/apiserver-etcd-client.crt",
            "--key",
            "/etc/kubernetes/pki/apiserver-etcd-client.key",
            "https://127.0.0.1:2379/v3/cluster/member/list",
            "-X",
            "POST",
            "-H",
            "'Content-Type: application/json'",
        ],
        capture_output=True,
        text=True,
    )
    etcd_members = json.loads(proc.stdout).get("members", [])
    assert len(etcd_members) == 2, f"Expected 2 etcd members, got {len(etcd_members)}"
    for member in etcd_members:
        assert not member.get(
            "isLearner", False
        ), f"etcd member {member.get('name', 'unknown')} should not be a learner after join"

    for instance in instances:
        result = instance.exec(["k8s", "kubectl", "get", "nodes"], capture_output=True)
        assert (
            result.returncode == 0
        ), f"kubectl failed on {instance.id}: {result.stderr}"


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
    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_cp)
    assert "worker" in util.get_local_node_status(joining_worker)


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
