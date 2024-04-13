#
# Copyright 2024 Canonical, Ltd.
#
import logging
from string import Template
from typing import List

from test_util import config
from test_util.harness import Harness, Instance
from test_util.util import get_default_ip

LOG = logging.getLogger(__name__)


class EtcdCluster:
    """
    An Etcd cluster abstraction based on the Harness.
    Opposed to the k8s cluster instances, we normally don't need to access the etcd
    instances directly. This class provides abstractions to work with the cluster as a whole.

    Attributes:
        etcd_url (str): The URL of the etcd cluster.
        etcd_version (str): The version of etcd used in the cluster.
        instances (List[Instance]): List of instances in the etcd cluster.
    """

    def __init__(self, h: Harness, initial_node_count: int = 1):
        """
        Initialize the EtcdCluster object.

        Args:
            h (Harness): The test harness object.
            initial_node_count (int): Number of etcd nodes to create in the cluster.
        """
        self.etcd_url = config.ETCD_URL
        self.etcd_version = config.ETCD_VERSION
        self.instances: List[Instance] = []
        self.harness: Harness = h

        self.add_nodes(initial_node_count)


    def add_nodes(self, count: int):
        """
        Add a given number of nodes to the cluster.
        """
        for _ in range(count):
            self.add_node()

    def add_node(self):
        """
        Add a new node to the etcd cluster.
        If this is the first cluster node, the required certificates will be generated.
        """
        LOG.info("Setup etcd node")
        join_existing = len(self.instances) > 0

        instance = self.harness.new_instance()
        self.instances.append(instance)
        ip = get_default_ip(instance)

        if join_existing:
            endpoints = [
                f"https://{get_default_ip(i)}:2379" for i in self.instances[:-1]
            ]
            # add the member information to the cluster
            self.instances[0].exec(
                [
                    "ETCDCTL_API=3",
                    "/tmp/test-etcd/etcdctl",
                    "member",
                    "add",
                    instance.id,
                    "--peer-urls",
                    f"https://{ip}:2380",
                    "--cert",
                    "/tmp/client-cert.pem",
                    "--key",
                    "/tmp/client-key.pem",
                    "--cacert",
                    "/tmp/ca-cert.pem",
                    "--endpoints",
                    ",".join(endpoints),
                    "-w",
                    "json",
                ],
            )

        initial_cluster = [
            f"{ins.id}=https://{get_default_ip(ins)}:2380" for ins in self.instances
        ]

        substitutes = {
            "NAME": instance.id,
            "IP": ip,
            "CLIENT_URL": f"https://{ip}:2379",
            "PEER_URL": f"https://{ip}:2380",
            "CLUSTER": ",".join(initial_cluster),
            "CLUSTER_STATE": "existing" if join_existing else "new",
        }

        with open(config.ETCD_DIR / "etcd-tls.conf", "r") as etcd_template:
            src = Template(etcd_template.read())
            instance.exec(
                ["dd", "of=/tmp/etcd-tls.conf"],
                input=str.encode(src.substitute(substitutes)),
            )

        # Only create CA on the first node.
        if join_existing:
            instance.exec(
                ["dd", "of=/tmp/ca-cert.pem"],
                input=str.encode(self.ca_cert),
            )
            instance.exec(
                ["dd", "of=/tmp/ca-key.pem"],
                input=str.encode(self.ca_key),
            )
        else:
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
                    "/tmp/ca-key.pem",
                    "-out",
                    "/tmp/ca-cert.pem",
                    "-days",
                    "36500",
                ]
            )

        # Generate client CSR
        instance.exec(
            [
                "openssl",
                "req",
                "-nodes",
                "-newkey",
                "rsa:4096",
                "-keyout",
                "/tmp/client-key.pem",
                "-out",
                "/tmp/client-cert.csr",
                "-config",
                "/tmp/etcd-tls.conf",
            ]
        )
        
        # Sign client CSR
        instance.exec(
            [
                "openssl",
                "x509",
                "-req",
                "-days",
                "36500",
                "-in",
                "/tmp/client-cert.csr",
                "-CA",
                "/tmp/ca-cert.pem",
                "-CAkey",
                "/tmp/ca-key.pem",
                "-out",
                "/tmp/client-cert.pem",
                "-extensions",
                "v3_req",
                "-extfile",
                "/tmp/etcd-tls.conf",
                "-CAcreateserial",
            ]
        )
        
        # Generate server CSR
        instance.exec(
            [
                "openssl",
                "req",
                "-nodes",
                "-newkey",
                "rsa:4096",
                "-keyout",
                "/tmp/server-key.pem",
                "-out",
                "/tmp/server-cert.csr",
                "-config",
                "/tmp/etcd-tls.conf",
            ]
        )

        # Sign server CSR
        instance.exec(
            [
                "openssl",
                "x509",
                "-req",
                "-days",
                "36500",
                "-in",
                "/tmp/server-cert.csr",
                "-CA",
                "/tmp/ca-cert.pem",
                "-CAkey",
                "/tmp/ca-key.pem",
                "-out",
                "/tmp/server-cert.pem",
                "-extensions",
                "v3_req",
                "-extfile",
                "/tmp/etcd-tls.conf",
                "-CAcreateserial",
            ]
        )

        with open(config.ETCD_DIR / "etcd.service", "r") as etcd_template:
            src = Template(etcd_template.read())
            instance.exec(
                ["dd", "of=/etc/systemd/system/etcd-s1.service"],
                input=str.encode(src.substitute(substitutes)),
            )

        instance.exec(
            [
                "curl",
                "-L",
                f"{self.etcd_url}/{self.etcd_version}/etcd-{self.etcd_version}-linux-amd64.tar.gz",
                "-o",
                f"/tmp/etcd-{self.etcd_version}-linux-amd64.tar.gz",
            ]
        )
        instance.exec(["mkdir", "-p", "/tmp/test-etcd"])
        instance.exec(
            [
                "tar",
                "xzvf",
                f"/tmp/etcd-{self.etcd_version}-linux-amd64.tar.gz",
                "-C",
                "/tmp/test-etcd",
                "--strip-components=1",
            ],
        )
        instance.exec(["systemctl", "daemon-reload"])
        instance.exec(["systemctl", "enable", "etcd-s1.service"])
        instance.exec(["systemctl", "start", "etcd-s1.service"])

    @property
    def ca_cert(self) -> str:
        """
        Get the CA certificate of the etcd cluster.

        Returns:
            str: The CA certificate in PEM format.
        """
        p = self.instances[0].exec(["cat", "/tmp/ca-cert.pem"], capture_output=True)
        return p.stdout.decode()

    @property
    def ca_key(self) -> str:
        """
        Get the CA key of the etcd cluster.

        Returns:
            str: The CA key in PEM format.
        """
        p = self.instances[0].exec(["cat", "/tmp/ca-key.pem"], capture_output=True)
        return p.stdout.decode()

    @property
    def cert(self) -> str:
        """
        Get the client certificate of the etcd cluster.

        Returns:
            str: The certificate in PEM format.
        """
        p = self.instances[0].exec(["cat", "/tmp/client-cert.pem"], capture_output=True)
        return p.stdout.decode()

    @property
    def key(self) -> str:
        """
        Get the client key of the etcd cluster.

        Returns:
            str: The key in PEM format.
        """
        p = self.instances[0].exec(["cat", "/tmp/client-key.pem"], capture_output=True)
        return p.stdout.decode()

    @property
    def client_urls(self) -> List[str]:
        """
        Get the client URLs of the etcd cluster.

        Returns:
            List[str]: List of client URLs.
        """
        return [
            f"https://{get_default_ip(instance)}:2379" for instance in self.instances
        ]
