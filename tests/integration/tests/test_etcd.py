#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
import yaml
from test_util import harness, util
from test_util.etcd import EtcdCluster

LOG = logging.getLogger(__name__)



@pytest.mark.node_count(1)
@pytest.mark.etcd_count(3)
def test_etcd(instances: List[harness.Instance], etcd_cluster: EtcdCluster):
    k8s_instance = instances[0]

    bootstrap_conf = yaml.safe_dump(
        {
            "cluster-config": {
                "network": {
                    "enabled": True
                },
                "dns": {
                    "enabled": True
                }
            },
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
