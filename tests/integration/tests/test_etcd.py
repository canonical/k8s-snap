#
# Copyright 2024 Canonical, Ltd.
#
import logging
import textwrap
from typing import List

import pytest
from test_util import harness, util

LOG = logging.getLogger(__name__)

ETCD_URL = "https://github.com/etcd-io/etcd/releases/download"
ETCD_VERSION = "v3.3.8"


@pytest.mark.node_count(2)
def test_etcd(instances: List[harness.Instance]):
    instance = instances[0]
    instance_default_ip = util.get_default_ip(instance)

    openssl_conf = f"""
    [req]
    default_bits  = 4096
    distinguished_name = req_distinguished_name
    req_extensions = v3_req
    prompt = no

    [req_distinguished_name]
    countryName = US
    stateOrProvinceName = CA
    localityName = San Francisco
    organizationName = etcd
    commonName = etcd-host

    [v3_req]
    keyUsage = digitalSignature, keyEncipherment, dataEncipherment
    extendedKeyUsage = serverAuth, clientAuth
    subjectAltName = @alt_names

    [alt_names]
    IP.1 = 127.0.0.1
    IP.2 = {instance_default_ip}
    DNS.1 = localhost
    DNS.2 = {instance.id}
    """
    instance.exec(
        ["dd", "of=/tmp/etcd_tls.conf"],
        input=str.encode(openssl_conf),
    )

    instance.exec(
        [
            "openssl",
            "req",
            "-x509",
            "-nodes",
            "-newkey",
            "rsa:4096",
            "-subj",
            "/CN=etcdRootCA",
            "-keyout",
            "/tmp/ca_key.pem",
            "-out",
            "/tmp/ca_cert.pem",
            "-days",
            "36500",
        ]
    )

    instance.exec(
        [
            "openssl",
            "req",
            "-nodes",
            "-newkey",
            "rsa:4096",
            "-keyout",
            "/tmp/etcd_key.pem",
            "-out",
            "/tmp/etcd_cert.csr",
            "-config",
            "/tmp/etcd_tls.conf",
        ]
    )

    instance.exec(
        [
            "openssl",
            "x509",
            "-req",
            "-days",
            "36500",
            "-in",
            "/tmp/etcd_cert.csr",
            "-CA",
            "/tmp/ca_cert.pem",
            "-CAkey",
            "/tmp/ca_key.pem",
            "-out",
            "/tmp/etcd_cert.pem",
            "-extensions",
            "v3_req",
            "-extfile",
            "/tmp/etcd_tls.conf",
            "-CAcreateserial",
        ]
    )

    instance.exec(["mkdir", "-p", "/tmp/test-etcd"])

    instance.exec(
        [
            "curl",
            "-L",
            f"{ETCD_URL}/{ETCD_VERSION}/etcd-{ETCD_VERSION}-linux-amd64.tar.gz",
            "-o",
            f"/tmp/etcd-{ETCD_VERSION}-linux-amd64.tar.gz",
        ]
    )

    instance.exec(
        [
            "tar",
            "xzvf",
            f"/tmp/etcd-{ETCD_VERSION}-linux-amd64.tar.gz",
            "-C",
            "/tmp/test-etcd",
            "--strip-components=1",
        ]
    )

    etcd_service = f"""
    [Unit]
    Description=etcd
    Documentation=https://github.com/etcd-io/etcd
    Conflicts=etcd.service
    Conflicts=etcd2.service

    [Service]
    Type=notify
    Restart=always
    RestartSec=5s
    LimitNOFILE=40000
    TimeoutStartSec=0

    ExecStart=/tmp/test-etcd/etcd --name s1 \
    --data-dir /tmp/etcd/s1 \
    --listen-client-urls https://{instance_default_ip}:2379 \
    --advertise-client-urls https://{instance_default_ip}:2379 \
    --listen-peer-urls https://{instance_default_ip}:2380 \
    --initial-advertise-peer-urls https://{instance_default_ip}:2380 \
    --initial-cluster s1=https://{instance_default_ip}:2380 \
    --initial-cluster-token tkn \
    --initial-cluster-state new \
    --client-cert-auth \
    --trusted-ca-file /tmp/ca_cert.pem \
    --cert-file /tmp/etcd_cert.pem \
    --key-file /tmp/etcd_key.pem \
    --peer-client-cert-auth \
    --peer-trusted-ca-file /tmp/ca_cert.pem \
    --peer-cert-file /tmp/etcd_cert.pem \
    --peer-key-file /tmp/etcd_key.pem

    [Install]
    WantedBy=multi-user.target
    """
    instance.exec(
        ["dd", "of=/etc/systemd/system/etcd-s1.service"],
        input=str.encode(etcd_service),
    )

    instance.exec(["systemctl", "daemon-reload"])
    instance.exec(["systemctl", "enable", "etcd-s1.service"])
    instance.exec(["systemctl", "start", "etcd-s1.service"])

    p = instance.exec(["cat", "/tmp/ca_cert.pem"], capture_output=True)
    ca_cert = p.stdout.decode()
    p = instance.exec(["cat", "/tmp/etcd_cert.pem"], capture_output=True)
    etcd_cert = p.stdout.decode()
    p = instance.exec(["cat", "/tmp/etcd_key.pem"], capture_output=True)
    etcd_key = p.stdout.decode()

    k8s_instance = instances[1]

    bootstrap_conf = f"""
datastore: external
datastore-url: https://{instance_default_ip}:2379
datastore-ca-crt: |
{textwrap.indent(ca_cert, "  ")}
datastore-client-crt: |
{textwrap.indent(etcd_cert, "  ")}
datastore-client-key: |
{textwrap.indent(etcd_key, "  ")}
    """

    k8s_instance.exec(
        ["dd", "of=/root/config.yaml"],
        input=str.encode(bootstrap_conf),
    )

    k8s_instance.exec(["k8s", "bootstrap", "--config", "/root/config.yaml"])
    util.wait_for_dns(k8s_instance)
    util.wait_for_network(k8s_instance)

    p = k8s_instance.exec(
        ["systemctl", "is-active", "--quiet", "snap.k8s.k8s-dqlite"], check=False
    )
    assert p.returncode != 0
