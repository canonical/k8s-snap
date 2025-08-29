#
# Copyright 2025 Canonical, Ltd.
#
import ipaddress
import json
import logging
import os
import re
import shlex
import subprocess
import time
import urllib.request
from datetime import datetime
from functools import partial
from pathlib import Path
from typing import Any, Callable, Dict, List, Mapping, Optional, Union

import pytest
import yaml
from tenacity import (
    RetryCallState,
    Retrying,
    retry,
    retry_if_exception_type,
    stop_after_attempt,
    stop_never,
    wait_fixed,
)
from test_util import config, harness

LOG = logging.getLogger(__name__)
RISKS = ["stable", "candidate", "beta", "edge"]
TRACK_RE = re.compile(r"^v?(\d+)\.(\d+)(.\d+)?(\S*)$")
MAIN_BRANCH = "main"


def run(command: list, **kwargs) -> subprocess.CompletedProcess:
    """Log and run command."""
    kwargs.setdefault("check", True)

    sensitive_command = kwargs.pop("sensitive_command", False)
    sensitive_kwargs = kwargs.pop("sensitive_kwargs", sensitive_command)

    logged_command = shlex.join(command) if not sensitive_command else "<sanitized>"
    logged_kwargs = kwargs if not sensitive_kwargs else "<sanitized>"

    LOG.debug("Execute command %s (kwargs=%s)", logged_command, logged_kwargs)
    return subprocess.run(command, **kwargs)


class Retriable:
    def __init__(self, retry_kwargs: Optional[Mapping[str, Any]]) -> None:
        self._condition = None
        self._run = partial(run, capture_output=True)
        if not retry_kwargs:
            retry_kwargs = {}
        self._retry_kwargs = retry_kwargs

    def exec(
        self,
        command_args: List[str],
        **command_kwds,
    ):
        return retry(**self._retry_kwargs)(self._exec)(command_args, **command_kwds)

    def _exec(
        self,
        command_args: List[str],
        **command_kwds,
    ):
        """
        Execute a command against a harness or locally with subprocess to be retried.

        :param List[str]        command_args: The command to be executed, as a str or list of str
        :param Mapping[str,str] command_kwds: Additional keyword arguments to be passed to exec
        """

        try:
            resp = self._run(command_args, **command_kwds)
        except subprocess.CalledProcessError as e:
            stdout = e.stdout or ""
            stderr = e.stderr or ""
            if not command_kwds.get("text"):
                # The output will be in bytes if text is not set / is False.
                stdout = e.stdout.decode() if e.stdout else ""
                stderr = e.stderr.decode() if e.stderr else ""

            LOG.warning(f"  rc={e.returncode}")
            LOG.warning(f"  stdout={stdout}")
            LOG.warning(f"  stderr={stderr}")
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
        self, condition: Callable[[subprocess.CompletedProcess], bool] | None = None
    ) -> "Retriable":
        """
        Test the output of the executed command against an expected response

        :param Callable condition: a callable which returns a truth about the command output
        """
        self._condition = condition
        return self


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
        tries = f"/{retries}" if retries else ""
        errstr = ""
        if retry_state.outcome:
            errstr = f" Error: {retry_state.outcome.exception()}"
        LOG.info(f"Attempt {attempt}{tries} failed.{errstr}")
        LOG.info(f"Retrying in {delay_s} seconds...")

    waits = wait_fixed(delay_s) if delay_s else wait_fixed(0)
    stops = stop_after_attempt(retries) if retries else stop_never
    exceptions = exceptions or (Exception,)  # default to retry on all exceptions

    retry_args = dict(
        wait=waits,
        stop=stops,
        retry=retry_if_exception_type(exceptions),
        before_sleep=_before_sleep,
    )
    # Permit any tenacity retry overrides from these ^defaults
    retry_args.update(retry_kds)

    return Retriable(retry_args)


def _as_int(value: Optional[str]) -> Optional[int]:
    """Convert a string to an integer."""
    if value is None:
        return value
    try:
        return int(value)
    except (TypeError, ValueError):
        return None


def download_preloaded_snaps():
    if not config.PRELOAD_SNAPS:
        LOG.info("Snap preloading disabled, skipping...")
        return

    LOG.info(f"Downloading snaps for preloading: {config.PRELOADED_SNAPS}")
    for snap in config.PRELOADED_SNAPS:
        run(
            [
                "snap",
                "download",
                snap,
                f"--basename={snap}",
                "--target-directory=/tmp",
            ]
        )


def preload_snaps(instance: harness.Instance):
    if not config.PRELOAD_SNAPS:
        LOG.info("Snap preloading disabled.")
        return

    preload_dir, remote_dir = Path("/tmp"), Path("/tmp")
    for preloaded_snap in config.PRELOADED_SNAPS:
        ack = preload_dir / f"{preloaded_snap}.assert"
        snap = preload_dir / f"{preloaded_snap}.snap"

        LOG.info("Acknowledge snap file %s.", preloaded_snap)
        remote = remote_dir / ack.name
        instance.send_file(source=ack.as_posix(), destination=remote.as_posix())

        LOG.info("Running snap ack for %s", remote.as_posix())
        stubbornly(retries=3, delay_s=2).on(instance).exec(
            ["snap", "ack", remote.as_posix()]
        )

        LOG.info("Wait for snap changes to finish...")
        stubbornly(retries=20, delay_s=5).on(instance).until(
            lambda p: "Doing" not in p.stdout.decode()
        ).exec(["snap", "changes"])

        LOG.info("Install snap file %s.", preloaded_snap)
        remote = remote_dir / snap.name
        instance.send_file(source=snap.as_posix(), destination=remote.as_posix())
        stubbornly(retries=3, delay_s=5).on(instance).exec(
            ["snap", "install", remote.as_posix()]
        )


def setup_core_dumps(instance: harness.Instance):
    core_pattern = os.path.join(config.CORE_DUMP_DIR, config.CORE_DUMP_PATTERN)
    LOG.info("Configuring core dumps. Pattern: %s", core_pattern)
    instance.exec(["echo", core_pattern, ">", "/proc/sys/kernel/core_pattern"])
    instance.exec(["echo", "1", ">", "/proc/sys/fs/suid_dumpable"])
    instance.exec(["snap", "set", "system", "system.coredump.enable=true"])


def setup_k8s_snap(
    instance: harness.Instance,
    snap: Optional[str] = None,
    connect_interfaces=True,
    tmp_path: Optional[Path] = Path("/home/ubuntu"),
):
    """Installs and sets up the snap on the given instance and connects the interfaces.

    Args:
        instance:   instance on which to install the snap
        snap: choice of track, channel, revision, or file path
            a snap track to install
            a snap channel to install
            a snap revision to install
            a path to the snap to install
        tmp_path:   path to store the snap on the instance (optional, defaults to /home/ubuntu)
    """
    cmd = ["snap", "install", "--classic"]
    which_snap = snap or config.SNAP

    if not which_snap:
        pytest.fail(
            "Cannot install without either a channel, revision, or path to the snap "
            + f"argument {snap=} and {config.SNAP=}"
        )

    if isinstance(which_snap, str) and which_snap.startswith("/"):
        LOG.info("Install k8s snap by path")
        snap_path = (tmp_path / "k8s.snap").as_posix()
        instance.send_file(which_snap, snap_path)
        cmd += ["--dangerous", snap_path]
    elif _as_int(which_snap):
        LOG.info("Install k8s snap by revision")
        cmd += [config.SNAP_NAME, "--revision", which_snap]
    elif "/" in which_snap or which_snap in RISKS:
        LOG.info("Install k8s snap by specific channel: %s", which_snap)
        cmd += [config.SNAP_NAME, "--channel", which_snap]
    elif channel := tracks_least_risk(which_snap, instance.arch):
        LOG.info("Install k8s snap by least risky channel: %s", channel)
        cmd += [config.SNAP_NAME, "--channel", channel]

    instance.exec(cmd)
    if connect_interfaces:
        LOG.info("Ensure k8s interfaces and network requirements")
        instance.exec(["/snap/k8s/current/k8s/hack/init.sh"], stdout=subprocess.DEVNULL)


def remove_k8s_snap(instance: harness.Instance):
    LOG.info("Uninstall k8s...")
    stubbornly(retries=20, delay_s=5).on(instance).exec(
        ["snap", "remove", config.SNAP_NAME, "--purge"]
    )

    LOG.info("Waiting for shims to go away...")
    stubbornly(retries=20, delay_s=5).on(instance).until(
        lambda p: all(
            x not in p.stdout.decode()
            for x in ["containerd-shim", "cilium", "coredns", "/pause"]
        )
    ).exec(["ps", "-fea"])

    LOG.info("Waiting for kubelet and containerd mounts to go away...")
    stubbornly(retries=20, delay_s=5).on(instance).until(
        lambda p: all(
            x not in p.stdout.decode()
            for x in ["/var/lib/kubelet/pods", "/run/containerd/io.containerd"]
        )
    ).exec(["mount"])

    # NOTE(neoaggelos): Temporarily disable this as it fails on strict.
    # For details, `snap changes` then `snap change $remove_k8s_snap_change`.
    # Example output follows:
    #
    # 2024-02-23T14:10:42Z ERROR ignoring failure in hook "remove":
    # -----
    # ...
    # ip netns delete cni-UUID1
    # Cannot remove namespace file "/run/netns/cni-UUID1": Device or resource busy
    # ip netns delete cni-UUID2
    # Cannot remove namespace file "/run/netns/cni-UUID2": Device or resource busy
    # ip netns delete cni-UUID3
    # Cannot remove namespace file "/run/netns/cni-UUID3": Device or resource busy

    # LOG.info("Waiting for CNI network namespaces to go away...")
    # stubbornly(retries=5, delay_s=5).on(instance).until(
    #     lambda p: "cni-" not in p.stdout.decode()
    # ).exec(["ip", "netns", "list"])


def wait_until_k8s_ready(
    control_node: harness.Instance,
    instances: List[harness.Instance],
    retries: int = config.DEFAULT_WAIT_RETRIES,
    delay_s: int = config.DEFAULT_WAIT_DELAY_S,
    node_names: Mapping[str, str] = {},
):
    """
    Validates that the K8s node is in Ready state.

    By default, the hostname of the instances is used as the node name.
    If the instance name is different from the hostname, the instance name should be passed to the
    node_names dictionary, e.g. {"instance_id": "node_name"}.
    """
    for instance in instances:
        node_name = node_names.get(instance.id)
        if not node_name:
            node_name = hostname(instance)

        for attempt in Retrying(
            stop=stop_after_attempt(retries), wait=wait_fixed(delay_s)
        ):
            with attempt:
                assert is_node_ready(control_node, node_name)
                check_snap_services_ready(instance)

    LOG.info("Successfully checked Kubelet registered on all harness instances.")
    result = control_node.exec(["k8s", "kubectl", "get", "node"], capture_output=True)
    LOG.info("%s", result.stdout.decode().strip())


def is_node_ready(
    control_node: harness.Instance,
    node_name: str = "",
    node_dict: Optional[dict] = None,
) -> bool:
    if not (node_name or node_dict):
        raise ValueError("No node name or dict specified.")

    try:
        if not node_dict:
            out = control_node.exec(
                [
                    "k8s",
                    "kubectl",
                    "get",
                    "node",
                    node_name,
                    "-o",
                    "json",
                    "--no-headers",
                ],
                capture_output=True,
            )
            node_dict = json.loads(out.stdout.decode())
        if not node_name:
            node_name = node_dict["metadata"]["name"]

        for condition in node_dict["status"]["conditions"]:
            # TODO: consider having a map that explicitly defines the state
            # of each condition. Another option would be to rely solely on the
            # "Ready" condition.
            if condition["type"] == "Ready":
                exp_status = "True"
            else:
                exp_status = "False"
            if condition["status"] != exp_status:
                LOG.info(
                    f"Node not ready yet: {node_name}, "
                    f"condition {condition['type']}={condition['status']}"
                )
                return False

    except Exception as ex:
        LOG.info(f"Node not ready yet: {node_name}, failed to retrieve node info: {ex}")
        return False

    LOG.info(f"Node ready: {node_name}")
    return True


def wait_for_dns(instance: harness.Instance):
    LOG.info("Waiting for DNS to be ready")
    instance.exec(["k8s", "x-wait-for", "dns", "--timeout", "20m"])


def wait_for_network(instance: harness.Instance):
    LOG.info("Waiting for network to be ready")
    instance.exec(["k8s", "x-wait-for", "network", "--timeout", "20m"])


def wait_for_load_balancer(instance: harness.Instance):
    """Wait for the load balancer to be ready."""
    LOG.info("Waiting for load balancer to be ready")
    stubbornly(retries=3, delay_s=5).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=available",
            "-n",
            "metallb-system",
            "deployment.apps/metallb-controller",
            "--timeout=20m",
        ]
    )


def hostname(instance: harness.Instance) -> str:
    """Return the hostname for a given instance."""
    resp = instance.exec(["hostname"], capture_output=True)
    return resp.stdout.decode().strip()


def get_local_node_status(instance: harness.Instance) -> str:
    resp = instance.exec(["k8s", "local-node-status"], capture_output=True)
    return resp.stdout.decode().strip()


def get_datastore_type(control_node: harness.Instance) -> str:
    """Get the type of datastore used by the Kubernetes cluster.

    Should only be used with control plane nodes.
    """
    resp = control_node.exec(
        ["k8s", "status", "--output-format", "json"], capture_output=True
    )
    status = json.loads(resp.stdout.decode().strip())
    return status["datastore"]["type"]


def get_nodes(control_node: harness.Instance) -> List[Any]:
    """Get a list of existing nodes.

    Args:
        control_node: instance on which to execute check

    Returns:
        list of nodes
    """
    result = control_node.exec(
        ["k8s", "kubectl", "get", "nodes", "-o", "json"], capture_output=True
    )
    assert result.returncode == 0, "Failed to get nodes with kubectl"
    node_list = json.loads(result.stdout.decode())
    assert node_list["kind"] == "List", "Should have found a list of nodes"
    return [node for node in node_list["items"]]


def ready_nodes(control_node: harness.Instance) -> List[Any]:
    """Get a list of the ready nodes.

    Args:
        control_node: instance on which to execute check

    Returns:
        list of nodes
    """
    return [
        node
        for node in get_nodes(control_node)
        if is_node_ready(control_node, node_dict=node)
    ]


# Bootstrap the instance
def bootstrap(
    instance: harness.Instance,
    datastore_type: str,
    bootstrap_config: Optional[Dict[str, Any] | str] = None,
    extra_args: Optional[List[str]] = None,
    **kwargs,
):
    if bootstrap_config:
        if isinstance(bootstrap_config, str):
            # If bootstrap_config is a string, assume it's a YAML string
            bootstrap_config = yaml.safe_load(bootstrap_config)
    else:
        # Use bootstrap-default.yaml as the base config
        default_config_path = config.MANIFESTS_DIR / "bootstrap-default.yaml"
        bootstrap_config = yaml.safe_load(default_config_path.read_text())

    if not extra_args:
        extra_args = []

    # Add/update datastore-type and convert to YAML
    bootstrap_config["datastore-type"] = datastore_type
    modified_config = yaml.dump(bootstrap_config, default_flow_style=False)

    return instance.exec(
        ["k8s", "bootstrap", "--file", "-", *extra_args],
        input=str.encode(modified_config),
        **kwargs,
    )


# Create a token to join a node to an existing cluster
def get_join_token(
    initial_node: harness.Instance, joining_node: harness.Instance, *args: str
) -> str:
    out = (
        stubbornly(retries=5, delay_s=3)
        .on(initial_node)
        .until(lambda p: len(p.stdout.decode().strip()) > 0)
        .exec(
            ["k8s", "get-join-token", joining_node.id, *args],
            capture_output=True,
        )
    )

    return out.stdout.decode().strip()


# Join an existing cluster.
def join_cluster(
    instance: harness.Instance, join_token: str, cfg: Optional[str] = None
):
    if cfg:
        instance.exec(
            ["k8s", "join-cluster", join_token, "--file", "-"], input=str.encode(cfg)
        )
    else:
        instance.exec(["k8s", "join-cluster", join_token])


def is_ipv6(ip: str) -> bool:
    addr = ipaddress.ip_address(ip)
    return isinstance(addr, ipaddress.IPv6Address)


def get_default_cidr(instance: harness.Instance, instance_default_ip: str):
    # ----
    # 1:  lo    inet 127.0.0.1/8 scope host lo .....
    # 28: eth0  inet 10.42.254.197/24 metric 100 brd 10.42.254.255 scope global dynamic eth0 ....
    # ----
    # Fetching the cidr for the default interface by matching with instance ip from the output
    addr_family = "-6" if is_ipv6(instance_default_ip) else "-4"
    p = instance.exec(["ip", "-o", addr_family, "addr", "show"], capture_output=True)
    out = p.stdout.decode().split(" ")
    return [i for i in out if instance_default_ip in i][0]


def get_default_ip(instance: harness.Instance, ipv6=False):
    # ---
    # default via 10.42.254.1 dev eth0 proto dhcp src 10.42.254.197 metric 100
    # ---
    # Fetching the default IP address from the output, e.g. 10.42.254.197
    if ipv6:
        p = instance.exec(
            ["ip", "-json", "-6", "addr", "show", "scope", "global"],
            capture_output=True,
        )
        addr_json = json.loads(p.stdout.decode())
        if not addr_json or not addr_json[0].get("addr_info"):
            raise ValueError(
                "No IPv6 address found in the output of 'ip -json -6 addr show scope global'"
            )
        return addr_json[0]["addr_info"][0]["local"]
    else:
        p = instance.exec(
            ["ip", "-o", "-4", "route", "show", "to", "default"], capture_output=True
        )
        return p.stdout.decode().split(" ")[8]


def get_global_unicast_ipv6(instance: harness.Instance, interface="eth0") -> str | None:
    # ---
    # 2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP group default qlen 1000
    #     link/ether 00:16:3e:0f:4d:1e brd ff:ff:ff:ff:ff:ff
    #     inet
    #     inet6 fe80::216:3eff:fe0f:4d1e/64 scope link
    # ---
    # Fetching the global unicast address for the specified interface, e.g. fe80::216:3eff:fe0f:4d1e
    result = instance.exec(
        ["ip", "-6", "addr", "show", "dev", interface, "scope", "global"],
        capture_output=True,
        text=True,
    )
    output = result.stdout
    ipv6_regex = re.compile(r"inet6\s+([a-f0-9:]+)\/[0-9]*\s+scope global")
    match = ipv6_regex.search(output)
    if match:
        return match.group(1)
    return None


# Checks if a datastring is a valid RFC3339 date.
def is_valid_rfc3339(date_str):
    try:
        # Attempt to parse the string according to the RFC3339 format
        datetime.strptime(date_str, "%Y-%m-%dT%H:%M:%S%z")
        return True
    except ValueError:
        return False


def tracks_least_risk(track: str, arch: str) -> str:
    """Determine the snap channel with the least risk in the provided track.

    Args:
        track: the track to determine the least risk channel for
        arch: the architecture to narrow the revision

    Returns:
        the channel associated with the least risk
    """
    LOG.debug("Determining least risk channel for track: %s on %s", track, arch)
    if track == "latest":
        return f"latest/edge/{config.FLAVOR or 'classic'}"

    INFO_URL = f"https://api.snapcraft.io/v2/snaps/info/{config.SNAP_NAME}"
    HEADERS = {
        "Snap-Device-Series": "16",
        "User-Agent": "Mozilla/5.0",
    }

    req = urllib.request.Request(INFO_URL, headers=HEADERS)
    with urllib.request.urlopen(req) as response:
        snap_info = json.loads(response.read().decode())

    risks = [
        channel["channel"]["risk"]
        for channel in snap_info["channel-map"]
        if channel["channel"]["track"] == track
        and channel["channel"]["architecture"] == arch
    ]
    if not risks:
        raise ValueError(f"No risks found for track: {track}")
    risk_level = {"stable": 0, "candidate": 1, "beta": 2, "edge": 3}
    channel = f"{track}/{min(risks, key=lambda r: risk_level[r])}"
    LOG.info("Least risk channel from track %s is %s", track, channel)
    return channel


def major_minor(version: str) -> Optional[tuple]:
    """Determine the major and minor version of a Kubernetes version string.

    Args:
        version: the version string to determine the major and minor version for

    Returns:
        a tuple containing the major and minor version or None if the version string is invalid
    """
    if match := TRACK_RE.match(version):
        maj, min, _, _ = match.groups()
        return int(maj), int(min)
    return None


def _get_flavor() -> str:
    """Determine the flavor of the snap."""
    return {"": "classic", "strict": ""}.get(config.FLAVOR, config.FLAVOR)


def _major_minor_from_stable_upstream(maj: Optional[int] = None) -> Optional[tuple]:
    """Determine the major and minor version of the latest stable upstream release.

    Args:
        maj: the major version to use for the URL

    Returns:
        a tuple containing the major and minor version or None if the version string is invalid
    """
    addr = "https://dl.k8s.io/release/stable{dash_maj}.txt".format(
        dash_maj=f"-{maj}" if maj else ""
    )
    LOG.info("Getting upstream version from %s", addr)
    with urllib.request.urlopen(addr) as r:
        stable = r.read().decode().strip()
        return major_minor(stable)


def _previous_track_from_branch(branch: str) -> Optional[str]:
    """Determine the previous track from the branch.

    Args:
        branch: the branch to determine the previous track for

    Returns:
        the previous track or None if fails to determine
    """
    if branch == MAIN_BRANCH:
        # NOTE(Hue): `latest/stable` is not populated at the moment.
        # When it is, we should return `latest` instead.
        LOG.info("Getting current version from upstream k8s")
        # For the main branch, the previous track is the latest release-branch, e.g.
        # `1.32/stable` for `main` branch which matches the current upstream version.
        maj_min = _major_minor_from_stable_upstream()
        if not maj_min:
            LOG.info("Failed to determine upstream version")
            return None

    elif branch.startswith("release-"):
        LOG.info("Getting current version from branch %s", branch)
        maj_min = major_minor(branch.lstrip("release-"))
        # Get the previous version from the branch, e.g. for branch `release-1.32` we want `1.31`
        if maj_min:
            _maj, _min = maj_min[0], maj_min[1]
            if _min == 0:
                maj_min = _major_minor_from_stable_upstream(_maj - 1)
            else:
                maj_min = (_maj, _min - 1)
    else:
        LOG.info(
            "Branch is neither `main` nor `release-X.Y`. Can't determine previous track."
        )
        return None

    flavor = _get_flavor()
    return f"{maj_min[0]}.{maj_min[1]}" + (flavor and f"-{flavor}") if maj_min else None


def previous_track(snap_version: str) -> str:
    """Determine the snap track preceding the provided version.

    Args:
        snap_version: the snap version to determine the previous track for

    Returns:
        the previous track
    """
    LOG.debug("Determining previous track for %s", snap_version)

    if not snap_version:
        assumed = "latest"
        LOG.info(
            "Cannot determine previous track for undefined snap -- assume %s",
            snap_version,
            assumed,
        )
        return assumed

    if snap_version.startswith("/") or _as_int(snap_version) is not None:
        branch = config.GH_BASE_REF or config.GH_REF or MAIN_BRANCH
        prev = _previous_track_from_branch(branch)
        if prev:
            LOG.info(
                "Previous track for %s from branch %s is %s", snap_version, branch, prev
            )
            return f"{prev}/edge"
        else:
            LOG.info(
                "Previous track for %s from branch %s is not found -- assume latest",
                snap_version,
                branch,
            )
            return f"latest/edge/{_get_flavor()}"

    if maj_min := major_minor(snap_version):
        maj, min = maj_min
        if min == 0:
            maj_min = _major_minor_from_stable_upstream(maj - 1)
        else:
            maj_min = (maj, min - 1)
    elif snap_version.startswith("latest") or "/" not in snap_version:
        maj_min = _major_minor_from_stable_upstream()

    if not maj_min:
        raise ValueError(
            "Failed to determine previous snap version track for "
            f"current version: {snap_version}"
        )

    flavor_track = _get_flavor()
    track = f"{maj_min[0]}.{maj_min[1]}" + (flavor_track and f"-{flavor_track}")
    LOG.info("Previous track for %s is from track: %s", snap_version, track)
    return track


def find_suitable_cidr(parent_cidr: str, excluded_ips: List[str]):
    """Find a suitable CIDR for LoadBalancer services"""
    net = ipaddress.ip_network(parent_cidr, False)
    ipv6 = isinstance(net, ipaddress.IPv6Network)
    if ipv6:
        ip_range = 126
    else:
        ip_range = 30

    # Starting from the first IP address from the parent cidr,
    # we search for a /30 cidr block(4 total ips, 2 available)
    # that doesn't contain the excluded ips to avoid collisions
    # /30 because this is the smallest CIDR cilium hands out IPs from.
    # For ipv6, we use a /126 block that contains 4 total ips.
    for i in range(4, 255, 4):
        lb_net = ipaddress.ip_network(f"{str(net[0]+i)}/{ip_range}", False)

        contains_excluded = False
        for excluded in excluded_ips:
            if ipaddress.ip_address(excluded) in lb_net:
                contains_excluded = True
                break

        if contains_excluded:
            continue

        return str(lb_net)
    raise RuntimeError("Could not find a suitable CIDR for LoadBalancer services")


def check_file_paths_exist(
    instance: harness.Instance, paths: List[str]
) -> Mapping[str, bool]:
    """Returns whether the given path(s) exist within the given harness instance
    by checking the output of a single `ls` command containing all of them.

    It is recommended to always use absolute paths, as the cwd relative to
    which the `ls` will get executed depends on the harness instance.
    """
    process = instance.exec(["ls", *paths], capture_output=True, text=True, check=False)
    return {
        p: f"cannot access '{p}': No such file or directory" not in process.stderr
        for p in paths
    }


def get_os_version_id_for_instance(instance: harness.Instance) -> str:
    """Returns the version of the OS on the given harness Instance
    by reading the `VERSION_ID` from `/etc/os-release`.
    """
    proc = instance.exec(["cat", "/etc/os-release"], capture_output=True)

    release = None
    var = "VERSION_ID"
    for line in proc.stdout.split(b"\n"):
        line = line.decode()
        if line.startswith(var):
            release = line.lstrip(f"{var}=")
            break

    if not release:
        raise ValueError(
            f"Failed to parse OS release var '{var}' from OS release "
            f"info: {proc.stdout}"
        )

    return release


def wait_for_daemonset(
    instance: harness.Instance,
    name: str,
    namespace: str = "default",
    retry_times: int = 5,
    retry_delay_s: int = 60,
    expected_pods_ready: int = 1,
):
    """Waits for the daemonset with the given name to have at least
    `expected_pods_ready` pods ready."""
    proc = None
    for i in range(retry_times):
        # NOTE: we can't reliably use `rollout status` on Daemonsets unless
        # they have `RollingUpdate` strategy, so we must go by the number of
        # pods which are Ready.
        proc = instance.exec(
            [
                "k8s",
                "kubectl",
                "-n",
                namespace,
                "get",
                "daemonset",
                name,
                "-o",
                "jsonpath={.status.numberReady}",
            ],
            check=True,
            capture_output=True,
        )
        if int(proc.stdout.decode()) >= expected_pods_ready:
            LOG.info(
                f"Successfully waited for daemonset '{name}' after "
                f"{(i+1)*retry_delay_s} seconds"
            )
            return

        LOG.info(
            f"Waiting {retry_delay_s} seconds for daemonset '{name}'.\n"
            f"code: {proc.returncode}\nstdout: {proc.stdout}\nstderr: {proc.stderr}"
        )
        time.sleep(retry_delay_s)

    raise AssertionError(
        f"Daemonset '{name}' failed to have at least one pod ready after "
        f"{retry_times} x {retry_delay_s} seconds."
    )


# sonobuoy_tar_gz returns the download URL of sonobuoy.
def sonobuoy_tar_gz(architecture: str) -> str:
    version = config.SONOBUOY_VERSION
    return f"https://github.com/vmware-tanzu/sonobuoy/releases/download/{version}/sonobuoy_{version[1:]}_linux_{architecture}.tar.gz"  # noqa


def check_snap_services_ready(
    instance: harness.Instance,
    node_type: Optional[str] = None,
    skip_services: Optional[List[str]] = None,
    datastore_type: Optional[str] = None,
):
    """Check that the snap services are active on the given harness instance.

    The expected services differ between control-plane and worker nodes.
    The function will determine the node type by checking the local node status.

    Args:
        instance: the harness instance to check the snap services on
        node_type: the node type to check the services for. If not provided, the
            function will determine the node type by checking the local node status.
            This is not always possible (e.g. if a node was already removed from the cluster).
            So, the user can provide the node type explicitly.
        skip_services: a list of services to ignore when checking for service readiness.
    """
    skip_services = skip_services or []

    expected_worker_services = {
        "containerd",
        "k8sd",
        "kubelet",
        "kube-proxy",
        "k8s-apiserver-proxy",
    }
    expected_control_plane_services = {
        "containerd",
        "k8sd",
        "kubelet",
        "kube-proxy",
        "kube-apiserver",
        "kube-controller-manager",
        "kube-scheduler",
    }
    if node_type:
        assert node_type in ("control-plane", "worker"), "Invalid node type provided"
        expected_active_services = (
            expected_control_plane_services
            if node_type == "control-plane"
            else expected_worker_services
        )
    else:
        node_type = (
            "control-plane"
            if "control-plane" in get_local_node_status(instance)
            else "worker"
        )

    if node_type == "control-plane":
        # If the node is a control-plane node, we need to check the datastore service.
        # The datastore type can be provided explicitly or determined from the instance.
        if datastore_type:
            assert datastore_type in (
                "etcd",
                "k8s-dqlite",
                "external",
            ), "Invalid datastore type provided"
        else:
            datastore_type = get_datastore_type(instance)

        if datastore_type != "external":
            expected_control_plane_services.add(datastore_type)

    expected_active_services = (
        expected_control_plane_services
        if node_type == "control-plane"
        else expected_worker_services
    )

    if skip_services:
        expected_active_services = [
            s for s in expected_active_services if s not in skip_services
        ]

    result = instance.exec(["snap", "services", "k8s"], capture_output=True, text=True)
    services_output = result.stdout.split("\n")[1:-1]  # Skip the header line

    service_status = {}
    for line in services_output:
        parts = line.split()
        if len(parts) >= 3:  # Ensure there are enough columns
            service_name = parts[0].replace("k8s.", "", 1)
            service_status[service_name] = parts[2]  # "active" or "inactive"

    for service in expected_active_services:
        assert (
            service in service_status
        ), f"Service {service} is missing from 'snap services' output"
        assert (
            service_status[service] == "active"
        ), f"Service {service} should be active, but it is {service_status[service]}"

    for service, status in service_status.items():
        if service in skip_services:
            continue
        if service not in expected_active_services:
            assert (
                status == "inactive"
            ), f"Unexpected service {service} is {status} but should be inactive"


def is_fips_enabled(instance: harness.Instance):
    """
    Returns True if the provided instance is running with FIPS enabled, False otherwise.
    """
    fips_path = "/proc/sys/crypto/fips_enabled"
    try:
        result = instance.exec(["cat", fips_path], capture_output=True, text=True)
        return result.stdout.strip() == "1"
    except subprocess.CalledProcessError:
        return False


def status_output_matches(
    p: subprocess.CompletedProcess, status_pattern: List[str]
) -> bool:
    """
    Check if the output of the `k8s status` command matches the expected pattern.
    """
    result_lines = p.stdout.decode().strip().split("\n")
    if len(result_lines) != len(status_pattern):
        LOG.info(
            "wrong number of results lines, expected %s, got %s",
            len(status_pattern),
            len(result_lines),
        )
        return False

    for i, l in enumerate(result_lines):
        line, pattern = l, status_pattern[i]
        if not re.search(pattern, line):
            LOG.info(
                "could not match `%s` with `%s`",
                line.strip(),
                pattern,
            )
            return False

    return True
