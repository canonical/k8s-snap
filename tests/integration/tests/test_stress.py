#
# Copyright 2025 Canonical, Ltd.
#
import contextlib
import logging
import random
import time
from typing import List

import kubernetes
import pytest
import yaml
from test_util import config, etcd, harness, tags, util

LOG = logging.getLogger(__name__)


# NOTE: tags.CONFORMANCE is used for testing the current PR.
@pytest.mark.node_count(3)
@pytest.mark.tags(tags.CONFORMANCE)
def test_stress(instances: List[harness.Instance]):
    _cluster_setup(instances)
    util.wait_for_dns(instances[0])

    _run_tests(instances)


# NOTE: tags.CONFORMANCE is used for testing the current PR.
@pytest.mark.node_count(3)
@pytest.mark.etcd_count(3)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.CONFORMANCE)
def test_stress_etcd(instances: List[harness.Instance], etcd_cluster: etcd.EtcdCluster):
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

    _cluster_setup(instances, skip_k8s_dqlite=True)
    util.wait_for_dns(cp_node)

    _run_tests(instances)


def _cluster_setup(instances: List[harness.Instance], skip_k8s_dqlite: bool = False):
    cluster_node = instances[0]
    joining_nodes = instances[1:]

    for joining_node in joining_nodes:
        join_token = util.get_join_token(cluster_node, joining_node)
        util.join_cluster(joining_node, join_token)

    skip_services = ["k8s-dqlite"] if skip_k8s_dqlite else []
    util.wait_until_k8s_ready(cluster_node, instances, skip_services=skip_services)

    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == len(instances), "node should have joined cluster"

    for instance in instances:
        assert "control-plane" in util.get_local_node_status(instance)
        instance.exec(["mkdir", "-p", "/root/.kube"])
        config = instance.exec(["k8s", "config"], capture_output=True)
        instance.exec(["dd", "of=/root/.kube/config"], input=config.stdout)


def _run_tests(instances: List[harness.Instance]):
    # The first node may be the leader.
    instance = instances[-1]

    # Install kubectl in the instance, it's faster.
    instance.exec(["snap", "install", "kubectl", "--classic"])

    # Taint the other nodes, so Pods do not schedule on them.
    for node in instances[:-1]:
        instance.exec(["kubectl", "cordon", node.id])

    yaml_bytes = (config.MANIFESTS_DIR / "nginx-pod.yaml").read_bytes()
    pod_spec = yaml.safe_load(yaml_bytes)
    pod_name = pod_spec["metadata"]["name"]

    # Create a pod, just to cache the image locally.
    instance.exec(["kubectl", "apply", "-f", "-"], input=yaml_bytes)
    _wait_pod(instance, pod_name)
    instance.exec(["kubectl", "delete", "pod", pod_name])

    # Initialize kubernetes clients.
    proc = instance.exec(["k8s", "config"], capture_output=True)
    config_dict = yaml.safe_load(proc.stdout)
    kubernetes.config.load_kube_config_from_dict(config_dict)
    v1 = kubernetes.client.CoreV1Api()

    proc = instances[0].exec(["k8s", "config"], capture_output=True)
    config_dict = yaml.safe_load(proc.stdout)
    kubernetes.config.load_kube_config_from_dict(config_dict)
    other_v1 = kubernetes.client.CoreV1Api()

    # Run scenarios.
    aff_errors, aff_non_running = _check_node_affinity(instance, v1, other_v1)
    taint_errors, taint_non_running = _check_node_taint(instance, v1, other_v1)

    assert aff_errors == 0, "label: encountered errors while testing."
    assert aff_non_running == 0, "label: pods didn't always enter a running state."
    assert taint_errors == 0, "taint: encountered errors while testing."
    assert taint_non_running == 0, "taint: pods didn't always enter a running state."


def _wait_pod(
    instance: harness.Instance,
    pod_name: str,
    wait_for: str = "condition=Ready",
    timeout: int = 180,
    check: bool = True,
):
    instance.exec(
        [
            "kubectl",
            "wait",
            f"--for={wait_for}",
            f"--timeout={timeout}s",
            "pod",
            pod_name,
        ],
        check=check,
    )


def _check_node_affinity(instance: harness.Instance, v1, other_v1):
    yaml_bytes = (config.MANIFESTS_DIR / "nginx-pod-selector.yaml").read_bytes()
    pod_spec = yaml.safe_load(yaml_bytes)
    affinity = pod_spec["spec"]["affinity"]["nodeAffinity"]
    required = affinity["requiredDuringSchedulingIgnoredDuringExecution"]
    affinity_expression = required["nodeSelectorTerms"][0]["matchExpressions"][0]

    labels = {}
    label_dict = {
        "metadata": {
            "labels": labels,
        },
    }

    @contextlib.contextmanager
    def _run_scenario():
        suffix = random.randint(0, 999999)
        label = f"foo-{suffix}"
        labels[label] = "lish"
        pod_name = f"nginx-{suffix}"
        pod_spec["metadata"]["name"] = pod_name

        # Label the node and create Pod.
        v1.patch_node(instance.id, label_dict)
        affinity_expression["key"] = label
        other_v1.create_namespaced_pod(body=pod_spec, namespace="default")

        yield pod_name

        # Cleanup.
        v1.delete_namespaced_pod(pod_name, namespace="default")

        labels[label] = None
        v1.patch_node(instance.id, label_dict)
        labels.pop(label)

    LOG.info("Testing node label selector scenario")
    return _check_scenario(instance, _run_scenario)


def _check_node_taint(instance: harness.Instance, v1, other_v1):
    yaml_bytes = (config.MANIFESTS_DIR / "nginx-pod.yaml").read_bytes()
    pod_spec = yaml.safe_load(yaml_bytes)

    add_taint = {
        "spec": {
            "taints": [
                {
                    "key": "foo",
                    "value": "lish",
                    "effect": "NoSchedule",
                }
            ],
        },
    }

    remove_taint = {
        "spec": {
            "taints": [],
        },
    }

    @contextlib.contextmanager
    def _run_scenario():
        suffix = random.randint(0, 999999)
        pod_name = f"nginx-{suffix}"
        pod_spec["metadata"]["name"] = pod_name

        # Taint the node.
        v1.patch_node(instance.id, add_taint)
        time.sleep(0.25)

        # Untaint the node, create Pod, and wait for it to become Running.
        v1.patch_node(instance.id, remove_taint)
        other_v1.create_namespaced_pod(body=pod_spec, namespace="default")

        yield pod_name

        # Cleanup.
        v1.delete_namespaced_pod(pod_name, namespace="default")

    LOG.info("Testing node taint scenario.")
    return _check_scenario(instance, _run_scenario)


def _check_scenario(instance: harness.Instance, run_scenario):
    tries = 500
    errors = 0
    non_running = 0
    for i in range(tries):
        try:
            with run_scenario() as pod_name:
                _wait_pod(instance, pod_name, timeout=5, check=False)

                proc = instance.exec(
                    [
                        "kubectl",
                        "get",
                        "pod",
                        pod_name,
                        "-o",
                        "jsonpath='{.status.phase}'",
                    ],
                    check=False,
                    capture_output=True,
                )

                state = proc.stdout.decode()
                if proc.returncode:
                    errors += 1
                    LOG.error(
                        "[%d] Getting pod '%s' returned nonzero exit code.", i, pod_name
                    )
                elif state != "'Running'":
                    LOG.error(
                        "[%d] Pod '%s' has unexpected state: %s", i, pod_name, state
                    )
                    non_running += 1
        except Exception as ex:
            LOG.error("[%d] Encountered error: %s", i, ex)
            errors += 1

    LOG.info(
        "Finished trying %d times. Errors: %d, non-running times: %d",
        tries,
        errors,
        non_running,
    )

    return errors, non_running
