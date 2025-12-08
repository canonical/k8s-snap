#
# Copyright 2025 Canonical, Ltd.
#
import collections
import itertools
import logging
import time
from typing import Iterable, List

import hvac
import pytest
import requests
import yaml
from test_util import harness, tags, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)

# https://developer.hashicorp.com/vault/api-docs/system/health
VAULT_STATUS_ACTIVE = 200

DEFAULT_BOOTSTRAP_CONFIG = {
    "cluster-config": {
        "network": {"enabled": True},
        "dns": {"enabled": True},
        "metrics-server": {"enabled": True},
    },
}

CONTROL_PLANE_INTERNAL_CERT_NAMES = {
    "front-proxy-client",
    "apiserver-kubelet-client",
    "admin.conf",
    "scheduler.conf",
    "controller.conf",
    "kubelet.conf",
    "proxy.conf",
}


K8S_ROLE = "k8s"
MASTERS_ROLE = "k8s-masters"
NODES_ROLE = "k8s-nodes"
ROLES = {
    # role_name: organization,
    K8S_ROLE: "",
    MASTERS_ROLE: "system:masters",
    NODES_ROLE: "system:nodes",
}
APISERVER_DNS_NAMES = [
    "kubernetes",
    "kubernetes.default",
    "kubernetes.default.svc",
    "kubernetes.default.svc.cluster",
    "kubernetes.default.svc.cluster.local",
]
APISERVER_IP_SANS = ["127.0.0.1", "10.152.183.1"]

CertOpts = collections.namedtuple(
    "CertOpts",
    ["common_name", "role", "alt_names", "ip_sans"],
    defaults=[K8S_ROLE, [], []],
)


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


def create_roles(client: hvac.Client, roles: dict):
    for role, organization in roles.items():
        client.secrets.pki.create_or_update_role(
            role,
            {
                "ttl": "5h",
                "allow_localhost": "false",
                "allow_any_name": "true",
                "enforce_hostnames": "false",
                "organization": organization,
            },
        )


def create_certificate(client: hvac.Client, opts: CertOpts):
    config = {"ip_sans": ",".join(opts.ip_sans), "alt_names": ",".join(opts.alt_names)}

    # Generate certificate.
    generate_resp = client.secrets.pki.generate_certificate(
        name=opts.role, common_name=opts.common_name, extra_params=config
    )
    cert = generate_resp["data"]["certificate"]
    private_key = generate_resp["data"]["private_key"]

    return cert, private_key


def create_and_assign_certs(
    client: hvac.Client, cert_opts: Iterable, config_dict: dict
):
    for prefix, options in cert_opts:
        cert, key = create_certificate(client, options)
        config_dict[f"{prefix}-crt"] = cert
        config_dict[f"{prefix}-key"] = key


def check_nginx_pod_runs(instance: harness.Instance):
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


def delete_nginx_pod(instance: harness.Instance):
    manifest = MANIFESTS_DIR / "nginx-pod.yaml"
    instance.exec(["k8s", "kubectl", "delete", "-f", "-"], input=manifest.read_bytes())


@pytest.mark.node_count(3)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
# For communication with Vault
@pytest.mark.required_ports(8200)
def test_vault_intermediate_ca(instances: List[harness.Instance], datastore_type: str):
    instance = instances[0]
    cp_node = instances[1]
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
    bootstrap_config = dict(DEFAULT_BOOTSTRAP_CONFIG)
    bootstrap_config.update(
        {
            "ca-crt": ca_cert,
            "ca-key": ca_key,
            "client-ca-crt": client_ca_cert,
            "client-ca-key": client_ca_key,
            "front-proxy-ca-crt": proxy_ca_cert,
            "front-proxy-ca-key": proxy_ca_key,
        }
    )

    util.bootstrap(
        instance, datastore_type=datastore_type, bootstrap_config=bootstrap_config
    )

    # Add a control plane node and a worker node.
    join_token = util.get_join_token(instance, cp_node)
    util.join_cluster(cp_node, join_token)

    join_token = util.get_join_token(instance, worker_node, "--worker")
    util.join_cluster(worker_node, join_token)

    util.wait_until_k8s_ready(instance, instances)
    util.wait_for_dns(instance)

    # If we deploy a Pod and it becomes Active, the cluster should be functional.
    check_nginx_pod_runs(instance)


@pytest.mark.node_count(3)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
# For communication with Vault
@pytest.mark.required_ports(8200)
def test_vault_certificates(instances: List[harness.Instance], datastore_type: str):
    instance = instances[0]
    bootstrap_node_ip = util.get_default_ip(instance)
    bootstrap_node_hostname = util.hostname(instance)

    cp_node = instances[1]
    cp_node_ip = util.get_default_ip(cp_node)
    cp_node_hostname = util.hostname(cp_node)

    worker_node = instances[2]
    worker_ip = util.get_default_ip(worker_node)
    worker_hostname = util.hostname(worker_node)

    client = setup_vault(instance, bootstrap_node_ip)
    LOG.info("Vault setup is ready. Creating certificates.")

    # Enable and tune PKI.
    client.sys.enable_secrets_engine("pki", config={"max_lease_ttl": "6h"})

    # Generate root CA.
    gen_root_resp = client.secrets.pki.generate_root("internal", common_name="vault")
    root_ca_cert = gen_root_resp["data"]["certificate"]

    # Create roles for generating certificates.
    create_roles(client, ROLES)

    bootstrap_config = dict(DEFAULT_BOOTSTRAP_CONFIG)
    bootstrap_config.update(
        {
            "ca-crt": root_ca_cert,
            "client-ca-crt": root_ca_cert,
            "front-proxy-ca-crt": root_ca_cert,
        }
    )

    # If the role is not specified, ROLE_K8S will be used instead.
    cp_certs = {
        "apiserver": CertOpts(
            common_name="kube-apiserver",
            alt_names=APISERVER_DNS_NAMES,
            ip_sans=APISERVER_IP_SANS + [bootstrap_node_ip, cp_node_ip],
        ),
        "admin-client": CertOpts(
            common_name="kubernetes:admin",
            role=MASTERS_ROLE,
        ),
        "kube-controller-manager-client": CertOpts("system:kube-controller-manager"),
        "kube-scheduler-client": CertOpts("system:kube-scheduler"),
        "front-proxy-client": CertOpts("front-proxy-client"),
    }
    worker_certs = {
        "kubelet": CertOpts(
            common_name=f"system:node:{bootstrap_node_hostname}",
            role=NODES_ROLE,
            alt_names=[bootstrap_node_hostname],
            ip_sans=["127.0.0.1", bootstrap_node_ip],
        ),
        "kubelet-client": CertOpts(
            common_name=f"system:node:{bootstrap_node_hostname}",
            role=NODES_ROLE,
        ),
        "kube-proxy-client": CertOpts("system:kube-proxy"),
    }

    bootstrap_certs = {
        "apiserver-kubelet-client": CertOpts(
            common_name="apiserver-kubelet-client",
            role=MASTERS_ROLE,
        )
    }
    bootstrap_certs.update(cp_certs)
    bootstrap_certs.update(worker_certs)

    create_and_assign_certs(client, bootstrap_certs.items(), bootstrap_config)

    # For BootstrapConfig, only this key has a different format.
    bootstrap_config["kube-ControllerManager-client-key"] = bootstrap_config[
        "kube-controller-manager-client-key"
    ]
    bootstrap_config.pop("kube-controller-manager-client-key")

    LOG.info("Certificates are ready. Bootstrapping.")
    util.bootstrap(
        instance, datastore_type=datastore_type, bootstrap_config=bootstrap_config
    )

    # Add a control plane node.
    worker_certs["kubelet"] = CertOpts(
        common_name=f"system:node:{cp_node_hostname}",
        role=NODES_ROLE,
        alt_names=[cp_node_hostname],
        ip_sans=["127.0.0.1", cp_node_ip],
    )
    worker_certs["kubelet-client"] = CertOpts(
        common_name=f"system:node:{cp_node_hostname}",
        role=NODES_ROLE,
    )
    cp_join_config = {}
    create_and_assign_certs(
        client, itertools.chain(cp_certs.items(), worker_certs.items()), cp_join_config
    )

    join_token = util.get_join_token(instance, cp_node)
    cp_node.exec(
        ["k8s", "join-cluster", join_token, "--file", "-"],
        input=str.encode(yaml.dump(cp_join_config)),
    )

    # Add a worker node.
    worker_certs["kubelet"] = CertOpts(
        common_name=f"system:node:{worker_hostname}",
        role=NODES_ROLE,
        alt_names=[worker_hostname],
        ip_sans=["127.0.0.1", worker_ip],
    )
    worker_certs["kubelet-client"] = CertOpts(
        common_name=f"system:node:{worker_hostname}",
        role=NODES_ROLE,
    )
    worker_config = {}
    create_and_assign_certs(client, worker_certs.items(), worker_config)

    join_token = util.get_join_token(instance, worker_node, "--worker")
    worker_node.exec(
        ["k8s", "join-cluster", join_token, "--file", "-"],
        input=str.encode(yaml.dump(worker_config)),
    )

    util.wait_until_k8s_ready(instance, instances)
    util.wait_for_dns(instance)

    # If we deploy a Pod and it becomes Active, the cluster should be functional.
    check_nginx_pod_runs(instance)
    delete_nginx_pod(instance)

    # Refresh all cluster's nodes PKI.
    leader_cp_certs = {}
    cp_certs["apiserver-kubelet-client"] = CertOpts(
        common_name="apiserver-kubelet-client",
        role=MASTERS_ROLE,
    )
    worker_certs["kubelet"] = CertOpts(
        common_name=f"system:node:{bootstrap_node_hostname}",
        role=NODES_ROLE,
        alt_names=[bootstrap_node_hostname],
        ip_sans=["127.0.0.1", bootstrap_node_ip],
    )
    worker_certs["kubelet-client"] = CertOpts(
        common_name=f"system:node:{bootstrap_node_hostname}",
        role=NODES_ROLE,
    )
    create_and_assign_certs(
        client, itertools.chain(cp_certs.items(), worker_certs.items()), leader_cp_certs
    )
    instance.exec(
        ["k8s", "refresh-certs", "--external-certificates", "-"],
        input=str.encode(yaml.dump(leader_cp_certs)),
    )

    new_cp_certs = {}
    worker_certs["kubelet"] = CertOpts(
        common_name=f"system:node:{cp_node_hostname}",
        role=NODES_ROLE,
        alt_names=[cp_node_hostname],
        ip_sans=["127.0.0.1", cp_node_ip],
    )
    worker_certs["kubelet-client"] = CertOpts(
        common_name=f"system:node:{cp_node_hostname}",
        role=NODES_ROLE,
    )
    create_and_assign_certs(
        client, itertools.chain(cp_certs.items(), worker_certs.items()), new_cp_certs
    )
    cp_node.exec(
        ["k8s", "refresh-certs", "--external-certificates", "-"],
        input=str.encode(yaml.dump(new_cp_certs)),
    )

    new_worker_certs = {}
    worker_certs["kubelet"] = CertOpts(
        common_name=f"system:node:{worker_hostname}",
        role=NODES_ROLE,
        alt_names=[worker_hostname],
        ip_sans=["127.0.0.1", worker_ip],
    )
    worker_certs["kubelet-client"] = CertOpts(
        common_name=f"system:node:{worker_hostname}",
        role=NODES_ROLE,
    )
    create_and_assign_certs(client, worker_certs.items(), new_worker_certs)
    worker_node.exec(
        ["k8s", "refresh-certs", "--external-certificates", "-"],
        input=str.encode(yaml.dump(new_worker_certs)),
    )

    # Deploy the Pod again to verify the cluster functionality.
    check_nginx_pod_runs(instance)


@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
# For communication with Vault
@pytest.mark.required_ports(8200)
def test_partial_refresh(instances: List[harness.Instance], datastore_type: str):
    instance = instances[0]
    bootstrap_node_ip = util.get_default_ip(instance)
    bootstrap_node_hostname = util.hostname(instance)

    cp_node = instances[1]
    cp_node_ip = util.get_default_ip(cp_node)
    cp_node_hostname = util.hostname(cp_node)

    client = setup_vault(instance, bootstrap_node_ip)
    LOG.info("Vault setup is ready. Creating certificates.")

    # Enable and tune PKI.
    client.sys.enable_secrets_engine("pki", config={"max_lease_ttl": "6h"})

    # Generate root CA.
    gen_root_resp = client.secrets.pki.generate_root("internal", common_name="vault")
    root_ca_cert = gen_root_resp["data"]["certificate"]

    # Create roles for generating certificates.
    create_roles(client, ROLES)

    bootstrap_config = dict(DEFAULT_BOOTSTRAP_CONFIG)
    bootstrap_config.update(
        {
            "ca-crt": root_ca_cert,
        }
    )

    cp_certs = {
        "apiserver": CertOpts(
            common_name="kube-apiserver",
            alt_names=APISERVER_DNS_NAMES,
            ip_sans=APISERVER_IP_SANS + [bootstrap_node_ip, cp_node_ip],
        ),
    }
    worker_certs = {
        "kubelet": CertOpts(
            common_name=f"system:node:{bootstrap_node_hostname}",
            role=NODES_ROLE,
            alt_names=[bootstrap_node_hostname],
            ip_sans=["127.0.0.1", bootstrap_node_ip],
        ),
    }

    bootstrap_certs = {}
    bootstrap_certs.update(cp_certs)
    bootstrap_certs.update(worker_certs)

    create_and_assign_certs(client, bootstrap_certs.items(), bootstrap_config)

    LOG.info("Certificates are ready. Bootstrapping.")
    util.bootstrap(
        instance, datastore_type=datastore_type, bootstrap_config=bootstrap_config
    )

    # Add a control plane node.
    worker_certs["kubelet"] = CertOpts(
        common_name=f"system:node:{cp_node_hostname}",
        role=NODES_ROLE,
        alt_names=[cp_node_hostname],
        ip_sans=["127.0.0.1", cp_node_ip],
    )
    cp_join_config = {}
    create_and_assign_certs(
        client, itertools.chain(cp_certs.items(), worker_certs.items()), cp_join_config
    )

    join_token = util.get_join_token(instance, cp_node)
    cp_node.exec(
        ["k8s", "join-cluster", join_token, "--file", "-"],
        input=str.encode(yaml.dump(cp_join_config)),
    )

    util.wait_until_k8s_ready(instance, instances)
    util.wait_for_dns(instance)

    # If we deploy a Pod and it becomes Active, the cluster should be functional.
    check_nginx_pod_runs(instance)
    delete_nginx_pod(instance)

    # Refresh all cluster's nodes PKI.
    leader_cp_certs = {}
    worker_certs["kubelet"] = CertOpts(
        common_name=f"system:node:{bootstrap_node_hostname}",
        role=NODES_ROLE,
        alt_names=[bootstrap_node_hostname],
        ip_sans=["127.0.0.1", bootstrap_node_ip],
    )
    create_and_assign_certs(
        client, itertools.chain(cp_certs.items(), worker_certs.items()), leader_cp_certs
    )
    # Refresh External PKI
    instance.exec(
        ["k8s", "refresh-certs", "--external-certificates", "-"],
        input=str.encode(yaml.dump(leader_cp_certs)),
    )

    instance.exec(
        [
            "k8s",
            "refresh-certs",
            "--certificates",
            ",".join(CONTROL_PLANE_INTERNAL_CERT_NAMES),
            "--expires-in",
            "1y",
        ]
    )

    new_cp_certs = {}
    worker_certs["kubelet"] = CertOpts(
        common_name=f"system:node:{cp_node_hostname}",
        role=NODES_ROLE,
        alt_names=[cp_node_hostname],
        ip_sans=["127.0.0.1", cp_node_ip],
    )
    create_and_assign_certs(
        client, itertools.chain(cp_certs.items(), worker_certs.items()), new_cp_certs
    )
    cp_node.exec(
        ["k8s", "refresh-certs", "--external-certificates", "-"],
        input=str.encode(yaml.dump(new_cp_certs)),
    )
    cp_node.exec(
        [
            "k8s",
            "refresh-certs",
            "--certificates",
            ",".join(CONTROL_PLANE_INTERNAL_CERT_NAMES),
            "--expires-in",
            "1y",
        ]
    )

    # Deploy the Pod again to verify the cluster functionality.
    check_nginx_pod_runs(instance)
