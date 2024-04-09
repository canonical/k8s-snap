#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
import shlex
from string import Template
import subprocess
from functools import partial
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, Tuple, Union

from tenacity import (
    RetryCallState,
    retry,
    retry_if_exception_type,
    stop_after_attempt,
    stop_never,
    wait_fixed,
)
from test_util import config, harness

LOG = logging.getLogger(__name__)


def run(command: list, **kwargs) -> subprocess.CompletedProcess:
    """Log and run command."""
    kwargs.setdefault("check", True)

    LOG.debug("Execute command %s (kwargs=%s)", shlex.join(command), kwargs)
    return subprocess.run(command, **kwargs)


def stubbornly(
    retries: Optional[int] = None,
    delay_s: Optional[Union[float, int]] = None,
    exceptions: Optional[tuple] = None,
    **retry_kds,
):
    """
    Retry a command for a while, using tenacity

    By default, retry immediately and forever until no exceptions occur.

    Some commands need to execute until they pass some condition
    > stubbornly(*retry_args).until(*some_condition).exec(*some_command)

    Some commands need to execute until they complete
    > stubbornly(*retry_args).exec(*some_command)

    : param    retries              int: convenience param to use stop=retry.stop_after_attempt(<int>)
    : param    delay_s        float|int: convenience param to use wait=retry.wait_fixed(delay_s)
    : param exceptions Tuple[Exception]: convenience param to use retry=retry.retry_if_exception_type(exceptions)
    : param retry_kds           Mapping: direct interface to all tenacity arguments for retrying
    """

    def _before_sleep(retry_state: RetryCallState):
        attempt = retry_state.attempt_number
        tries = f"/{retries}" if retries is not None else ""
        LOG.info(
            f"Attempt {attempt}{tries} failed. Error: {retry_state.outcome.exception()}"
        )
        LOG.info(f"Retrying in {delay_s} seconds...")

    _waits = wait_fixed(delay_s) if delay_s is not None else wait_fixed(0)
    _stops = stop_after_attempt(retries) if retries is not None else stop_never
    _exceptions = exceptions or (Exception,)  # default to retry on all exceptions

    _retry_args = dict(
        wait=_waits,
        stop=_stops,
        retry=retry_if_exception_type(_exceptions),
        before_sleep=_before_sleep,
    )
    # Permit any tenacity retry overrides from these ^defaults
    _retry_args.update(retry_kds)

    class Retriable:
        def __init__(self) -> None:
            self._condition = None
            self._run = partial(run, capture_output=True)

        @retry(**_retry_args)
        def exec(
            self,
            command_args: List[str],
            **command_kwds,
        ):
            """
            Execute a command against a harness or locally with subprocess to be retried.

            :param  List[str]        command_args: The command to be executed, as a str or list of str
            :param Mapping[str,str]      command_kwds: Additional keyword arguments to be passed to exec
            """

            try:
                resp = self._run(command_args, **command_kwds)
            except subprocess.CalledProcessError as e:
                LOG.warning(f"  rc={e.returncode}")
                LOG.warning(f"  stdout={e.stdout.decode()}")
                LOG.warning(f"  stderr={e.stderr.decode()}")
                raise
            if self._condition:
                assert self._condition(resp), "Failed to meet condition"
            return resp

        def on(self, instance: harness.Instance) -> "Retriable":
            """
            Target the command at some instance.

            :param instance Instance: Instance on a test harness.
            """
            self._run = partial(instance.exec, capture_output=True)
            return self

        def until(
            self, condition: Callable[[subprocess.CompletedProcess], bool] = None
        ) -> "Retriable":
            """
            Test the output of the executed command against an expected response

            :param Callable condition: a callable which returns a truth about the command output
            """
            self._condition = condition
            return self

    return Retriable()


# Installs and setups the k8s snap on the given instance and connects the interfaces.
def setup_k8s_snap(instance: harness.Instance, snap_path: Path):
    LOG.info("Install k8s snap")
    instance.send_file(config.SNAP, snap_path)
    instance.exec(["snap", "install", snap_path, "--classic", "--dangerous"])

    LOG.info("Ensure k8s interfaces and network requirements")
    instance.exec(["/snap/k8s/current/k8s/hack/init.sh"], stdout=subprocess.DEVNULL)


# Setups an etcd on the given instance.
# Returns the cluster members as dict.
# TODO: Make ports configurable if required.
def setup_etcd(
    instance: harness.Instance,
    cluster_members: Dict[str, str],
    etcd_url: str,
    etcd_version: str,
    join_existing: bool = False,
    ca_cert: Optional[str] = None,
    ca_key: Optional[str] = None,
) -> Dict[str, str]:
    """
    Set up etcd on the given instance.

    Args:
        instance (Instance): The instance on which to set up etcd.
        cluster_members (Dict[str, str]): Dictionary containing existing cluster members as {"name", "peer_url"}.
        etcd_url (str): The URL of the etcd service.
        etcd_version (str): The version of etcd to be set up.
        join_existing (bool): Whether the cluster is already bootstrapped or not. Defaults to false.

    Returns:
        Dict[str, str]: Updated cluster members dictionary.
    """
    LOG.info("Setup etcd")
    ip = get_default_ip(instance)
    peer_url = f"https://{ip}:2380"
    members = dict.copy(cluster_members)
    members[instance.id] = peer_url
    substitutes = {
        "NAME": instance.id,
        "IP": ip,
        "CLIENT_URL": f"https://{ip}:2379",
        "PEER_URL": peer_url,
        "PEER_URLS": ",".join(members.values()),
        "CLUSTER": ",".join([f"{key}={value}" for key, value in members.items()]),
        "CLUSTER_STATE": "existing" if join_existing else "new",
    }

    with open(config.ETCD_DIR / "etcd-tls.conf", "r") as etcd_template:
        src = Template(etcd_template.read())
        instance.exec(
            ["dd", f"of=/tmp/etcd-tls.conf"],
            input=str.encode(src.substitute(substitutes)),
        )

    # Only create CA on the first node.
    if join_existing:
        
        instance.exec(
            ["dd", f"of=/tmp/ca-cert.pem"],
            input=str.encode(ca_cert),
        )
        instance.exec(
            ["dd", f"of=/tmp/ca-key.pem"],
            input=str.encode(ca_key),
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

    instance.exec(
        [
            "openssl",
            "req",
            "-nodes",
            "-newkey",
            "rsa:4096",
            "-keyout",
            "/tmp/etcd-key.pem",
            "-out",
            "/tmp/etcd-cert.csr",
            "-config",
            "/tmp/etcd-tls.conf",
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
            "/tmp/etcd-cert.csr",
            "-CA",
            "/tmp/ca-cert.pem",
            "-CAkey",
            "/tmp/ca-key.pem",
            "-out",
            "/tmp/etcd-cert.pem",
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
            ["dd", f"of=/etc/systemd/system/etcd-s1.service"],
            input=str.encode(src.substitute(substitutes)),
        )

    instance.exec(
        [
            "curl",
            "-L",
            f"{etcd_url}/{etcd_version}/etcd-{etcd_version}-linux-amd64.tar.gz",
            "-o",
            f"/tmp/etcd-{etcd_version}-linux-amd64.tar.gz",
        ]
    )
    instance.exec(["mkdir", "-p", "/tmp/test-etcd"])
    instance.exec(
        [
            "tar",
            "xzvf",
            f"/tmp/etcd-{etcd_version}-linux-amd64.tar.gz",
            "-C",
            "/tmp/test-etcd",
            "--strip-components=1",
        ],
    )
    instance.exec(["systemctl", "daemon-reload"])
    instance.exec(["systemctl", "enable", "etcd-s1.service"])
    instance.exec(["systemctl", "start", "etcd-s1.service"])

    return members


# Validates that the K8s node is in Ready state.
def wait_until_k8s_ready(
    control_node: harness.Instance, instances: List[harness.Instance]
):
    for instance in instances:
        host = hostname(instance)
        result = (
            stubbornly(retries=15, delay_s=5)
            .on(control_node)
            .until(lambda p: " Ready" in p.stdout.decode())
            .exec(["k8s", "kubectl", "get", "node", host, "--no-headers"])
        )
    LOG.info("Kubelet registered successfully!")
    LOG.info("%s", result.stdout.decode())


def wait_for_dns(instance: harness.Instance):
    LOG.info("Waiting for CoreDNS pod to show up...")
    stubbornly(retries=15, delay_s=5).on(instance).until(
        lambda p: "coredns" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"])
    LOG.info("CoreDNS pod showed up.")

    stubbornly(retries=3, delay_s=1).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-n",
            "kube-system",
            "-l",
            "app.kubernetes.io/name=coredns",
            "--timeout",
            "180s",
        ]
    )


def wait_for_network(instance: harness.Instance):
    LOG.info("Waiting for cilium pods to show up...")
    stubbornly(retries=15, delay_s=5).on(instance).until(
        lambda p: "cilium" in p.stdout.decode()
    ).exec(
        ["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"],
    )
    LOG.info("Cilium pods showed up.")

    stubbornly(retries=3, delay_s=1).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-n",
            "kube-system",
            "-l",
            "io.cilium/app=operator",
            "--timeout",
            "180s",
        ],
    )

    stubbornly(retries=3, delay_s=1).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-n",
            "kube-system",
            "-l",
            "k8s-app=cilium",
            "--timeout",
            "180s",
        ]
    )


def hostname(instance: harness.Instance) -> str:
    """Return the hostname for a given instance."""
    resp = instance.exec(["hostname"], capture_output=True)
    return resp.stdout.decode().strip()


def get_local_node_status(instance: harness.Instance) -> str:
    resp = instance.exec(["k8s", "local-node-status"], capture_output=True)
    return resp.stdout.decode().strip()


def ready_nodes(control_node: harness.Instance) -> List[Any]:
    """Get a list of the ready nodes.

    Args:
        control_node: instance on which to execute check

    Returns:
        list of nodes
    """
    result = control_node.exec(
        "k8s kubectl get nodes -o json".split(" "), capture_output=True
    )
    assert result.returncode == 0, "Failed to get nodes with kubectl"
    node_list = json.loads(result.stdout.decode())
    assert node_list["kind"] == "List", "Should have found a list of nodes"
    nodes = [
        node
        for node in node_list["items"]
        if all(
            condition["status"] == "False"
            for condition in node["status"]["conditions"]
            if condition["type"] != "Ready"
        )
    ]
    return nodes


# Create a token to join a node to an existing cluster
def get_join_token(
    initial_node: harness.Instance, joining_cplane_node: harness.Instance, *args: str
) -> str:
    out = initial_node.exec(
        [
            "k8s",
            "get-join-token",
            joining_cplane_node.id,
            "--output-format",
            "json",
            *args,
        ],
        capture_output=True,
    )
    result = json.loads(out.stdout.decode())
    return result["join-token"]


# Join an existing cluster.
def join_cluster(instance: harness.Instance, join_token: str):
    instance.exec(["k8s", "join-cluster", join_token])


def get_default_cidr(instance: harness.Instance, instance_default_ip: str):
    # ----
    # 1:  lo    inet 127.0.0.1/8 scope host lo .....
    # 28: eth0  inet 10.42.254.197/24 metric 100 brd 10.42.254.255 scope global dynamic eth0 ....
    # ----
    # Fetching the cidr for the default interface by matching with instance ip from the output
    p = instance.exec(["ip", "-o", "-f", "inet", "addr", "show"], capture_output=True)
    out = p.stdout.decode().split(" ")
    return [i for i in out if instance_default_ip in i][0]


def get_default_ip(instance: harness.Instance):
    # ---
    # default via 10.42.254.1 dev eth0 proto dhcp src 10.42.254.197 metric 100
    # ---
    # Fetching the default IP address from the output, e.g. 10.42.254.197
    p = instance.exec(
        ["ip", "-o", "-4", "route", "show", "to", "default"], capture_output=True
    )
    return p.stdout.decode().split(" ")[8]
