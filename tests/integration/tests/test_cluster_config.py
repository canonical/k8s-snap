#
# Copyright 2025 Canonical, Ltd.
#
import json
import logging
from typing import Any, List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(2)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-cluster-config.yaml").read_text()
)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_cluster_config(instances: List[harness.Instance]):
    """Test cluster config endpoint"""
    cp_node = instances[0]
    worker_node = instances[1]
    join_token = util.get_join_token(cp_node, worker_node, "--worker")
    util.join_cluster(
        worker_node,
        join_token,
        (config.MANIFESTS_DIR / "worker-join-cluster-config.yaml").read_text(),
    )

    util.wait_until_k8s_ready(cp_node, instances)
    nodes = util.ready_nodes(cp_node)
    assert len(nodes) == 2, "worker should have joined cluster"

    cp_resp = get_cluster_config(cp_node)
    worker_resp = get_cluster_config(worker_node)

    # NOTE(Hue): The following are based on the `bootstrap-cluster-config.yaml` and `worker-join-cluster-config.yaml`
    # manifest files. If the manifest files are changed, the following expected values should be updated.
    exp_cp_config = {
        "network": {
            "enabled": True,
            "pod-cidr": "10.1.0.0/16",
            "service-cidr": "10.152.183.0/24",
        },
        "dns": {
            "enabled": True,
            "cluster-domain": "cluster.local",
            "service-ip": "10.152.183.200",
            "upstream-nameservers": ["/etc/resolv.conf"],
        },
        "ingress": {
            "enabled": True,
            "default-tls-secret": "",
            "enable-proxy-protocol": False,
        },
        "load-balancer": {
            "enabled": True,
            "cidrs": [],
            "l2-mode": True,
            "l2-interfaces": [],
            "bgp-mode": False,
            "bgp-local-asn": 0,
            "bgp-peer-address": "",
            "bgp-peer-asn": 0,
            "bgp-peer-port": 0,
        },
        "local-storage": {
            "enabled": True,
            "local-path": "/var/snap/k8s/common/rawfile-storage",
            "reclaim-policy": "Delete",
            "default": True,
        },
        "gateway": {
            "enabled": True,
        },
        "metrics-server": {
            "enabled": True,
        },
    }
    exp_cp_datastore = {"type": "k8s-dqlite"}
    exp_cp_taints = ["taint1=:PreferNoSchedule", "taint2=value:PreferNoSchedule"]
    exp_worker_taints = [
        "workerTaint1=:PreferNoSchedule",
        "workerTaint2=workerValue:PreferNoSchedule",
    ]

    assert (
        cp_resp["datastore"] == exp_cp_datastore
    ), f"Mismatch in {cp_resp['datastore']} and {exp_cp_datastore=}"
    assert (
        cp_resp["nodeTaints"] == exp_cp_taints
    ), f"Mismatch in {cp_resp['nodeTaints']} and {exp_cp_taints=}"
    assert (
        worker_resp["nodeTaints"] == exp_worker_taints
    ), f"Mismatch in {worker_resp['nodeTaints']} and {exp_worker_taints=}"
    assert (
        cp_resp["status"]["network"] == exp_cp_config["network"]
    ), f"Mismatch in {cp_resp['status']['network']} and {exp_cp_config['network']=}"
    assert (
        cp_resp["status"]["dns"]["enabled"] == exp_cp_config["dns"]["enabled"]
    ), f"Mismatch in {cp_resp['status']['dns']['enabled']} and {exp_cp_config['dns']['enabled']=}"
    assert (
        cp_resp["status"]["dns"]["cluster-domain"]
        == exp_cp_config["dns"]["cluster-domain"]
    ), f"Mismatch in {cp_resp['status']['dns']['cluster-domain']} and {exp_cp_config['dns']['cluster-domain']=}"
    assert (
        cp_resp["status"]["dns"]["upstream-nameservers"]
        == exp_cp_config["dns"]["upstream-nameservers"]
    ), f"Mismatch in {cp_resp['status']['dns']['upstream-nameservers']} and "
    "{exp_cp_config['dns']['upstream-nameservers']=}"
    assert (
        cp_resp["status"]["ingress"] == exp_cp_config["ingress"]
    ), f"Mismatch in {cp_resp['status']['ingress']} and {exp_cp_config['ingress']=}"
    assert (
        cp_resp["status"]["load-balancer"] == exp_cp_config["load-balancer"]
    ), f"Mismatch in {cp_resp['status']['load-balancer']} and {exp_cp_config['load-balancer']=}"
    assert (
        cp_resp["status"]["local-storage"] == exp_cp_config["local-storage"]
    ), f"Mismatch in {cp_resp['status']['local-storage']} and {exp_cp_config['local-storage']=}"
    assert (
        cp_resp["status"]["gateway"] == exp_cp_config["gateway"]
    ), f"Mismatch in {cp_resp['status']['gateway']} and {exp_cp_config['gateway']=}"
    assert (
        cp_resp["status"]["metrics-server"] == exp_cp_config["metrics-server"]
    ), f"Mismatch in {cp_resp['status']['metrics-server']} and {exp_cp_config['metrics-server']=}"


def get_cluster_config(instance: harness.Instance) -> Any:
    """Get the cluster config."""
    resp = instance.exec(
        [
            "curl",
            "-H",
            "Content-Type: application/json",
            "--unix-socket",
            "/var/snap/k8s/common/var/lib/k8sd/state/control.socket",
            "http://localhost/1.0/k8sd/cluster/config",
        ],
        capture_output=True,
    )
    assert resp.returncode == 0, f"Failed to get cluster config. {resp=}"
    response = json.loads(resp.stdout.decode())
    assert response["error_code"] == 0, f"Failed to get cluster config. {response=}"
    assert response["error"] == "", f"Failed to get cluster config. {response=}"

    metadata = response.get("metadata")
    assert metadata is not None, "Metadata not found in the cluster config response."
    assert (
        metadata.get("status") is not None
    ), "Config not found in the cluster config response."
    assert (
        metadata.get("datastore") is not None
    ), "Datastore not found in the cluster config response."
    assert (
        metadata.get("nodeTaints") is not None
    ), "Node taints not found in the cluster config response."

    return metadata
