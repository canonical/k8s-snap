#
# Copyright 2024 Canonical, Ltd.
#
import datetime
import logging
import os
import subprocess
import tempfile
from typing import List

import pytest
from cryptography import x509
from cryptography.hazmat.backends import default_backend
from test_util import config, harness, util, tags

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_control_plane_nodes(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node = instances[1]

    join_token = util.get_join_token(cluster_node, joining_node)
    util.join_cluster(joining_node, join_token)

    util.wait_until_k8s_ready(cluster_node, instances)
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "node should have joined cluster"

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_node)

    cluster_node.exec(["k8s", "remove-node", joining_node.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 1, "node should have been removed from cluster"
    assert (
        nodes[0]["metadata"]["name"] == cluster_node.id
    ), f"only {cluster_node.id} should be left in cluster"


@pytest.mark.node_count(2)
@pytest.mark.snap_versions([util.previous_track(config.SNAP), config.SNAP])
@pytest.mark.tags(tags.NIGHTLY)
def test_mixed_version_join(instances: List[harness.Instance]):
    """Test n versioned node joining a n-1 versioned cluster."""
    cluster_node = instances[0]  # bootstrapped on the previous channel
    joining_node = instances[1]  # installed with the snap under test

    join_token = util.get_join_token(cluster_node, joining_node)
    util.join_cluster(joining_node, join_token)

    util.wait_until_k8s_ready(cluster_node, instances)
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "node should have joined cluster"

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

    join_token = util.get_join_token(cluster_node, joining_node, "--worker")
    join_token_2 = util.get_join_token(cluster_node, other_joining_node, "--worker")

    assert join_token != join_token_2

    util.join_cluster(joining_node, join_token)

    util.join_cluster(other_joining_node, join_token_2)

    util.wait_until_k8s_ready(cluster_node, instances)
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 3, "workers should have joined cluster"

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
@pytest.mark.tags(tags.NIGHTLY)
def test_join_with_custom_token_name(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp = instances[1]
    joining_cp_with_hostname = instances[2]

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
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 3, "nodes should have joined cluster"

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

    join_token_worker = util.get_join_token(cluster_node, joining_worker, "--worker")
    util.join_cluster(joining_worker, join_token_worker)

    util.wait_until_k8s_ready(cluster_node, instances)
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "nodes should have joined cluster"

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
    with tempfile.NamedTemporaryFile() as fp:
        instance.pull_file(remote_path, fp.name)

        pem = fp.read()
        cert = x509.load_pem_x509_certificate(pem, default_backend())
        return cert
