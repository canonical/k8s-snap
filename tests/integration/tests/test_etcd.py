#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
from typing import List

import pytest
import yaml
from test_util import harness, tags, util
from test_util.etcd import EtcdCluster

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(2)
@pytest.mark.etcd_count(1)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
def test_etcd(instances: List[harness.Instance], etcd_cluster: EtcdCluster):
    initial_node = instances[0]

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

    initial_node.exec(
        ["dd", "of=/root/config.yaml"],
        input=str.encode(bootstrap_conf),
    )

    initial_node.exec(["k8s", "bootstrap", "--file", "/root/config.yaml"])
    util.wait_for_dns(initial_node)
    util.wait_for_network(initial_node)

    p = initial_node.exec(
        ["systemctl", "is-active", "--quiet", "snap.k8s.k8s-dqlite"], check=False
    )
    assert p.returncode != 0, "k8s-dqlite service is still active"

    LOG.info("Add new etcd nodes")
    etcd_cluster.add_nodes(2)

    # Update server-urls in cluster
    body = {
        "datastore": {
            "type": "external",
            "servers": etcd_cluster.client_urls,
            "ca-crt": etcd_cluster.ca_cert,
            "client-crt": etcd_cluster.cert,
            "client-key": etcd_cluster.key,
        }
    }
    initial_node.exec(
        [
            "curl",
            "-XPUT",
            "--header",
            "Content-Type: application/json",
            "--data",
            json.dumps(body),
            "--unix-socket",
            "/var/snap/k8s/common/var/lib/k8sd/state/control.socket",
            "http://localhost/1.0/k8sd/cluster/config",
        ]
    )

    # check that we can still connect to the kubernetes cluster
    util.stubbornly(retries=10, delay_s=2).on(initial_node).exec(
        ["k8s", "kubectl", "get", "pods", "-A"]
    )

    # Changing the datastore back to k8s-dqlite after using the external datastore should fail.
    body = {
        "datastore": {
            "type": "k8s-dqlite",
        }
    }

    resp = initial_node.exec(
        [
            "curl",
            "-XPUT",
            "--header",
            "Content-Type: application/json",
            "--data",
            json.dumps(body),
            "--unix-socket",
            "/var/snap/k8s/common/var/lib/k8sd/state/control.socket",
            "http://localhost/1.0/k8sd/cluster/config",
        ],
        capture_output=True,
    )
    response = json.loads(resp.stdout.decode())
    assert response["error_code"] == 400, "changing the datastore type should fail"


    joining_cplane_node = instances[1]
    joining_worker_node = instances[2]

    join_token = util.get_join_token(initial_node, joining_cplane_node)
    join_token_2 = util.get_join_token(initial_node, joining_worker_node, "--worker")

    assert join_token != join_token_2

    util.join_cluster(joining_cplane_node, join_token)

    util.join_cluster(joining_worker_node, join_token_2)

    util.wait_until_k8s_ready(initial_node, instances)
    nodes = util.ready_nodes(initial_node)
    assert len(nodes) == 3, "all nodes should have joined cluster"

    initial_node.exec(["k8s", "set", "dns.cluster-domain=integration.local"])

    util.stubbornly(retries=5, delay_s=10).on(joining_cplane_node).until(
        lambda p: "--cluster-domain=integration.local" in p.stdout.decode()
    ).exec(["cat", "/var/snap/k8s/common/args/kubelet"])

    util.stubbornly(retries=5, delay_s=10).on(joining_worker_node).until(
        lambda p: "--cluster-domain=integration.local" in p.stdout.decode()
    ).exec(["cat", "/var/snap/k8s/common/args/kubelet"])
