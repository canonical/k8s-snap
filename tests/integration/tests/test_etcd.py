#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
import time
from typing import List

import pytest
import yaml
from test_util import harness, util
from test_util.etcd import EtcdCluster

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.etcd_count(1)
@pytest.mark.disable_k8s_bootstrapping()
def test_etcd(instances: List[harness.Instance], etcd_cluster: EtcdCluster):
    k8s_instance = instances[0]

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

    k8s_instance.exec(
        ["dd", "of=/root/config.yaml"],
        input=str.encode(bootstrap_conf),
    )

    k8s_instance.exec(["k8s", "bootstrap", "--file", "/root/config.yaml"])
    util.wait_for_dns(k8s_instance)
    util.wait_for_network(k8s_instance)

    p = k8s_instance.exec(
        ["systemctl", "is-active", "--quiet", "snap.k8s.k8s-dqlite"], check=False
    )
    assert p.returncode != 0, "k8s-dqlite service is still active"

    LOG.info("Add new etcd nodes")
    etcd_cluster.add_nodes(2)

    # Update  server-urls in cluster
    body = {
        "datastore-config": {
            "type": "external",
            "servers": etcd_cluster.client_urls,
            "ca-crt": etcd_cluster.ca_cert,
            "client-crt": etcd_cluster.cert,
            "client-key": etcd_cluster.key,
        }
    }
    k8s_instance.exec(
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

    # Give some time to reconcile and restart services.
    time.sleep(2)

    # check that we can still connect to the kubernetes cluster
    k8s_instance.exec(["k8s", "kubectl", "get", "pods", "-A"])

    # Changing the datastore back to k8s-dqlite after using the external datastore should fail.
    body = {
        "datastore-config": {
            "type": "k8s-dqlite",
        }
    }

    resp = k8s_instance.exec(
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
