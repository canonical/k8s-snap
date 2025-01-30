#
# Copyright 2025 Canonical, Ltd.
#
import logging
import time
from typing import List

import hvac
import pytest
import requests
import yaml
from test_util import harness, tags, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)

# https://developer.hashicorp.com/vault/api-docs/system/health
VAULT_STATUS_ACTIVE = 200


def wait_for_vault(client: hvac.Client, timeout: int = 30):
    """Wait for Vault to be available."""

    start = time.time()
    while time.time() - start < timeout:
        try:
            client.sys.read_health_status()
            return True
        except requests.exceptions.ConnectionError:
            LOG.info("Vault API is not yet available.")
        except Exception:
            pass

        time.sleep(1)

    return False


def setup_vault(instance: harness.Instance, instance_ip: str):
    """Install and initialize Vault."""

    LOG.info("Installing Vault.")
    instance.exec(["snap", "install", "vault"])
    instance.exec(["snap", "start", "vault"])

    LOG.info("Waiting for Vault to become available.")
    url = f"http://{instance_ip}:8200"
    client = hvac.Client(url)
    available = wait_for_vault(client)
    assert available, "Expected Vault to be available"

    LOG.info("Initializing Vault.")
    result = client.sys.initialize(1, 1)
    keys = result["keys"]
    root_token = result["root_token"]
    client.token = root_token

    client.sys.submit_unseal_key(keys[0])

    assert client.is_authenticated(), "Expected Vault to be authenticated"
    assert client.sys.is_initialized(), "Expected Vault to be initialized"
    assert not client.sys.is_sealed(), "Expected Vault to be unsealed"

    resp = client.sys.read_health_status()
    assert resp.status_code == VAULT_STATUS_ACTIVE, "Expected Vault to be active"

    return client


def create_intermediate_ca(client: hvac.Client, common_name: str):
    # Generate intermediate CA.
    interm_ca_resp = client.secrets.pki.generate_intermediate(
        "exported", common_name=common_name
    )
    private_key = interm_ca_resp["data"]["private_key"]

    # Sign intermediate CA.
    csr = interm_ca_resp["data"]["csr"]
    sign_resp = client.secrets.pki.sign_intermediate(csr=csr, common_name=common_name)
    cert = sign_resp["data"]["certificate"]

    return cert, private_key


@pytest.mark.node_count(3)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
def test_vault_intermediate_ca(instances: List[harness.Instance]):
    instance = instances[0]
    control_plane_node = instances[1]
    worker_node = instances[2]

    instance_ip = util.get_default_ip(instance)
    client = setup_vault(instance, instance_ip)

    LOG.info("Vault setup is ready. Creating certificates.")

    # Enable and tune PKI.
    client.sys.enable_secrets_engine("pki", config={"max_lease_ttl": "6h"})

    # Generate root CA.
    client.secrets.pki.generate_root("internal", common_name="vault")

    # Generate intermediate CAs.
    ca_cert, ca_key = create_intermediate_ca(client, instance_ip)
    client_ca_cert, client_ca_key = create_intermediate_ca(client, instance_ip)
    proxy_ca_cert, proxy_ca_key = create_intermediate_ca(client, instance_ip)

    LOG.info("Certificates are ready. Bootstrapping.")

    # Bootstrap K8s.
    bootstrap_config = {
        "cluster-config": {
            "network": {"enabled": True},
            "dns": {"enabled": True},
            "metrics-server": {"enabled": True},
        },
        "ca-crt": ca_cert,
        "ca-key": ca_key,
        "client-ca-crt": client_ca_cert,
        "client-ca-key": client_ca_key,
        "front-proxy-ca-crt": proxy_ca_cert,
        "front-proxy-ca-key": proxy_ca_key,
    }

    instance.exec(
        ["k8s", "bootstrap", "--file", "-"],
        input=str.encode(yaml.dump(bootstrap_config)),
    )

    # Add a control plane node and a worker node.
    join_token = util.get_join_token(instance, control_plane_node)
    util.join_cluster(control_plane_node, join_token)

    join_token = util.get_join_token(instance, worker_node, "--worker")
    util.join_cluster(worker_node, join_token)

    util.wait_until_k8s_ready(instance, instances)
    nodes = util.ready_nodes(instance)
    assert len(nodes) == 3, "node should have joined cluster"
    util.wait_for_dns(instance)

    # If we deploy a Pod and it becomes Active, the cluster should be functional.
    manifest = MANIFESTS_DIR / "nginx-pod.yaml"
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=manifest.read_bytes(),
    )

    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "app=nginx",
            "--timeout",
            "180s",
        ]
    )
