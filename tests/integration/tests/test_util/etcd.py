import json
import logging
from string import Template
from typing import List
from test_util import config
from test_util.harness import Harness, Instance
from test_util.util import setup_etcd, get_default_ip

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
        self._peers = {}

        for _ in range(initial_node_count):
            self.add_node()

    def add_node(self):
        """
        Add a new node to the etcd cluster.
        If this is the first cluster node, the required certificates will be generated.
        """
        if len(self.instances) == 0:
            self._peers = self.setup_etcd_cluster(
                self._peers,
                etcd_url=self.etcd_url,
                etcd_version=self.etcd_version,
                join_existing=True,
                ca_cert=self.ca_cert,
                ca_key=self.ca_key,
            )
        else:
            self._peers = setup_etcd(
                instance,
                self._peers,
                etcd_url=self.etcd_url,
                etcd_version=self.etcd_version,
                join_existing=False,
            )

    def _setup_etcd_cluster(self, instance):
        LOG.info("Setup etcd")
        ip = get_default_ip(instance)
        join_existing = len(self.instances) > 0

        initial_cluster=f"{instance.id}=https://{ip}:2380"
        if join_existing:
            # add the member information to the cluster
            result = self.instances[0].exec(
                [
                    "/tmp/test-etcd/etcdctl",
                    "member",
                    "add",
                    instance.id,
                    f"https://{ip}:2380",
                    "--cert-file",
                    "/tmp/etcd-cert.pem",
                    "--key-file",
                    "/tmp/etcd-key.pem",
                    "--ca-file",
                    "/tmp/ca-cert.pem",
                    "-o",
                    "json",
                ],
                capture_output=True
            )
            print(result.stdout.decode())
            output = json.decode(result.stdout.decode())
            print(output)
            initial_cluster = output["ETCD_INITIAL_CLUSTER"]

        substitutes = {
            "NAME": instance.id,
            "IP": ip,
            "CLIENT_URL": f"https://{ip}:2379",
            "PEER_URL": f"https://{ip}:2380",
            "CLUSTER": initial_cluster,
            "CLUSTER_STATE": "existing" if join_existing else "new",
        }

        with open(config.ETCD_DIR / "etcd-tls.conf", "r") as etcd_template:
            src = Template(etcd_template.read())
            instance.exec(
                ["dd", f"of=/tmp/etcd-tls.conf"],
                input=str.encode(src.substitute(substitutes)),
            )

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
        p = self.instances[0].exec(["cat", "/tmp/etcd-cert.pem"], capture_output=True)
        return p.stdout.decode()

    @property
    def key(self) -> str:
        """
        Get the client key of the etcd cluster.

        Returns:
            str: The key in PEM format.
        """
        p = self.instances[0].exec(["cat", "/tmp/etcd-key.pem"], capture_output=True)
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
